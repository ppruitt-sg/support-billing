package view

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

var tpl *template.Template

func init() {
	pwd, _ := os.Getwd()
	var err error

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	tpl, err = template.New("").Funcs(funcMap).ParseGlob(pwd + "/templates/*.gohtml")
	if err != nil {
		log.Panic(err)
	}
	/* tpl = template.Must(template.ParseGlob(pwd + "/templates/*.gohtml"))
	tpl.Funcs(funcMap) */

}

func Render(w http.ResponseWriter, name string, data interface{}) error {

	err := tpl.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}
	return nil
}
