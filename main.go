package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func templateHandler(name string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tpl.ExecuteTemplate(w, name+".gohtml", nil)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func createTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		for key, value := range r.Form {
			fmt.Printf("%s = %s\n", key, value[0])
		}
	}
}

func main() {
	http.HandleFunc("/new", templateHandler("new"))
	http.HandleFunc("/create", createTicket)

	http.ListenAndServe(":8080", nil)

}
