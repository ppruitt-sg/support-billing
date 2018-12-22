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
	http.HandleFunc("/view/open/", ticket.DisplayNext5(false))
	http.HandleFunc("/view/solved/", ticket.DisplayNext5(true))
	http.HandleFunc("/view/", ticket.Display())
	http.HandleFunc("/solve/", ticket.Solve)

	http.ListenAndServe(":8080", nil)
}
