package database

import (
	"strings"
	"time"

	. "github.com/ppruitt-sg/support-billing/structs"
)

const queryMCTickets = `SELECT t.ticket_id, t.zdticket, t.userid, t.issue, t.initials, t.status, t.submitted, c.text 
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

func (d *DB) GetTicket(num int64) (Ticket, error) {
	r := d.QueryRow(querySelectTicket, num)
	t := Ticket{}
	var timestamp int64
	err := r.Scan(&t.Number, &t.ZDTicket, &t.UserID, &t.Issue, &t.Initials, &t.Status, &timestamp)
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
	switch status {
	case StatusOpen:
		// If status is open list in ascending order
		query = `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted 
		FROM tickets 
		WHERE status=? AND issue IN (?` + strings.Repeat(`,?`, len(issues)-1) + `)
		LIMIT ?, 10`
	case StatusSolved:
		// If status is solved list in descending order
		query = `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted 
		FROM tickets 
		WHERE status=? AND issue IN (?` + strings.Repeat(`,?`, len(issues)-1) + `)
		ORDER BY ticket_id DESC
		LIMIT ?, 10`
	}
	// Create and append args for DB query
	args := []interface{}{status}
	for _, issue := range issues {
		args = append(args, issue)
	}
	args = append(args, offset)

	r, err := d.Query(query, args...)
	if err != nil {
		return ts, err
	}

	t := Ticket{}
	var timestamp int64
	// Create tickets and add to tickets slice
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

func (d *DB) GetMCTickets(startTime int64, endTime int64) (ts []Ticket, err error) {
	r, err := d.Query(queryMCTickets, startTime, endTime)
	if err != nil {
		return ts, err
	}

	t := Ticket{}
	var timestamp int64
	// Create tickets and add to tickets slice
	for r.Next() {
		err = r.Scan(&t.Number, &t.ZDTicket, &t.UserID, &t.Issue, &t.Initials, &t.Status, &timestamp, &t.Comment.Text)
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
