package routes

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"../database"
	. "../structs"
	"../view"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

func retrieveMCTickets(d database.Datastore) ([]Ticket, error) {
	currentMonth := time.Now()
	pacific, err := time.LoadLocation("America/Los_Angeles")
	startOfCurrentMonth := time.Date(currentMonth.Year(), currentMonth.Month(), 1, 0, 0, 0, 0, pacific)
	startOfNextMonth := startOfCurrentMonth.AddDate(0, 1, 0)

	ts, err := d.GetMCTicketsFromDB(startOfCurrentMonth.Unix(), startOfNextMonth.Unix())
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
		return t, err
	}

	decoder := schema.NewDecoder()

	err = decoder.Decode(&t, r.PostForm)
	if err != nil {
		return t, err
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

func Retrieve10(d database.Datastore, status StatusType, issues ...IssueType) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var page int64
		var tp TicketsPage
		var err error

		// Find type of ticket being viewed
		re := regexp.MustCompile(`/view/(\w*)`)
		typeViewed := re.FindStringSubmatch(r.URL.Path)[1]

		// Determine ticket page type by typeViewed
		switch typeViewed {
		case "cx":
			tp.Type = "CX"
		case "lead":
			tp.Type = "Lead"
		case "solved":
			tp.Type = "Solved"
		default:
			tp.Type = "[undefined]"
		}

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
		offset := page*10 - 10

		// Get 10 tickets from page based off offset and status
		tp.Tickets, err = d.GetNext10TicketsFromDB(offset, status, issues...)
		if err != nil {
			logError(fmt.Sprintf("Getting 10 tickets for page %d", page), err, w)
			return
		}

		// Set next page if there are more than 10 tickets, NextPage remains 0 if not
		if len(tp.Tickets) == 10 {
			tp.NextPage = page + 1
		}

		// Set previous page
		tp.PrevPage = page - 1

		view.Render(w, "listtickets.gohtml", tp)
	}
}

func Create(d database.Datastore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
		t, err = d.AddTicketToDB(t)
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
}

func Retrieve(d database.Datastore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse "number" variable from URL
		vars := mux.Vars(r)
		ticketNumber, err := strconv.ParseInt(vars["number"], 10, 64)
		if err != nil {
			logError("Parsing number from URL", err, w)
			return
		}

		// Get specific ticket number
		t, err := d.GetTicketFromDB(ticketNumber)
		if err != nil {
			switch err {
			// If the ticket doesn't exist, return 404 and display ticketnotfound.gohtml
			case sql.ErrNoRows:
				w.WriteHeader(http.StatusNotFound)
				view.Render(w, "ticketnotfound.gohtml", ticketNumber)
				return
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

func Solve(d database.Datastore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse "number" variable from URL
		vars := mux.Vars(r)
		ticketNumber, err := strconv.ParseInt(vars["number"], 10, 64)
		if err != nil {
			logError("Parsing number from URL", err, w)
			return
		}

		// Get specific ticket number
		t, err := d.GetTicketFromDB(ticketNumber)
		if err != nil {
			logError(fmt.Sprintf("Getting ticket %d from database", ticketNumber), err, w)
			return
		}

		// Solve ticket and update it to the database
		t.Status = StatusSolved
		err = d.UpdateTicketToDB(t)
		if err != nil {
			logError(fmt.Sprintf("Updating ticket %d in database", ticketNumber), err, w)
			return
		}
		// Redirect back to the ticket view
		http.Redirect(w, r, "/view/"+strconv.FormatInt(t.Number, 10), http.StatusMovedPermanently)
	}
}

func Admin(d database.Datastore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ts, err := retrieveMCTickets(d)
		if err != nil {
			logError("Error retrieving MC Tickets", err, w)
		}
		_ = ts
		view.Render(w, "admin.gohtml", ts)
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	view.Render(w, "404.gohtml", nil)
}
