package main

import (
	"log"
	"net/http"

	"./ticket"
	"./view"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	err := ticket.Test()
	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/new/", view.TemplateHandler("new"))
	http.HandleFunc("/create", ticket.Create)
	http.HandleFunc("/view/", ticket.Display())
	http.HandleFunc("/solve/", ticket.Solve)

	http.ListenAndServe(":8080", nil)
}
