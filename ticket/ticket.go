package ticket

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"../comment"
	"../view"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

// Ticket structure
type Ticket struct {
	Number    int64      `schema:"-"`
	ZDTicket  int        `schema:"zdticket"`
	UserID    int        `schema:"userid"`
	Issue     IssueType  `schema:"issue"`
	Initials  string     `schema:"initials"`
	Status    StatusType `schema:"-"`
	Submitted time.Time
	Comment   comment.Comment `schema:"comment"`
}

// Tickets page structure for paginating
type TicketsPage struct {
	Tickets    []Ticket
	NextButton bool
	NextPage   int64
	PrevPage   int64
	PrevButton bool
	Status     StatusType
}

func RetrieveMCTickets() ([]Ticket, error) {
	currentMonth := time.Now()
	pacific, err := time.LoadLocation("America/Los_Angeles")
	startOfCurrentMonth := time.Date(currentMonth.Year(), currentMonth.Month(), 1, 0, 0, 0, 0, pacific)
	startOfNextMonth := startOfCurrentMonth.AddDate(0, 1, 0)

	ts, err := getMCTicketsFromDB(startOfCurrentMonth.Unix(), startOfNextMonth.Unix())
	if err != nil {
		return ts, err
	}

	return ts, nil
}

func logError(action string, err error, w http.ResponseWriter) {
	// Print action and error message
	log.Printf("Error - %s - %v", action, err)
	w.WriteHeader(http.StatusInternalServerError)
}

func parseNewForm(r *http.Request) (t Ticket, err error) {
	// Parse the new ticket form in /templates/new.gohtml
	err = r.ParseForm()
	if err != nil {
		return Ticket{}, err
	}

	decoder := schema.NewDecoder()

	err = decoder.Decode(&t, r.PostForm)
	if err != nil {
		return Ticket{}, err
	}

	// Check max length of comment
	if len(t.Comment.Text) > 255 {
		return t, errors.New("Comment exceeds max width (255 characters)")
	}

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

		// Checks for page parameter
		if keys, ok := r.URL.Query()["page"]; ok {
			page, err = strconv.ParseInt(keys[0], 10, 64)
			if err != nil {
				logError("Parsing page parameter", err, w)
				return
			}
		} else {
			// If it doesn't exist, set page to 1
			page = 1
		}

		// Get offset value based off page number
		offset := getOffsetFromPage(page)

		// Get 10 tickets from page based off offset and status
		tp.Tickets, err = getNext10FromDB(offset, status)
		if err != nil {
			logError(fmt.Sprintf("Getting 10 tickets for page %d", page), err, w)
			return
		}

		// Include Next Button if there are 10 tickets
		tp.NextPage = page + 1
		if len(tp.Tickets) == 10 {
			tp.NextButton = true
		}

		// Included Previous Button if this isn't page 1
		tp.PrevPage = page - 1
		if page > 1 {
			tp.PrevButton = true
		}
		view.Render(w, "listtickets.gohtml", tp)
	}
}

func Create(w http.ResponseWriter, r *http.Request) {
	// Decode form post to Ticket struct
	t, err := parseNewForm(r)
	if err != nil {
		logError("Parsing new ticket form", err, w)
		return
	}

	// Set up timestamp on ticket and comment
	t.Submitted = time.Now()
	t.Comment.Timestamp = time.Now()

	// Add to database
	err = t.addToDB()
	if err != nil {
		logError("Adding ticket to database", err, w)
		return
	}
	// Display submitted text
	tpl := "submitted.gohtml"
	err = view.Render(w, tpl, t)
	if err != nil {
		logError(fmt.Sprintf("Displaying %s template", tpl), err, w)
		return
	}
}

func Retrieve() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse "number" variable from URL
		vars := mux.Vars(r)
		ticketNumber, err := strconv.ParseInt(vars["number"], 10, 64)
		if err != nil {
			logError("Parsing number from URL", err, w)
			return
		}

		// Get specific ticket number
		t, err := getFromDB(ticketNumber)
		if err != nil {
			switch err {
			// If the ticket doesn't exist, return 404 and display ticketnotfound.gohtml
			case sql.ErrNoRows:
				w.WriteHeader(http.StatusNotFound)
				view.Render(w, "ticketnotfound.gohtml", ticketNumber)

			default:
				logError("Getting ticket from DB", err, w)
				return
			}
		}

		// Render viewticket.gohtml
		tpl := "viewticket.gohtml"
		err = view.Render(w, tpl, t)
		if err != nil {
			logError(fmt.Sprintf("Rendering %s template", tpl), err, w)
			return
		}
	}
}

func Solve(w http.ResponseWriter, r *http.Request) {
	// Parse "number" variable from URL
	vars := mux.Vars(r)
	ticketNumber, err := strconv.ParseInt(vars["number"], 10, 64)
	if err != nil {
		logError("Parsing number from URL", err, w)
		return
	}

	// Get specific ticket number
	t, err := getFromDB(ticketNumber)
	if err != nil {
		logError(fmt.Sprintf("Getting ticket %d from database", ticketNumber), err, w)
		return
	}

	// Solve ticket and update it to the database
	t.Status = StatusSolved
	err = t.updateToDB()
	if err != nil {
		logError(fmt.Sprintf("Updating ticket %d in database", ticketNumber), err, w)
		return
	}
	// Redirect back to the ticket view
	http.Redirect(w, r, "/view/"+strconv.FormatInt(t.Number, 10), http.StatusMovedPermanently)
}
