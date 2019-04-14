package main

import (
	"log"
	"net/http"
	"os"

	"./admin"
	"./database"
	"./ticket"
	"./view"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	var db database.DB
	var err error
	err = db.NewDB(os.Getenv("RDS_USERNAME") + ":" + os.Getenv("RDS_PASSWORD") + "@tcp(" + os.Getenv("RDS_HOSTNAME") + ":" + os.Getenv("RDS_PORT") + ")/" + os.Getenv("RDS_DB_NAME"))
	if err != nil {
		log.Fatalln(err)
	}

	r := mux.NewRouter()

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	r.StrictSlash(true)

	r.HandleFunc("/", ticket.Home).Methods("GET")
	r.HandleFunc("/new/", ticket.New).Methods("GET")
	r.HandleFunc("/create", ticket.Create(&db)).Methods("POST")
	r.HandleFunc("/view/open/", ticket.Retrieve10(&db, database.StatusOpen)).Methods("GET")
	r.HandleFunc("/view/solved/", ticket.Retrieve10(&db, database.StatusSolved)).Methods("GET")
	r.HandleFunc("/view/{number:[0-9]+}", ticket.Retrieve(&db)).Methods("GET")
	r.HandleFunc("/solve/{number:[0-9]+}", ticket.Solve(&db)).Methods("POST")
	r.HandleFunc("/admin", admin.Admin(&db)).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(notFound)

	http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, r))

}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	view.Render(w, "404.gohtml", nil)
}
