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

func DisplayNext5(solved bool) func(http.ResponseWriter, *http.Request) {
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

		ts.Tickets, err = getNext5FromDB(lastTicket, solved)
		if err != nil {
			log.Fatalln(err)
		}
		if len(ts.Tickets) > 0 {
			ts.LastTicket = ts.Tickets[len(ts.Tickets)-1].Number
		}

		rowsFound, err := getRowsFound(lastTicket, solved)
		if err != nil {
			log.Fatalln(err)
		}

		if rowsFound > 5 {
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
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func parseIntFromURL(path string, r *http.Request) (int64, error) {
	return strconv.ParseInt(strings.Replace(r.URL.Path, path, "", 1), 10, 64)
}

func Display() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ticketNumber int64
		var err error
		switch r.Method {
		case "GET":
			ticketNumber, err = parseIntFromURL("/view/", r)
			if err != nil {
				log.Fatalln(err)
			}
		case "POST":
			r.ParseForm()
			http.Redirect(w, r, r.Form["number"][0], 301)
			return
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

/* func DisplayAll(w http.ResponseWriter, r *http.Request) {
	var ts, err = getAllFromDB()
	if err != nil {
		log.Fatalln(err)
	}
	view.Render(w, "listtickets.gohtml", ts)
} */

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
