package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"./comment"
	"./view"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/schema"
)

type IssueType int

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

func templateHandler(name string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := view.Render(w, name+".gohtml", nil)
		if err != nil {
			log.Fatalln(err)
		}

	}
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
	fmt.Println(t.Issue)

	return t, nil
}
func addTicketToDB(t *Ticket) error {
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

func getTicketFromDB(num int64) (Ticket, error) {
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

func createTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Decode form post to Ticket struct
		t, err := parseForm(r)
		if err != nil {
			log.Fatalln(err)
		}
		// Add to database
		err = addTicketToDB(&t)
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

func displayTicket() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ticketNumber, err := strconv.ParseInt(strings.Replace(r.URL.Path, "/view/", "", 1), 10, 64)
		if err != nil {
			log.Fatalln(err)
		}
		t, err := getTicketFromDB(ticketNumber)
		if err != nil {
			log.Fatalln(err)
		}

		err = view.Render(w, "viewticket.gohtml", t)
		if err != nil {
			log.Fatalln(err)
		}
		// If not return 404
	}

}

func main() {
	http.HandleFunc("/new/", templateHandler("new"))
	http.HandleFunc("/create", createTicket)
	http.HandleFunc("/view/", displayTicket())

	http.ListenAndServe(":8080", nil)
}
