package ticket

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"../comment"
	"../view"
	"github.com/gorilla/schema"
)

type IssueType int

func (i IssueType) ToString() string {
	switch i {
	case 0:
		return "Refund"
	case 1:
		return "Billing Terminated"
	case 2:
		return "DNA FP"
	case 3:
		return "Extension"
	default:
		return ""
	}
}

const (
	Refund     IssueType = 0
	Terminated IssueType = 1
	DNAFP      IssueType = 2
	Extension  IssueType = 3
)

type Ticket struct {
	Number   int64           `schema:"-"`
	ZDNum    int             `schema:"zdnum"`
	UserID   int             `schema:"userid"`
	Issue    IssueType       `schema:"issue"`
	Initials string          `schema:"initials"`
	Solved   bool            `schema:"-"`
	Comment  comment.Comment `schema:"comment"`
}

type Tickets struct {
	Tickets    []Ticket
	NextButton bool
	LastTicket int64
}

func parseForm(r *http.Request) (Ticket, error) {
	err := r.ParseForm()
	if err != nil {
		return Ticket{}, err
	}

	t := Ticket{}
	decoder := schema.NewDecoder()

	err = decoder.Decode(&t, r.PostForm)
	if err != nil {
		return Ticket{}, err
	}
	t.Comment.Timestamp = time.Now()

	return t, nil
}

func (t *Ticket) addToDB() error {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()

	query := `INSERT INTO tickets (zdticket, userid, issue, initials, solved)
		VALUES (?, ?, ?, ?, 0);`
	result, err := db.Exec(query, t.ZDNum, t.UserID, t.Issue, t.Initials)
	if err != nil {
		return err
	}
	t.Number, err = result.LastInsertId()
	if err != nil {
		return err
	}
	t.Comment.TicketNumber = t.Number

	err = t.Comment.AddToDB()
	if err != nil {
		return err
	}

	return nil
}

func getFromDB(num int64) (Ticket, error) {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return Ticket{}, err
	}

	query := `SELECT ticket_id, zdticket, userid, issue, initials, solved FROM tickets
		WHERE ticket_id=?`
	r := db.QueryRow(query, num)
	t := Ticket{}
	err = r.Scan(&t.Number, &t.ZDNum, &t.UserID, &t.Issue, &t.Initials, &t.Solved)
	if err != nil {
		return Ticket{}, err
	}

	t.Comment, err = comment.GetFromDB(num)
	if err != nil {
		return Ticket{}, err
	}

	return t, nil
}

func getNext5FromDB(lastTicket int64) ([]Ticket, error) {
	var ts []Ticket
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return ts, err
	}

	// Select rows with limit
	query := `SELECT ticket_id, zdticket, userid, issue, initials, solved 
		FROM tickets 
		WHERE solved=0 AND ticket_id>?
		LIMIT 5`
	r, err := db.Query(query, lastTicket)
	if err != nil {
		return ts, err
	}

	t := Ticket{}
	for r.Next() {
		err = r.Scan(&t.Number, &t.ZDNum, &t.UserID, &t.Issue, &t.Initials, &t.Solved)
		if err != nil {
			return ts, err
		}
		ts = append(ts, t)
	}
	if r.Err() != nil {
		return ts, err
	}
	return ts, nil
}

func findRowsFound(lastTicket int64) (int64, error) {
	var rowsFound int64

	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return rowsFound, err
	}

	query := `SELECT COUNT(*) 
		FROM tickets 
		WHERE solved=0 AND ticket_id>?`
	count := db.QueryRow(query, lastTicket)

	err = count.Scan(&rowsFound)
	if err != nil {
		return rowsFound, err
	}

	return rowsFound, nil
}

func DisplayNext5() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var lastTicket int64
		var ts Tickets
		var err error
		if keys, ok := r.URL.Query()["last_ticket"]; ok {
			lastTicket, err = strconv.ParseInt(keys[0], 10, 64)
			if err != nil {
				log.Fatalln(err)
			}
		}

		ts.Tickets, err = getNext5FromDB(lastTicket)
		if err != nil {
			log.Fatalln(err)
		}
		if len(ts.Tickets) > 0 {
			ts.LastTicket = ts.Tickets[len(ts.Tickets)-1].Number
		}

		rowsFound, err := findRowsFound(lastTicket)
		if err != nil {
			log.Fatalln(err)
		}

		if rowsFound > 5 {
			ts.NextButton = true
		}
		view.Render(w, "listtickets.gohtml", ts)
	}
}

func (t Ticket) updateToDB() error {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `UPDATE tickets
		SET solved=?
		WHERE ticket_id=?`

	_, err = db.Exec(query, t.Solved, t.Number)
	if err != nil {
		return err
	}

	return nil

}

func Create(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Decode form post to Ticket struct
		t, err := parseForm(r)
		if err != nil {
			log.Fatalln(err)
		}
		// Add to database
		err = t.addToDB()
		if err != nil {
			log.Fatalln(err)
		}
		// Display submitted text
		err = view.Render(w, "submitted.gohtml", t)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func parseIntFromURL(path string, r *http.Request) (int64, error) {
	return strconv.ParseInt(strings.Replace(r.URL.Path, path, "", 1), 10, 64)
}

func Display() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ticketNumber, err := parseIntFromURL("/view/", r)
		if err != nil {
			log.Fatalln(err)
		}
		t, err := getFromDB(ticketNumber)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				w.WriteHeader(http.StatusNotFound)
				return
			default:
				log.Fatalln(err)
			}
		}

		err = view.Render(w, "viewticket.gohtml", t)
		if err != nil {
			log.Fatalln(err)
		}
	}

}

func getAllFromDB() ([]Ticket, error) {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	var ts []Ticket
	if err != nil {
		return []Ticket{}, err
	}

	query := `SELECT ticket_id, zdticket, userid, issue, initials, solved FROM tickets`
	r, err := db.Query(query)
	if err != nil {
		return ts, err
	}
	t := Ticket{}
	for r.Next() {
		err = r.Scan(&t.Number, &t.ZDNum, &t.UserID, &t.Issue, &t.Initials, &t.Solved)
		if err != nil {
			return ts, err
		}
		ts = append(ts, t)
	}
	if r.Err() != nil {
		return ts, r.Err()
	}

	return ts, nil
}

func DisplayAll(w http.ResponseWriter, r *http.Request) {
	var ts, err = getAllFromDB()
	if err != nil {
		log.Fatalln(err)
	}
	view.Render(w, "listtickets.gohtml", ts)
}

func Solve(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		ticketNumber, err := parseIntFromURL("/solve/", r)
		if err != nil {
			log.Fatalln(err)
		}

		t, err := getFromDB(ticketNumber)
		if err != nil {
			log.Fatalln(err)
		}
		t.Solved = true
		err = t.updateToDB()
		if err != nil {
			log.Fatalln(err)
		}
		http.Redirect(w, r, "/view/"+strconv.FormatInt(t.Number, 10), http.StatusPermanentRedirect)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
