package view

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

var tpl *template.Template

func init() {
	pwd, _ := os.Getwd()
	tpl = template.Must(template.ParseGlob(pwd + "/templates/*.gohtml"))
}

func TemplateHandler(name string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := Render(w, name+".gohtml", nil)
		if err != nil {
			log.Fatalln(err)
		}

	}
}

func Render(w http.ResponseWriter, name string, data interface{}) error {
	err := tpl.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}
	return nil
}
