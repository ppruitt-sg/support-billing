package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/go-sql-driver/mysql"

	"./admin"
	"./database"
	"./ticket"
)

func TestNewHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/new", nil)
	rr := httptest.NewRecorder()

	require.Nil(t, err)
	require.NotNil(t, req)
	require.NotNil(t, rr)

	http.HandlerFunc(ticket.New).ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
}

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	require.Nil(t, err)
	require.NotNil(t, req)
	require.NotNil(t, rr)

	http.HandlerFunc(ticket.Home).ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
}

func TestViewTicketHandler(t *testing.T) {
	var err error
	var tests = []struct {
		n        string // ticket number
		expected int
	}{
		{"1", 200},                       // Existing ticket
		{"", 404},                        // Blank ticket
		{"25325235235235235235253", 500}, // Ticket too long
		{"0", 404},
		{"65000", 404},
	}

	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	r := mux.NewRouter()

	r.HandleFunc("/view/{number}", ticket.Retrieve())

	rr := httptest.NewServer(r)

	require.Nil(t, err)
	require.NotNil(t, database.DBCon)
	require.NotNil(t, r)
	require.NotNil(t, rr)

	for _, test := range tests {
		url := rr.URL + "/view/" + test.n
		resp, err := http.Get(url)
		require.Nil(t, err)

		assert.Equal(t, resp.StatusCode, test.expected)
	}
}

func TestSolveTicketHandler(t *testing.T) {
	var tests = []struct {
		n        string // ticket number
		expected int
	}{
		{"1", 200},                       // Existing ticket
		{"", 404},                        // Blank ticket
		{"25325235235235235235253", 500}, // Ticket too long
		{"0", 404},
		{"65000", 404},
	}
	database.DBCon, _ = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	r := mux.NewRouter()
	r.HandleFunc("/solve/{number}", ticket.Solve)
	r.HandleFunc("/view/{number}", ticket.Retrieve())

	rr := httptest.NewServer(r)

	require.NotNil(t, database.DBCon)
	require.NotNil(t, r)
	require.NotNil(t, rr)

	for _, test := range tests {
		url := rr.URL + "/view/" + test.n
		resp, err := http.Post(url, "", nil)
		require.Nil(t, err)

		assert.Equal(t, resp.StatusCode, test.expected)
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

func TestAdmin(t *testing.T) {
	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/admin", nil)

	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()

	http.HandlerFunc(admin.Admin).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

/*************************************************************
// Benchmarks
*************************************************************/
func BenchmarkNewHandler(b *testing.B) {
	req, err := http.NewRequest("GET", "/new", nil)

	if err != nil {
		b.Errorf("An error occurred. %v", err)
	}

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		http.HandlerFunc(ticket.New).ServeHTTP(rr, req)
	}
}

func BenchmarkHomeHandler(b *testing.B) {
	req, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		b.Errorf("An error occurred. %v", err)
	}

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		http.HandlerFunc(ticket.New).ServeHTTP(rr, req)
	}
}

func BenchmarkViewTicketHandler(b *testing.B) {
	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	if err != nil {
		b.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/solve/{number}", ticket.Retrieve())

	rr := httptest.NewServer(r)
	defer rr.Close()
	for i := 0; i < b.N; i++ {
		resp, _ := http.Post(rr.URL+"/solve/1", "", nil)
		resp.Body.Close()
	}

}

func BenchmarkSolveTicketHandler(b *testing.B) {
	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	if err != nil {
		b.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/view/{number}", func(http.ResponseWriter, *http.Request) {})
	r.HandleFunc("/solve/{number}", ticket.Solve)

	rr := httptest.NewServer(r)
	defer rr.Close()
	for i := 0; i < b.N; i++ {
		resp, _ := http.Post(rr.URL+"/solve/1", "", nil)
		resp.Body.Close()
	}

}

func BenchmarkViewOpenHandler(b *testing.B) {
	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	if err != nil {
		b.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/view/open", nil)

	if err != nil {
		b.Errorf("An error occurred. %v", err)
	}
	rr := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		http.HandlerFunc(ticket.Retrieve10(ticket.StatusOpen)).ServeHTTP(rr, req)
	}

}

func BenchmarkViewSolvedHandler(b *testing.B) {
	var err error
	database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	if err != nil {
		b.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/view/solved", nil)

	if err != nil {
		b.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		http.HandlerFunc(ticket.Retrieve10(ticket.StatusSolved)).ServeHTTP(rr, req)
	}

}
