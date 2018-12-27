package main

import (
	"log"
	"net/http"
	"os"

	"database/sql"

	"./database"
	"./ticket"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	http.HandleFunc("/", ticket.LogHandler(ticket.Home))
	http.HandleFunc("/new/", ticket.LogHandler(ticket.New))
	http.HandleFunc("/create", ticket.LogHandler(ticket.Create))
	http.HandleFunc("/view/open/", ticket.LogHandler(ticket.RetrieveNext10(ticket.StatusOpen)))
	http.HandleFunc("/view/solved/", ticket.LogHandler(ticket.RetrieveNext10(ticket.StatusSolved)))
	http.HandleFunc("/view/", ticket.LogHandler(ticket.Retrieve()))
	http.HandleFunc("/solve/", ticket.LogHandler(ticket.Solve))
	http.HandleFunc("/search/", ticket.LogHandler(ticket.Search))

	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	if err != nil {
		log.Fatalln(err)
	}

	http.ListenAndServe(":8080", nil)

}
