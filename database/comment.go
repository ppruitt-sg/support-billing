package database

import (
	"time"

	. "../structs"
)

func (d *DB) AddCommentToDB(c Comment) (err error) {
	query := `INSERT INTO comments (timestamp, text, ticket_id)
		VALUES (?, ?, ?)`
	_, err = d.Exec(query, c.Timestamp.Unix(), c.Text, c.TicketNumber)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) GetCommentFromDB(num int64) (c Comment, err error) {
	query := `SELECT timestamp, text, ticket_id FROM comments
		WHERE ticket_id=?`

	r := d.QueryRow(query, num)
	var ts int64
	err = r.Scan(&ts, &c.Text, &c.TicketNumber)
	if err != nil {
		return c, err
	}

	c.Timestamp = time.Unix(ts, 0)

	return c, nil
}
