package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
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
	Comment   Comment `schema:"comment"`
}

type Comment struct {
	TicketNumber int64
	Timestamp    int64
	Text         string `schema:"text"`
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
	t.Comment.Timestamp = time.Now().UTC().Unix()

	return t, nil
}

func addCommentToDB(c Comment) error {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	query := `INSERT INTO comments (timestamp, text, ticket_id)
		VALUES (?, ?, ?)`
	_, err = db.Exec(query, c.Timestamp, c.Text, c.TicketNumber)
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
	}
}

func main() {
	http.HandleFunc("/new", templateHandler("new"))
	http.HandleFunc("/create", createTicket)

	http.ListenAndServe(":8080", nil)
}
