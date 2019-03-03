package ticket

import (
	"time"

	"../database"
)

func (t Ticket) updateToDB() error {
	query := `UPDATE tickets
		SET zdticket=?,
		userid=?,
		issue=?,
		initials=?,
		status=?
		WHERE ticket_id=?`

	_, err := database.DBCon.Exec(query, t.ZDTicket, t.UserID, t.Issue, t.Initials, t.Status, t.Number)
	if err != nil {
		return err
	}

	return nil

}

func (t *Ticket) addToDB() error {
	query := `INSERT INTO tickets (zdticket, userid, issue, initials, status, submitted)
		VALUES (?, ?, ?, ?, ?, ?);`
	result, err := database.DBCon.Exec(query, t.ZDTicket, t.UserID, t.Issue, t.Initials, t.Status, t.Submitted.Unix())
	if err != nil {
		return err
	}
	t.Number, err = result.LastInsertId()
	if err != nil {
		return err
	}
	t.Comment.TicketNumber = t.Number

	err = t.Comment.AddToDB()
	if err != nil {
		return err
	}

	return nil
}

func getFromDB(num int64) (Ticket, error) {
	query := `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted FROM tickets
		WHERE ticket_id=?`
	r := database.DBCon.QueryRow(query, num)
	t := Ticket{}
	var timestamp int64
	err := r.Scan(&t.Number, &t.ZDTicket, &t.UserID, &t.Issue, &t.Initials, &t.Status, &timestamp)
	if err != nil {
		return Ticket{}, err
	}
	t.Submitted = time.Unix(timestamp, 0)

	err = t.Comment.GetFromDB(num)
	if err != nil {
		return Ticket{}, err
	}

	return t, nil
}

func getNext10FromDB(offset int64, status StatusType) (ts []Ticket, err error) {
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
	r, err := database.DBCon.Query(query, status, offset)
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

func getMCTicketsFromDB(startTime int64, endTime int64) (ts []Ticket, err error) {
	query := `SELECT ticket_id, zdticket, userid, issue, initials, status, submitted 
		FROM tickets 
		WHERE issue=4
		AND submitted >= ?
		AND submitted < ?`
	_ = query

	r, err := database.DBCon.Query(query, startTime, endTime)
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
