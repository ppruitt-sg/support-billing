package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/ppruitt-sg/support-billing/database"
	. "github.com/ppruitt-sg/support-billing/structs"
	"github.com/ppruitt-sg/support-billing/view"
)

func retrieveMCTickets(d database.Datastore) ([]Ticket, error) {
	currentMonth := time.Now()
	pacific, err := time.LoadLocation("America/Los_Angeles")
	startOfCurrentMonth := time.Date(currentMonth.Year(), currentMonth.Month(), 1, 0, 0, 0, 0, pacific)
	startOfNextMonth := startOfCurrentMonth.AddDate(0, 1, 0)

	ts, err := d.GetMCTickets(startOfCurrentMonth.Unix(), startOfNextMonth.Unix())
	if err != nil {
		return ts, err
	}

	return ts, nil
}

func logError(action string, err error, w http.ResponseWriter) {
	// Print action and error message
	log.Printf("Error - %s - %v", action, err)

	// Wrote just for localhost
	//kafka.WriteError(action, err)

	w.WriteHeader(http.StatusInternalServerError)
}

func parseForm(r *http.Request) (t Ticket, err error) {
	// Parse the new form in /templates/new.gohtml
	err = r.ParseForm()
	if err != nil {
		return t, err
	}

	decoder := schema.NewDecoder()

	err = decoder.Decode(&t, r.PostForm)
	if err != nil {
		return t, err
	}

	return t, nil
}

func validateInput(t Ticket) (err error) {
	maxIntSize := 2147483647
	if t.UserID > maxIntSize {
		return fmt.Errorf("UserID is greater than %d", maxIntSize)
	}

	if t.ZDTicket > maxIntSize {
		return fmt.Errorf("ZDTicket is greater than %d", maxIntSize)
	}

	comment := strings.ReplaceAll(t.Comment.Text, "\n", `\n`)
	if len(comment) > 255 {
		return fmt.Errorf("Comment exceeds 255 characters, was %d characters", len(comment))
	}
	return nil
}

func checkURLParameter(url *url.URL, parameter string) string {
	// Checks for page parameter
	if keys, ok := url.Query()[parameter]; ok {
		return keys[0]
	}
	// If it doesn't exist, return empty string
	return ""
}

func findTicketType(url *url.URL) string {
	// Find type of ticket being viewed
	re := regexp.MustCompile(`/view/(\w*)`)
	typeViewed := re.FindStringSubmatch(url.Path)[1]

	// Determine ticket page type by typeViewed
	switch typeViewed {
	case "cx":
		return "CX"
	case "lead":
		return "Lead"
	case "solved":
		return "Solved"
	default:
		return "[undefined]"
	}
}

func parseNumberFromURL(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	return strconv.ParseInt(vars["number"], 10, 64)
}

func containsIssue(issueTypeSlice []IssueType, issue IssueType) bool {
	for _, value := range issueTypeSlice {
		if value == issue {
			return true
		}
	}
	return false
}

func Home(w http.ResponseWriter, r *http.Request) {
	New(w, r)
}

func New(w http.ResponseWriter, r *http.Request) {
	view.Render(w, "new.gohtml", nil)
}

func Retrieve10(d database.Datastore, status StatusType, issues ...IssueType) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var tp TicketsPage
		var page int64
		var solvedTicket int64
		var err error

		// Get ticket type for title of table on site
		tp.Type = findTicketType(r.URL)

		// Checks for page parameter
		parameter := checkURLParameter(r.URL, "page")

		if parameter != "" {
			page, err = strconv.ParseInt(parameter, 10, 64)
			if err != nil {
				logError(fmt.Sprintf("Converting page parameter %s", parameter), err, w)
			}
		} else {
			page = 1
		}

		// Checks for solved_ticket parameter
		parameter = checkURLParameter(r.URL, "solved_ticket")

		if parameter != "" {
			solvedTicket, err = strconv.ParseInt(parameter, 10, 64)
			if err != nil {
				logError(fmt.Sprintf("Converting page parameter %s", parameter), err, w)
			}
		} else {
			solvedTicket = 0
		}

		// Get offset value based off page number
		offset := page*10 - 10

		// Get 10 tickets from page based off offset and status
		tp.Tickets, err = d.GetNext10Tickets(offset, status, issues...)
		if err != nil {
			logError(fmt.Sprintf("Getting 10 tickets for page %d", page), err, w)
			return
		}

		// Set the previous and next page
		tp.SetPrevAndNextPage(page)
		tp.SolvedTicket = solvedTicket

		view.Render(w, "listtickets.gohtml", tp)
	}
}

