package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	. "github.com/ppruitt-sg/support-billing/structs"
)

const queryMCTickets = `SELECT t.userid, c.text 
	FROM tickets t
	INNER JOIN
	comments c ON t.ticket_id = c.ticket_id
	WHERE t.issue=4
	AND t.submitted >= ?
	AND t.submitted < ?`

const querySelectTicket = `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted FROM tickets
	WHERE ticket_id=?`

const queryUpdateTicket = `UPDATE tickets
	SET zdticket=?,
	userid=?,
	issue=?,
	initials=?,
	status=?
	WHERE ticket_id=?`

const queryAddTicket = `INSERT INTO tickets (zdticket, userid, issue, initials, status, submitted)
	VALUES (?, ?, ?, ?, ?, ?);`

const querySelect10Open = `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted 
	FROM tickets 
	WHERE status=? AND issue IN (%s)
	LIMIT ?, 10`

const querySelect10Closed = `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted 
	FROM tickets 
	WHERE status=? AND issue IN (%s)
	ORDER BY ticket_id DESC
	LIMIT ?, 10`

func (d *DB) getTicketsFromRows(r *sql.Rows) (ts []Ticket, err error) {
	t := Ticket{}
	var timestamp int64
	for r.Next() {
		err = r.Scan(&t.Number, &t.ZDTicket, &t.UserID, &t.Issue, &t.Initials, &t.Status, &timestamp)
		if err != nil {
			return ts, err
		}
		// Convert int64 to time.Time
		t.Submitted = time.Unix(timestamp, 0)
		ts = append(ts, t)
	}
	if r.Err() != nil {
		return ts, err
	}
	return ts, nil
}

func (d *DB) UpdateTicket(t Ticket) error {
	_, err := d.Exec(queryUpdateTicket, t.ZDTicket, t.UserID, t.Issue, t.Initials, t.Status, t.Number)
	if err != nil {
		return err
	}

	return nil

}

func (d *DB) AddTicket(t Ticket) (Ticket, error) {
	result, err := d.Exec(queryAddTicket, t.ZDTicket, t.UserID, t.Issue, t.Initials, t.Status, t.Submitted.Unix())
	if err != nil {
		return t, err
	}
	t.Number, err = result.LastInsertId()
	if err != nil {
		return t, err
	}
	t.Comment.TicketNumber = t.Number

	err = d.AddComment(t.Comment)
	if err != nil {
		return t, err
	}

	return t, nil
}

func (d *DB) GetTicket(num int64) (t Ticket, err error) {
	var timestamp int64

	r := d.QueryRow(querySelectTicket, num)
	err = r.Scan(&t.Number, &t.ZDTicket, &t.UserID, &t.Issue, &t.Initials, &t.Status, &timestamp)
	if err != nil {
		return Ticket{}, err
	}
	t.Submitted = time.Unix(timestamp, 0)

	t.Comment, err = d.GetComment(num)
	if err != nil {
		return Ticket{}, err
	}

	return t, nil
}

func (d *DB) GetNext10Tickets(offset int64, status StatusType, issues ...IssueType) (ts []Ticket, err error) {
	// Select rows with limit
	var query string

	// Convert slice of Issues to comma separated numbers
	s, _ := json.Marshal(issues)
	strIssues := strings.Trim(string(s), "[]")

	switch status {
	case StatusOpen:
		// If status is open list in ascending order
		query = fmt.Sprintf(querySelect10Open, strIssues)
	case StatusSolved:
		// If status is solved list in descending order
		query = fmt.Sprintf(querySelect10Closed, strIssues)
	}

	r, err := d.Query(query, status, offset)
	if err != nil {
		return ts, err
	}

	// Create tickets and add to tickets slice
	ts, err = d.getTicketsFromRows(r)
	if err != nil {
		return ts, err
	}

	return ts, nil
}

func (d *DB) GetMCTickets(startTime int64, endTime int64) (ts []Ticket, err error) {
	t := Ticket{}
	r, err := d.Query(queryMCTickets, startTime, endTime)
	if err != nil {
		return ts, err
	}

	// Add ticket userID and ticket Comment text
	for r.Next() {
		err = r.Scan(&t.UserID, &t.Comment.Text)
		if err != nil {
			return ts, err
		}
		ts = append(ts, t)
	}
	if r.Err() != nil {
		return ts, err
	}

	return ts, nil
}
