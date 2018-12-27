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
)

func main() {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	http.HandleFunc("/", view.TemplateHandler("new"))
	http.HandleFunc("/new/", view.TemplateHandler("new"))
	http.HandleFunc("/create", ticket.Create)
	http.HandleFunc("/view/open/", ticket.DisplayNext10(ticket.StatusOpen))
	http.HandleFunc("/view/solved/", ticket.DisplayNext10(ticket.StatusSolved))
	http.HandleFunc("/view/", ticket.Display())
	http.HandleFunc("/solve/", ticket.Solve)
	http.HandleFunc("/search/", ticket.Search)

	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	if err != nil {
		log.Fatalln(err)
	}

	http.ListenAndServe(":8080", nil)

}
