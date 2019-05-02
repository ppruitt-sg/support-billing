package main

import (
	"log"
	"net/http"
	"os"

	"./database"
	"./routes"
	. "./structs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	var db database.DB
	var err error

	cxIssues := []IssueType{Refund, Terminated, DNAFP, Extension}
	leadIssues := []IssueType{Discount, Downgrade, UndoDowngrade}
	allIssues := []IssueType{Refund, Terminated, DNAFP, Extension, Discount, Downgrade, UndoDowngrade}

	err = db.NewDB(os.Getenv("RDS_USERNAME") + ":" + os.Getenv("RDS_PASSWORD") + "@tcp(" + os.Getenv("RDS_HOSTNAME") + ":" + os.Getenv("RDS_PORT") + ")/" + os.Getenv("RDS_DB_NAME"))
	if err != nil {
		log.Fatalln(err)
	}

	r := mux.NewRouter()

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	r.StrictSlash(true)

	r.HandleFunc("/", routes.Home).Methods("GET")
	r.HandleFunc("/new/", routes.New).Methods("GET")
	r.HandleFunc("/create", routes.Create(&db)).Methods("POST")
	r.HandleFunc("/view/cx/", routes.Retrieve10(&db, StatusOpen, cxIssues...)).Methods("GET")
	r.HandleFunc("/view/lead/", routes.Retrieve10(&db, StatusOpen, leadIssues...)).Methods("GET")
	r.HandleFunc("/view/solved/", routes.Retrieve10(&db, StatusSolved, allIssues...)).Methods("GET")
	r.HandleFunc("/view/{number:[0-9]+}", routes.Retrieve(&db)).Methods("GET")
	r.HandleFunc("/solve/{number:[0-9]+}", routes.Solve(&db)).Methods("POST")
	r.HandleFunc("/admin", routes.Admin(&db)).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(routes.NotFound)

	http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, r))

}
