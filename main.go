package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/schema"
)

type Ticket struct {
	Number    int64   `schema:"-"`
	ZDNum     int     `schema:"zdnum"`
	UserID    int     `schema:"userid"`
	IssueType string  `schema:"issuetype"`
	Initials  string  `schema:"initials"`
	Solved    bool    `schema:"-"`
	Comment   Comment `schema:"comment"`
}

type Comment struct {
	Timestamp    time.Time
	Text         string `schema:"text"`
	TicketNumber int64
}

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func templateHandler(name string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tpl.ExecuteTemplate(w, name+".gohtml", nil)
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

	return t, nil
}

func addCommentToDB(c Comment) error {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	query := `INSERT INTO comments (timestamp, text, ticket_id)
		VALUES (?, ?, ?)`
	_, err = db.Exec(query, c.Timestamp.Unix(), c.Text, c.TicketNumber)
	if err != nil {
		return err
	}
	return nil
}

func addTicketToDB(t *Ticket) error {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()

	query := `INSERT INTO tickets (zdticket, userid, issuetype, initials, solved)
		VALUES (?, ?, ?, ?, 0);`
	result, err := db.Exec(query, t.ZDNum, t.UserID, t.IssueType, t.Initials)
	if err != nil {
		return err
	}

	t.Number, err = result.LastInsertId()
	if err != nil {
		return err
	}
	t.Comment.TicketNumber = t.Number

	err = addCommentToDB(t.Comment)
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

	query := `SELECT ticket_id, zdticket, userid, issuetype, initials, solved FROM tickets
		WHERE ticket_id=?`
	r := db.QueryRow(query, num)
	t := Ticket{}
	err = r.Scan(&t.Number, &t.ZDNum, &t.UserID, &t.IssueType, &t.Initials, &t.Solved)
	if err != nil {
		return Ticket{}, err
	}

	t.Comment, err = getCommentFromDB(num)
	if err != nil {
		return Ticket{}, err
	}

	return t, nil
}

func getCommentFromDB(num int64) (Comment, error) {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return Comment{}, err
	}

	query := `SELECT timestamp, text, ticket_id FROM comments
		WHERE ticket_id=?`

	r := db.QueryRow(query, num)
	c := Comment{}
	var ts int64
	err = r.Scan(&ts, &c.Text, &c.TicketNumber)
	c.Timestamp = time.Unix(ts, 0)
	if err != nil {
		return Comment{}, err
	}
	fmt.Println(c.Text)

	return c, nil
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
		err = tpl.ExecuteTemplate(w, "submitted.gohtml", t)
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
			fmt.Println(err.Error())
			log.Fatalln(err)
		}
		t, err := getTicketFromDB(ticketNumber)
		if err != nil {
			log.Fatalln(err)
		}

		err = tpl.ExecuteTemplate(w, "viewticket.gohtml", t)
		if err != nil {
			log.Fatalln(err)
		}
		// If not return 404
	}

}

func main() {
	http.HandleFunc("/new", templateHandler("new"))
	http.HandleFunc("/create", createTicket)
	http.HandleFunc("/view/", displayTicket())

	http.ListenAndServe(":8080", nil)
}
