package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/schema"
)

type Ticket struct {
	ZDNum     int    `schema:"zdnum"`
	UserID    int    `schema:"userid"`
	IssueType string `schema:"issuetype"`
	Initials  string `schema:"initials"`
	Comments  string `schema:"comments"`
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

func createTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Fatalln(err)
		}

		t := Ticket{}
		decoder := schema.NewDecoder()

		err = decoder.Decode(&t, r.PostForm)
		if err != nil {
			log.Fatalln(err)
		}

		err = tpl.ExecuteTemplate(w, "submitted.gohtml", nil)
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
