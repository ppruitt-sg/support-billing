package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/go-sql-driver/mysql"

	"./admin"
	"./database"
	"./ticket"
)

/*******************************
// Mock DB
*******************************/
type mockDB struct {
}

/*type Datastore interface {
	// Comment
	AddCommentToDB(Comment) error
	GetCommentFromDB(int64) (Comment, error)
	// Ticket
	UpdateTicketToDB(Ticket) error
	AddTicketToDB(Ticket) (Ticket, error)
	GetTicketFromDB(int64) (Ticket, error)
	GetNext10TicketsFromDB(int64, StatusType) ([]Ticket, error)
	GetMCTicketsFromDB(int64, int64) ([]Ticket, error)
}*/
func (d mockDB) AddCommentToDB(c database.Comment) error {
	return nil
}

func (d mockDB) GetCommentFromDB(stamp int64) (database.Comment, error) {
	return database.Comment{
		Timestamp:    time.Unix(stamp, 0),
		Text:         "Hello",
		TicketNumber: 1,
	}, nil
}

func (d mockDB) UpdateTicketToDB(t database.Ticket) error {
	return nil
}

func (d mockDB) AddTicketToDB(t database.Ticket) (database.Ticket, error) {
	t.Number = 1
	return t, nil
}

func (d mockDB) GetTicketFromDB(number int64) (database.Ticket, error) {
	return database.Ticket{
		Number:    1,
		ZDTicket:  1234,
		UserID:    5678,
		Issue:     database.Refund,
		Initials:  "PP",
		Status:    database.StatusOpen,
		Submitted: time.Now(),
		Comment: database.Comment{
			Timestamp:    time.Now(),
			Text:         "Testing",
			TicketNumber: 1,
		},
	}, nil
}

func (d mockDB) GetNext10TicketsFromDB(offset int64, status database.StatusType) ([]database.Ticket, error) {
	return []database.Ticket{
		database.Ticket{
			Number:    1,
			ZDTicket:  1234,
			UserID:    5678,
			Issue:     database.Refund,
			Initials:  "PP",
			Status:    status,
			Submitted: time.Now(),
			Comment: database.Comment{
				Timestamp:    time.Now(),
				Text:         "Testing",
				TicketNumber: 1,
			},
		},
		database.Ticket{
			Number:    2,
			ZDTicket:  5678,
			UserID:    1234,
			Issue:     database.Refund,
			Initials:  "PP",
			Status:    status,
			Submitted: time.Now(),
			Comment: database.Comment{
				Timestamp:    time.Now(),
				Text:         "Testing Too",
				TicketNumber: 2,
			},
		},
	}, nil
}

func (d mockDB) GetMCTicketsFromDB(start int64, end int64) ([]database.Ticket, error) {
	return []database.Ticket{
		database.Ticket{
			Number:    1,
			ZDTicket:  1234,
			UserID:    5678,
			Issue:     database.MCContacts,
			Initials:  "PP",
			Status:    database.StatusOpen,
			Submitted: time.Now(),
			Comment: database.Comment{
				Timestamp:    time.Now(),
				Text:         "Testing",
				TicketNumber: 1,
			},
		},
		database.Ticket{
			Number:    2,
			ZDTicket:  5678,
			UserID:    1234,
			Issue:     database.MCContacts,
			Initials:  "PP",
			Status:    database.StatusOpen,
			Submitted: time.Now(),
			Comment: database.Comment{
				Timestamp:    time.Now(),
				Text:         "Testing Too",
				TicketNumber: 2,
			},
		},
	}, nil
}

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
	}

	//database.DBCon, err = sql.Open("mysql", os.Getenv("RDS_USERNAME")+":"+os.Getenv("RDS_PASSWORD")+"@tcp("+os.Getenv("RDS_HOSTNAME")+":"+os.Getenv("RDS_PORT")+")/"+os.Getenv("RDS_DB_NAME"))
	db := new(mockDB)
	r := mux.NewRouter()

	r.HandleFunc("/view/{number}", ticket.Retrieve(db))

	rr := httptest.NewServer(r)

	require.Nil(t, err)
	require.NotNil(t, db)
	require.NotNil(t, r)
	require.NotNil(t, rr)

	for _, test := range tests {
		url := rr.URL + "/view/" + test.n
		resp, err := http.Get(url)
		require.Nil(t, err)

		assert.Equal(t, test.expected, resp.StatusCode)
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
	}
	db := new(mockDB)
	r := mux.NewRouter()
	r.HandleFunc("/solve/{number}", ticket.Solve(db))
	r.HandleFunc("/view/{number}", ticket.Retrieve(db))

	rr := httptest.NewServer(r)

	require.NotNil(t, db)
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
	db := new(mockDB)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/view/open", nil)

	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()

	http.HandlerFunc(ticket.Retrieve10(db, database.StatusOpen)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func TestViewSolvedHandler(t *testing.T) {
	var err error
	db := new(mockDB)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/view/solved", nil)

	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()

	http.HandlerFunc(ticket.Retrieve10(db, database.StatusSolved)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func TestAdmin(t *testing.T) {
	var err error
	db := new(mockDB)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/admin", nil)

	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()

	http.HandlerFunc(admin.Admin(db)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

/*
//////////////////////////////////////////////////////////////
// Benchmarks
//////////////////////////////////////////////////////////////
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
*/
