package main

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/ppruitt-sg/support-billing/structs"

	"github.com/ppruitt-sg/support-billing/routes"
)

////////////////////////////////
// Mock DB
////////////////////////////////
type FakeTicketStruct struct {
	FakeTicket Ticket
	FakeError  error
}

type FakeTicketSliceStruct struct {
	FakeTicketSlice []Ticket
	FakeError       error
}

type FakeCommentStruct struct {
	FakeComment Comment
	FakeError   error
}
type mockDB struct {
	FakeTicketStruct
	FakeTicketSliceStruct
	FakeCommentStruct
}

func (d mockDB) AddComment(c Comment) error {
	return d.FakeCommentStruct.FakeError
}

func (d mockDB) GetComment(stamp int64) (Comment, error) {
	return d.FakeCommentStruct.FakeComment, d.FakeCommentStruct.FakeError
}

func (d mockDB) UpdateComment(c Comment) error {
	return d.FakeCommentStruct.FakeError
}

func (d mockDB) UpdateTicket(t Ticket) error {
	return d.FakeTicketStruct.FakeError
}

func (d mockDB) AddTicket(t Ticket) (Ticket, error) {
	return d.FakeTicketStruct.FakeTicket, d.FakeTicketStruct.FakeError
}

func (d mockDB) GetTicket(number int64) (Ticket, error) {
	return d.FakeTicketStruct.FakeTicket, d.FakeTicketStruct.FakeError
}

func (d mockDB) GetNext10Tickets(offset int64, status StatusType, issues ...IssueType) ([]Ticket, error) {
	return d.FakeTicketSliceStruct.FakeTicketSlice, d.FakeTicketSliceStruct.FakeError
}

func (d mockDB) GetMCTickets(start int64, end int64) ([]Ticket, []Ticket, error) {
	return d.FakeTicketSliceStruct.FakeTicketSlice, d.FakeTicketSliceStruct.FakeTicketSlice, d.FakeTicketSliceStruct.FakeError
}

var expectedDB = mockDB{
	FakeTicketStruct{ // Fake Ticket
		Ticket{},
		nil,
	},
	FakeTicketSliceStruct{ // Fake Ticket Slice
		[]Ticket{},
		nil,
	},
	FakeCommentStruct{ // Fake Ticket
		Comment{},
		nil,
	},
}

var errorDB = mockDB{
	FakeTicketStruct{ // Fake Ticket
		Ticket{},
		errors.New("Generic error"),
	},
	FakeTicketSliceStruct{ // Fake Ticket Slice
		[]Ticket{},
		errors.New("Generic error"),
	},
	FakeCommentStruct{ // Fake Ticket
		Comment{},
		errors.New("Generic error"),
	},
}

var rowNotFoundDB = mockDB{
	FakeTicketStruct{ // Fake Ticket
		Ticket{},
		sql.ErrNoRows,
	},
	FakeTicketSliceStruct{ // Fake Ticket Slice
		[]Ticket{},
		sql.ErrNoRows,
	},
	FakeCommentStruct{ // Fake Ticket
		Comment{},
		sql.ErrNoRows,
	},
}

func TestNewHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/new", nil)
	rr := httptest.NewRecorder()

	require.Nil(t, err)
	require.NotNil(t, req)
	require.NotNil(t, rr)

	http.HandlerFunc(routes.New).ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
}

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	require.Nil(t, err)
	require.NotNil(t, req)
	require.NotNil(t, rr)

	http.HandlerFunc(routes.Home).ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
}

func TestViewTicketHandler(t *testing.T) {
	var rr *httptest.Server
	var r *mux.Router
	var tests = []struct {
		n        string // ticket number
		expected int
		db       mockDB
	}{
		{"1", 200, expectedDB},
		{"2", 404, rowNotFoundDB},
		{"", 404, expectedDB},
		{"1", 500, errorDB},
		{"25325235235235235235253", 500, expectedDB},
	}

	for _, test := range tests {
		r = mux.NewRouter()
		r.HandleFunc("/view/{number}", routes.Retrieve(test.db))

		rr = httptest.NewServer(r)

		require.NotNil(t, r)
		require.NotNil(t, rr)

		url := rr.URL + "/view/" + test.n
		resp, err := http.Get(url)
		require.Nil(t, err)

		assert.Equal(t, test.expected, resp.StatusCode)
	}
}

func TestSolveTicketHandler(t *testing.T) {
	var rr *httptest.Server
	var r *mux.Router
	var tests = []struct {
		n        string // ticket number
		expected int
		db       mockDB
	}{
		{"1", 200, expectedDB},
		{"2", 404, rowNotFoundDB},
		{"", 404, expectedDB},
		{"1", 500, errorDB},
		{"25325235235235235235253", 500, expectedDB},
	}

	for _, test := range tests {
		r = mux.NewRouter()
		r.HandleFunc("/solve/{number}", routes.Solve(test.db))
		r.HandleFunc("/view/{number}", routes.Retrieve(test.db))

		rr = httptest.NewServer(r)

		require.NotNil(t, r)
		require.NotNil(t, rr)
		url := rr.URL + "/view/" + test.n
		resp, err := http.Post(url, "", nil)
		require.Nil(t, err)

		assert.Equal(t, resp.StatusCode, test.expected)
	}
}

