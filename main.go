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

	r.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	r.HandleFunc("/", ticket.Home)
	r.HandleFunc("/new/", ticket.New)
	r.HandleFunc("/create", ticket.Create)
	r.HandleFunc("/view/open/", ticket.RetrieveNext10(ticket.StatusOpen))
	r.HandleFunc("/view/solved/", ticket.RetrieveNext10(ticket.StatusSolved))
	r.HandleFunc("/view/{number:[0-9]+}", ticket.Retrieve())
	r.HandleFunc("/solve/", ticket.Solve)
	r.HandleFunc("/search/", ticket.Search)

	r.NotFoundHandler = http.HandlerFunc(notFound)

	http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, r))

}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	view.Render(w, "404.gohtml", nil)
}
