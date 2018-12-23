package ticket

import (
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
	query := `INSERT INTO tickets (zdticket, userid, issue, initials, status)
		VALUES (?, ?, ?, ?, 0);`
	result, err := database.DBCon.Exec(query, t.ZDTicket, t.UserID, t.Issue, t.Initials)
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
	query := `SELECT ticket_id, zdticket, userid, issue, initials, status FROM tickets
		WHERE ticket_id=?`
	r := database.DBCon.QueryRow(query, num)
	t := Ticket{}
	err := r.Scan(&t.Number, &t.ZDTicket, &t.UserID, &t.Issue, &t.Initials, &t.Status)
	if err != nil {
		return Ticket{}, err
	}

	err = t.Comment.GetFromDB(num)
	if err != nil {
		return Ticket{}, err
	}

	return t, nil
}

func getNext10FromDB(lastTicket int64, status StatusType) (ts []Ticket, err error) {
	// Select rows with limit
	query := `SELECT ticket_id, zdticket, userid, issue, initials, status 
		FROM tickets 
		WHERE ticket_id>? AND status=?
		LIMIT 10`
	r, err := database.DBCon.Query(query, lastTicket, status)
	if err != nil {
		return ts, err
	}

	t := Ticket{}
	for r.Next() {
		err = r.Scan(&t.Number, &t.ZDTicket, &t.UserID, &t.Issue, &t.Initials, &t.Status)
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

func getRowsFound(lastTicket int64, status StatusType) (rowsFound int64, err error) {
	query := `SELECT COUNT(*) 
		FROM tickets 
		WHERE ticket_id>? AND status=?`
	count := database.DBCon.QueryRow(query, lastTicket, status)

	err = count.Scan(&rowsFound)
	if err != nil {
		return rowsFound, err
	}

	return rowsFound, nil
}
