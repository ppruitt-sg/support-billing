package view

import (
	"html/template"
	"net/http"
	"os"
)

var tpl *template.Template

func init() {
	pwd, _ := os.Getwd()
	tpl = template.Must(template.ParseGlob(pwd + "/templates/*.gohtml"))
}

func Render(w http.ResponseWriter, name string, data interface{}) error {
	err := tpl.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}
	return nil
}
