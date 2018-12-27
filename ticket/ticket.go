package ticket

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"../comment"
	"../view"
	"github.com/gorilla/schema"
)

type Ticket struct {
	Number   int64           `schema:"-"`
	ZDTicket int             `schema:"zdticket"`
	UserID   int             `schema:"userid"`
	Issue    IssueType       `schema:"issue"`
	Initials string          `schema:"initials"`
	Status   StatusType      `schema:"-"`
	Comment  comment.Comment `schema:"comment"`
}

type Tickets struct {
	Tickets    []Ticket
	NextButton bool
	LastTicket int64
	Status     StatusType
}

func parseNewForm(r *http.Request) (t Ticket, err error) {
	err = r.ParseForm()
	if err != nil {
		return Ticket{}, err
	}

	decoder := schema.NewDecoder()

	err = decoder.Decode(&t, r.PostForm)
	if err != nil {
		return Ticket{}, err
	}
	t.Comment.Timestamp = time.Now()

	return t, nil
}

func parseSearchForm(r *http.Request) (t int, err error) {
	err = r.ParseForm()
	if err != nil {
		return t, err
	}

	decoder := schema.NewDecoder()

	err = decoder.Decode(&t, r.PostForm)
	if err != nil {
		return t, err
	}

	return t, err
}

func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFoundHandler(w, r)
		return
	}
	New(w, r)
}

func New(w http.ResponseWriter, r *http.Request) {
	view.Render(w, "new.gohtml", nil)
}

func DisplayNext10(status StatusType) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var lastTicket int64
		var ts Tickets
		ts.Status = status
		var err error
		if keys, ok := r.URL.Query()["last_ticket"]; ok {
			lastTicket, err = strconv.ParseInt(keys[0], 10, 64)
			if err != nil {
				log.Fatalln(err)
			}
		}

		ts.Tickets, err = getNext10FromDB(lastTicket, status)
		if err != nil {
			log.Fatalln(err)
		}
		if len(ts.Tickets) > 0 {
			ts.LastTicket = ts.Tickets[len(ts.Tickets)-1].Number
		}

		rowsFound, err := getRowsFound(lastTicket, status)
		if err != nil {
			log.Fatalln(err)
		}

		if rowsFound > 10 {
			ts.NextButton = true
		}
		view.Render(w, "listtickets.gohtml", ts)
	}
}

func Create(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Decode form post to Ticket struct
		t, err := parseNewForm(r)
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
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

func parseIntFromURL(path string, r *http.Request) (int64, error) {
	return strconv.ParseInt(strings.Replace(r.URL.Path, path, "", 1), 10, 64)
}

func Search(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	http.Redirect(w, r, "../view/"+r.Form["number"][0], http.StatusMovedPermanently)
	return
}

func Display() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ticketNumber, err := parseIntFromURL("/view/", r)
		if err != nil {
			// Treat error as 404
			notFoundHandler(w, r)
			return
		}

		t, err := getFromDB(ticketNumber)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				w.WriteHeader(http.StatusNotFound)
				view.Render(w, "ticketnotfound.gohtml", ticketNumber)
				log.Printf("%s - 404", r.URL.Path)
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

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	view.Render(w, "404.gohtml", nil)
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
		t.Status = StatusSolved
		err = t.updateToDB()
		if err != nil {
			log.Fatalln(err)
		}
		http.Redirect(w, r, "/view/"+strconv.FormatInt(t.Number, 10), http.StatusMovedPermanently)
	} else {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

func LogHandler(f func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s - %s", r.Method, r.URL.Path)
		f(w, r)
	}
}
