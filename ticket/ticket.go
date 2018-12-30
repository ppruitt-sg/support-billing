package ticket

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"../comment"
	"../view"
	"github.com/gorilla/mux"
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

type TicketsPage struct {
	Tickets    []Ticket
	NextButton bool
	NextPage   int64
	PrevPage   int64
	PrevButton bool
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

func Home(w http.ResponseWriter, r *http.Request) {
	New(w, r)
}

func New(w http.ResponseWriter, r *http.Request) {
	view.Render(w, "new.gohtml", nil)
}

func getOffsetFromPage(page int64) int64 {
	if page == 1 {
		return 0
	}
	return page*10 - 10
}

func Retrieve10(status StatusType) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var page int64
		var tp TicketsPage
		var err error

		tp.Status = status

		if keys, ok := r.URL.Query()["page"]; ok {
			page, err = strconv.ParseInt(keys[0], 10, 64)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			page = 1
		}

		offset := getOffsetFromPage(page)

		tp.Tickets, err = getNext10FromDB(offset, status)
		if err != nil {
			log.Fatalln(err)
		}
		tp.NextPage = page + 1
		if len(tp.Tickets) == 10 {
			tp.NextButton = true
		}

		tp.PrevPage = page - 1
		if page > 1 {
			tp.PrevButton = true
		}
		view.Render(w, "listtickets.gohtml", tp)
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

func Search(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	http.Redirect(w, r, "../view/"+r.Form["number"][0], http.StatusMovedPermanently)
	return
}

func Retrieve() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ticketNumber, err := strconv.ParseInt(vars["number"], 10, 64)
		if err != nil {
			log.Fatalln(err)
		}

		t, err := getFromDB(ticketNumber)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				w.WriteHeader(http.StatusNotFound)
				view.Render(w, "ticketnotfound.gohtml", ticketNumber)
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

func Solve(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		vars := mux.Vars(r)
		ticketNumber, err := strconv.ParseInt(vars["number"], 10, 64)
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
