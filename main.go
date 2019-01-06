package main

import (
	"log"
	"net/http"
	"os"

	"database/sql"

	"./database"
	"./ticket"
	"./view"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	if err != nil {
		log.Fatalln(err)
	}

	r := mux.NewRouter()

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	r.StrictSlash(true)

	r.HandleFunc("/", ticket.Home).Methods("GET")
	r.HandleFunc("/new/", ticket.New).Methods("GET")
	r.HandleFunc("/create", ticket.Create).Methods("POST")
	r.HandleFunc("/view/open/", ticket.Retrieve10(ticket.StatusOpen)).Methods("GET")
	r.HandleFunc("/view/solved/", ticket.Retrieve10(ticket.StatusSolved)).Methods("GET")
	r.HandleFunc("/view/{number:[0-9]+}", ticket.Retrieve()).Methods("GET")
	r.HandleFunc("/solve/{number:[0-9]+}", ticket.Solve).Methods("POST")

	r.NotFoundHandler = http.HandlerFunc(notFound)

	http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, r))

}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	view.Render(w, "404.gohtml", nil)
}
