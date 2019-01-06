package view

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var tpl *template.Template
var loc *time.Location

func init() {
	pwd, _ := os.Getwd()
	var err error

	funcMap := template.FuncMap{
		"ToLower":    strings.ToLower,
		"ToDateTime": toDateTime,
		"ToDate":     toDate,
	}

	// Parse templates in /template
	tpl, err = template.New("").Funcs(funcMap).ParseGlob(pwd + "/templates/*.gohtml")
	if err != nil {
		log.Panic(err)
	}

	// Set timezone to "America/Denver"
	loc, err = time.LoadLocation("America/Denver")
	if err != nil {
		log.Panic(err)
	}

}

// Render template with data and included funcmap
func Render(w http.ResponseWriter, name string, data interface{}) error {
	err := tpl.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}
	return nil
}

func toDateTime(t time.Time) string {
	// Convert to readable time and date
	return t.In(loc).Format("January 02 2006 3:04pm MST")
}

func toDate(t time.Time) string {
	// Conver to readable time
	return t.In(loc).Format("Jan _2 2006")
}
