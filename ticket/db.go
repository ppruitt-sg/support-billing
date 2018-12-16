package ticket

import (
	"database/sql"
	"os"
)

/* func getAllFromDB() ([]Ticket, error) {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	var ts []Ticket
	if err != nil {
		return []Ticket{}, err
	}

	query := `SELECT ticket_id, zdticket, userid, issue, initials, solved FROM tickets`
	r, err := db.Query(query)
	if err != nil {
		return ts, err
	}
	t := Ticket{}
	for r.Next() {
		err = r.Scan(&t.Number, &t.ZDNum, &t.UserID, &t.Issue, &t.Initials, &t.Solved)
		if err != nil {
			return ts, err
		}
		ts = append(ts, t)
	}
	if r.Err() != nil {
		return ts, r.Err()
	}

	return ts, nil
} */

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
		solved=?
		WHERE ticket_id=?`

	_, err = db.Exec(query, t.ZDNum, t.UserID, t.Issue, t.Initials, t.Solved, t.Number)
	if err != nil {
		return err
	}

	return nil

}

func (t *Ticket) addToDB() error {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()

	query := `INSERT INTO tickets (zdticket, userid, issue, initials, solved)
		VALUES (?, ?, ?, ?, 0);`
	result, err := db.Exec(query, t.ZDNum, t.UserID, t.Issue, t.Initials)
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

	query := `SELECT ticket_id, zdticket, userid, issue, initials, solved FROM tickets
		WHERE ticket_id=?`
	r := db.QueryRow(query, num)
	t := Ticket{}
	err = r.Scan(&t.Number, &t.ZDNum, &t.UserID, &t.Issue, &t.Initials, &t.Solved)
	if err != nil {
		return Ticket{}, err
	}

	err = t.Comment.GetFromDB(num)
	if err != nil {
		return Ticket{}, err
	}

	return t, nil
}

func getNext5FromDB(lastTicket int64, solved bool) ([]Ticket, error) {
	var ts []Ticket
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return ts, err
	}

	// Select rows with limit
	query := `SELECT ticket_id, zdticket, userid, issue, initials, solved 
		FROM tickets 
		WHERE ticket_id>? AND solved=?
		LIMIT 5`
	r, err := db.Query(query, lastTicket, solved)
	if err != nil {
		return ts, err
	}

	t := Ticket{}
	for r.Next() {
		err = r.Scan(&t.Number, &t.ZDNum, &t.UserID, &t.Issue, &t.Initials, &t.Solved)
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

func getRowsFound(lastTicket int64, solved bool) (int64, error) {
	var rowsFound int64

	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return rowsFound, err
	}

	query := `SELECT COUNT(*) 
		FROM tickets 
		WHERE ticket_id>? AND solved=?`
	count := db.QueryRow(query, lastTicket, solved)

	err = count.Scan(&rowsFound)
	if err != nil {
		return rowsFound, err
	}

	return rowsFound, nil
}
