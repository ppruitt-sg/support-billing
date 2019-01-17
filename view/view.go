package view

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var tpl *template.Template
var loc *time.Location

func init() {
	var err error

	funcMap := template.FuncMap{
		"ToLower":    strings.ToLower,
		"ToDateTime": toDateTime,
		"ToDate":     toDate,
		"AddBreaks":  addBreaks,
	}

	// Parse templates in /template
	tpl, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.gohtml")
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
	// Convert to readable time
	return t.In(loc).Format("Jan _2 2006")
}

func addBreaks(s string) template.HTML {
	// Convert newline to HTML readable line break
	return template.HTML(strings.Replace(template.HTMLEscapeString(s), "\r\n", "<br>", -1))
}