func TestUpdateTicketHandler(t *testing.T) {
	var rr *httptest.Server
	var r *mux.Router
	var tests = []struct {
		n        string // ticket number
		expected int
		db       mockDB
	}{
		{"1", 200, expectedDB},
		{"2", 404, rowNotFoundDB},
		{"", 404, expectedDB},
		{"1", 500, errorDB},
		{"25325235235235235235253", 500, expectedDB},
	}

	for _, test := range tests {
		r = mux.NewRouter()
		r.HandleFunc("/update/{number}", routes.Update(test.db))
		r.HandleFunc("/view/{number}", routes.Retrieve(test.db))

		rr = httptest.NewServer(r)

		require.NotNil(t, r)
		require.NotNil(t, rr)
		url := rr.URL + "/view/" + test.n
		resp, err := http.Post(url, "", nil)
		require.Nil(t, err)

		assert.Equal(t, resp.StatusCode, test.expected)
	}
}

func TestViewCXHandler(t *testing.T) {
	var rr *httptest.ResponseRecorder
	var tests = []struct {
		expected int
		db       mockDB
	}{
		{200, expectedDB},
		{500, errorDB},
	}

	req, err := http.NewRequest("GET", "/view/cx", nil)
	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	for _, test := range tests {
		rr = httptest.NewRecorder()

		http.HandlerFunc(routes.Retrieve10(test.db, StatusOpen)).ServeHTTP(rr, req)
		assert.Equal(t, rr.Code, test.expected)
	}
}

func TestViewSolvedHandler(t *testing.T) {
	var rr *httptest.ResponseRecorder
	var tests = []struct {
		expected int
		db       mockDB
	}{
		{200, expectedDB},
		{500, errorDB},
	}

	req, err := http.NewRequest("GET", "/view/solved", nil)
	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	for _, test := range tests {
		rr = httptest.NewRecorder()

		http.HandlerFunc(routes.Retrieve10(test.db, StatusSolved)).ServeHTTP(rr, req)
		assert.Equal(t, rr.Code, test.expected)
	}
}

func TestAdmin(t *testing.T) {
	var rr *httptest.ResponseRecorder
	var tests = []struct {
		expected int
		db       mockDB
	}{
		{200, expectedDB},
		{500, errorDB},
	}

	req, err := http.NewRequest("GET", "/admin", nil)
	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	for _, test := range tests {
		rr = httptest.NewRecorder()

		http.HandlerFunc(routes.Admin(test.db)).ServeHTTP(rr, req)
		assert.Equal(t, rr.Code, test.expected)
	}
}

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
		http.HandlerFunc(routes.New).ServeHTTP(rr, req)
	}
}

func BenchmarkHomeHandler(b *testing.B) {
	req, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		b.Errorf("An error occurred. %v", err)
	}

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		http.HandlerFunc(routes.New).ServeHTTP(rr, req)
	}
}

func BenchmarkViewTicketHandler(b *testing.B) {
	r := mux.NewRouter()
	r.HandleFunc("/view/{number}", routes.Retrieve(expectedDB))

	rr := httptest.NewServer(r)
	defer rr.Close()
	for i := 0; i < b.N; i++ {
		resp, _ := http.Post(rr.URL+"/view/1", "", nil)
		resp.Body.Close()
	}

}

func BenchmarkSolveTicketHandler(b *testing.B) {
	r := mux.NewRouter()
	r.HandleFunc("/view/{number}", routes.Retrieve(expectedDB))
	r.HandleFunc("/solve/{number}", routes.Solve(expectedDB))

	rr := httptest.NewServer(r)
	defer rr.Close()
	for i := 0; i < b.N; i++ {
		resp, _ := http.Post(rr.URL+"/solve/1", "", nil)
		resp.Body.Close()
	}

}

func BenchmarkViewCXHandler(b *testing.B) {
	req, err := http.NewRequest("GET", "/view/cx", nil)
	if err != nil {
		b.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		http.HandlerFunc(routes.Retrieve10(expectedDB, StatusOpen)).ServeHTTP(rr, req)
	}

}

func BenchmarkViewSolvedHandler(b *testing.B) {
	req, err := http.NewRequest("GET", "/view/solved", nil)
	if err != nil {
		b.Errorf("An error occurred. %v", err)
	}

	rr := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		http.HandlerFunc(routes.Retrieve10(expectedDB, StatusSolved)).ServeHTTP(rr, req)
	}
}
