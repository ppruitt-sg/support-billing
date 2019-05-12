package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/ppruitt-sg/support-billing/database"
	"github.com/ppruitt-sg/support-billing/routes"
	. "github.com/ppruitt-sg/support-billing/structs"
)

type Specification struct {
	Username string `default:"ppruitt" envconfig:"RDS_USERNAME"`
	Password string `default:"password" envconfig:"RDS_PASSWORD"`
	Hostname string `default:"localhost" envconfig:"RDS_HOSTNAME"`
	Port     string `default:"3306" envconfig:"RDS_PORT"`
	DBName   string `default:"supportbilling" envconfig:"RDS_DB_NAME"`
}

func main() {
	var db database.DB
	var err error
	var s Specification

	err = envconfig.Process("RDS", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(s.Username)

	cxIssues := []IssueType{Refund, Terminated, DNAFP, Extension}
	leadIssues := []IssueType{Discount, Downgrade, UndoDowngrade}
	allIssues := []IssueType{Refund, Terminated, DNAFP, Extension, Discount, Downgrade, UndoDowngrade}

	err = db.NewDB(s.Username + ":" + s.Password + "@tcp(" + s.Hostname + ":" + s.Port + ")/" + s.DBName)
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
