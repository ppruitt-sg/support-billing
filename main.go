package main

import (
	"log"
	"net/http"

	"./ticket"
	"./view"
	_ "github.com/go-sql-driver/mysql"
)

func templateHandler(name string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := view.Render(w, name+".gohtml", nil)
		if err != nil {
			log.Fatalln(err)
		}

	}
}

func main() {
	http.HandleFunc("/new/", templateHandler("new"))
	http.HandleFunc("/create", ticket.Create)
	http.HandleFunc("/view/", ticket.Display())

	http.ListenAndServe(":8080", nil)
}
