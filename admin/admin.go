package admin

import (
	"log"
	"net/http"

	"../ticket"
	"../view"
)

func logError(action string, err error, w http.ResponseWriter) {
	// Print action and error message
	log.Printf("Error - %s - %v", action, err)
	w.WriteHeader(http.StatusInternalServerError)
}

func Admin(w http.ResponseWriter, r *http.Request) {
	ts, err := ticket.RetrieveMCTickets()
	if err != nil {
		logError("Error retrieving MC Tickets", err, w)
	}
	_ = ts
	view.Render(w, "admin.gohtml", ts)

}
