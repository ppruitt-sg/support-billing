package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	. "github.com/ppruitt-sg/support-billing/structs"
)

const queryMCTickets = `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted 
	FROM tickets
	WHERE issue=?
	AND submitted >= ?
	AND submitted < ?`

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

const querySelectAll = `SELECT t.ticket_id, t.zdticket, t.userid, t.issue, t.initials, t.status, t.submitted, c.text
	FROM tickets AS t
	INNER JOIN comments AS c 
	ON t.ticket_id = c.ticket_id`

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

	err = d.UpdateComment(t.Comment)
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

func (d *DB) GetMCTickets(startTime int64, endTime int64) (legacyTickets []Ticket, tneTickets []Ticket, err error) {

	// Get Legacy Tickets from DB
	r, err := d.Query(queryMCTickets, LegacyContacts, startTime, endTime)
	if err != nil {
		return legacyTickets, tneTickets, err
	}

	legacyTickets, err = d.getTicketsFromRows(r)
	if r.Err() != nil {
		return legacyTickets, tneTickets, err
	}

	// Get TNE Tickets from DB
	r, err = d.Query(queryMCTickets, TNEContacts, startTime, endTime)
	if err != nil {
		return legacyTickets, tneTickets, err
	}

	tneTickets, err = d.getTicketsFromRows(r)
	if r.Err() != nil {
		return legacyTickets, tneTickets, err
	}

	return legacyTickets, tneTickets, nil
}

func (d *DB) Export() (tickets []Ticket, err error) {
	r, err := d.Query(querySelectAll)
	if err != nil {
		return tickets, err
	}

	// Parse rows and set up on tickets
	var timestamp int64
	t := Ticket{}
	for r.Next() {
		err = r.Scan(&t.Number, &t.ZDTicket, &t.UserID, &t.Issue, &t.Initials, &t.Status, &timestamp, &t.Comment.Text)
		if err != nil {
			return tickets, err
		}
		// Convert int64 to time.Time
		t.Submitted = time.Unix(timestamp, 0)
		tickets = append(tickets, t)
	}
	if r.Err() != nil {
		return tickets, err
	}
	return tickets, nil

}
