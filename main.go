package main

import (
	"net/http"

	"./ticket"
	"./view"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	http.HandleFunc("/", view.TemplateHandler("new"))
	http.HandleFunc("/new/", view.TemplateHandler("new"))
	http.HandleFunc("/create", ticket.Create)
	http.HandleFunc("/view/open/", ticket.DisplayNext10(ticket.StatusOpen))
	http.HandleFunc("/view/solved/", ticket.DisplayNext10(ticket.StatusSolved))
	http.HandleFunc("/view/", ticket.Display())
	http.HandleFunc("/solve/", ticket.Solve)
	http.HandleFunc("/search/", ticket.Search)

	http.ListenAndServe(":8080", nil)
}
