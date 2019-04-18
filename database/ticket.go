package database

import (
	"fmt"
	"time"

	. "../structs"
)

func (d *DB) UpdateTicketToDB(t Ticket) error {
	query := `UPDATE tickets
		SET zdticket=?,
		userid=?,
		issue=?,
		initials=?,
		status=?
		WHERE ticket_id=?`

	_, err := d.Exec(query, t.ZDTicket, t.UserID, t.Issue, t.Initials, t.Status, t.Number)
	if err != nil {
		return err
	}

	return nil

}

func (d *DB) AddTicketToDB(t Ticket) (Ticket, error) {
	query := `INSERT INTO tickets (zdticket, userid, issue, initials, status, submitted)
		VALUES (?, ?, ?, ?, ?, ?);`
	result, err := d.Exec(query, t.ZDTicket, t.UserID, t.Issue, t.Initials, t.Status, t.Submitted.Unix())
	if err != nil {
		return t, err
	}
	t.Number, err = result.LastInsertId()
	if err != nil {
		return t, err
	}
	t.Comment.TicketNumber = t.Number

	err = d.AddCommentToDB(t.Comment)
	if err != nil {
		return t, err
	}

	return t, nil
}

func (d *DB) GetTicketFromDB(num int64) (Ticket, error) {
	query := `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted FROM tickets
		WHERE ticket_id=?`
	r := d.QueryRow(query, num)
	t := Ticket{}
	var timestamp int64
	err := r.Scan(&t.Number, &t.ZDTicket, &t.UserID, &t.Issue, &t.Initials, &t.Status, &timestamp)
	if err != nil {
		return Ticket{}, err
	}
	t.Submitted = time.Unix(timestamp, 0)

	t.Comment, err = d.GetCommentFromDB(num)
	if err != nil {
		return Ticket{}, err
	}

	return t, nil
}

func (d *DB) GetNext10TicketsFromDB(offset int64, status StatusType) (ts []Ticket, err error) {
	// Select rows with limit
	var query string
	switch status {
	case StatusOpen:
		query = `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted 
		FROM tickets 
		WHERE status=? AND issue<>4
		LIMIT ?, 10`
	case StatusSolved:
		query = `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted 
		FROM tickets 
		WHERE status=? AND issue<>4
		ORDER BY ticket_id DESC
		LIMIT ?, 10`
	}
	r, err := d.Query(query, status, offset)
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

func (d *DB) GetMCTicketsFromDB(startTime int64, endTime int64) (ts []Ticket, err error) {
	query := `SELECT t.ticket_id, t.zdticket, t.userid, t.issue, t.initials, t.status, t.submitted, c.text 
		FROM tickets t
		INNER JOIN
		comments c ON t.ticket_id = c.ticket_id
		WHERE t.issue=4
		AND t.submitted >= ?
		AND t.submitted < ?`
	_ = query

	r, err := d.Query(query, startTime, endTime)
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
		fmt.Println("WORKS")
	}
	if r.Err() != nil {
		return ts, err
	}
	return ts, nil
}
