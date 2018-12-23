package ticket

import (
	"database/sql"
	"os"
)

func (t Ticket) updateToDB() error {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `UPDATE tickets
		SET zdticket=?,
		userid=?,
		issue=?,
		initials=?,
		status=?
		WHERE ticket_id=?`

	_, err = db.Exec(query, t.ZDTicket, t.UserID, t.Issue, t.Initials, t.Status, t.Number)
	if err != nil {
		return err
	}

	return nil

}

func (t *Ticket) addToDB() error {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()

	query := `INSERT INTO tickets (zdticket, userid, issue, initials, status)
		VALUES (?, ?, ?, ?, 0);`
	result, err := db.Exec(query, t.ZDTicket, t.UserID, t.Issue, t.Initials)
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
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return Ticket{}, err
	}

	query := `SELECT ticket_id, zdticket, userid, issue, initials, status FROM tickets
		WHERE ticket_id=?`
	r := db.QueryRow(query, num)
	t := Ticket{}
	err = r.Scan(&t.Number, &t.ZDTicket, &t.UserID, &t.Issue, &t.Initials, &t.Status)
	if err != nil {
		return Ticket{}, err
	}

	err = t.Comment.GetFromDB(num)
	if err != nil {
		return Ticket{}, err
	}

	return t, nil
}

func getNext10FromDB(lastTicket int64, status StatusType) ([]Ticket, error) {
	var ts []Ticket
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return ts, err
	}

	// Select rows with limit
	query := `SELECT ticket_id, zdticket, userid, issue, initials, status 
		FROM tickets 
		WHERE ticket_id>? AND status=?
		LIMIT 10`
	r, err := db.Query(query, lastTicket, status)
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

func getRowsFound(lastTicket int64, status StatusType) (int64, error) {
	var rowsFound int64

	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return rowsFound, err
	}

	query := `SELECT COUNT(*) 
		FROM tickets 
		WHERE ticket_id>? AND status=?`
	count := db.QueryRow(query, lastTicket, status)

	err = count.Scan(&rowsFound)
	if err != nil {
		return rowsFound, err
	}

	return rowsFound, nil
}
