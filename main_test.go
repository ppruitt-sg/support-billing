package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"

	"./database"
	"./ticket"
)

func TestNewHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/new", nil)

	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()

	http.HandlerFunc(ticket.New).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()

	http.HandlerFunc(ticket.Home).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func TestViewTicketHandler(t *testing.T) {
	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))

	r := mux.NewRouter()
	r.HandleFunc("/view/{number}", ticket.Retrieve())

	rr := httptest.NewServer(r)

	url := rr.URL + "/view/1"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func TestViewOpenHandler(t *testing.T) {
	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/view/open", nil)

	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()

	http.HandlerFunc(ticket.Retrieve10(ticket.StatusOpen)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func TestViewSolvedHandler(t *testing.T) {
	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/view/solved", nil)

	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()

	http.HandlerFunc(ticket.Retrieve10(ticket.StatusSolved)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}