func Create(d database.Datastore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode form post to Ticket struct
		t, err := parseForm(r)
		if err != nil {
			logError("Parsing new ticket form", err, w)
			return
		}

		// Validate input
		err = validateInput(t)
		if err != nil {
			logError("Validating input", err, w)
			return
		}

		// Set up timestamp on ticket and comment
		t.Submitted = time.Now()
		t.Comment.Timestamp = time.Now()

		// Add to database
		t, err = d.AddTicket(t)
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
		ticketNumber, err := parseNumberFromURL(r)
		if err != nil {
			logError("Parsing number from URL", err, w)
			return
		}

		// Get specific ticket number
		t, err := d.GetTicket(ticketNumber)
		if err != nil {
			switch err {
			// If the ticket doesn't exist, return 404 and display ticketnotfound.gohtml
			case sql.ErrNoRows:
				w.WriteHeader(http.StatusNotFound)
				view.Render(w, "ticketnotfound.gohtml", ticketNumber)
			default:
				logError("Getting ticket from DB", err, w)
			}
			return
		}
		view.Render(w, "viewticket.gohtml", t)
	}
}

func Edit(d database.Datastore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse "number" variable from URL
		ticketNumber, err := parseNumberFromURL(r)
		if err != nil {
			logError("Parsing number from URL", err, w)
			return
		}

		// Get specific ticket number
		t, err := d.GetTicket(ticketNumber)
		if err != nil {
			switch err {
			// If the ticket doesn't exist, return 404 and display ticketnotfound.gohtml
			case sql.ErrNoRows:
				w.WriteHeader(http.StatusNotFound)
				view.Render(w, "ticketnotfound.gohtml", ticketNumber)
			default:
				logError("Getting ticket from DB", err, w)
			}
			return
		}
		view.Render(w, "editticket.gohtml", t)
	}
}

func Update(d database.Datastore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse "number" variable from URL
		ticketNumber, err := parseNumberFromURL(r)
		if err != nil {
			logError("Parsing number from URL", err, w)
			return
		}

		// Get specific ticket number
		ticket, err := d.GetTicket(ticketNumber)
		if err != nil {
			logError(fmt.Sprintf("Getting ticket %d from database", ticketNumber), err, w)
			return
		}

		// Get updated ticket information and patch it into the original ticket
		updatedTicket, err := parseForm(r)
		ticket.Patch(updatedTicket)

		// Update the ticket
		err = d.UpdateTicket(ticket)
		if err != nil {
			logError(fmt.Sprintf("Updating ticket %d in database", ticketNumber), err, w)
			return
		}
		// Redirect back to the ticket view
		http.Redirect(w, r, "/view/"+strconv.FormatInt(ticket.Number, 10), http.StatusMovedPermanently)
	}
}

func Solve(d database.Datastore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var redirectURL = "/view/%s/?solved_ticket=%d"
		var ticketType string
		// Parse "number" variable from URL
		ticketNumber, err := parseNumberFromURL(r)
		if err != nil {
			logError("Parsing number from URL", err, w)
			return
		}

		// Get specific ticket number
		t, err := d.GetTicket(ticketNumber)
		if err != nil {
			logError(fmt.Sprintf("Getting ticket %d from database", ticketNumber), err, w)
			return
		}

		// Solve ticket and update it to the database
		t.Status = StatusSolved
		err = d.UpdateTicket(t)
		if err != nil {
			logError(fmt.Sprintf("Updating ticket %d in database", ticketNumber), err, w)
			return
		}

		if containsIssue(LeadIssues, t.Issue) {
			ticketType = "lead"
		} else {
			ticketType = "cx"
		}

		// Redirect back to the tickets view
		http.Redirect(w, r, fmt.Sprintf(redirectURL, ticketType, t.Number), http.StatusMovedPermanently)
	}
}

func Admin(d database.Datastore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ts, err := retrieveMCTickets(d)
		if err != nil {
			logError("Error retrieving MC Tickets", err, w)
			return
		}
		_ = ts
		view.Render(w, "admin.gohtml", ts)
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	view.Render(w, "404.gohtml", nil)
}
