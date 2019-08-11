package database

import (
	"time"

	. "github.com/ppruitt-sg/support-billing/structs"
)

const queryAddComment = `INSERT INTO comments (timestamp, text, ticket_id)
	VALUES (?, ?, ?)`

const querySelectComment = `SELECT comment_id, timestamp, text, ticket_id FROM comments
	WHERE ticket_id=?`

func (d *DB) AddComment(c Comment) (err error) {
	_, err = d.Exec(queryAddComment, c.Timestamp.Unix(), c.Text, c.TicketNumber)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) GetComment(num int64) (c Comment, err error) {
	r := d.QueryRow(querySelectComment, num)
	var ts int64
	err = r.Scan(&c.ID, &ts, &c.Text, &c.TicketNumber)
	if err != nil {
		return c, err
	}

	c.Timestamp = time.Unix(ts, 0)

	return c, nil
}

func (d *DB) UpdateComment(c Comment) (err error) {
	queryUpdateComment := `UPDATE comments
	SET text=?
	WHERE comment_id=?`

	_, err = d.Exec(queryUpdateComment, c.Text, c.ID)
	if err != nil {
		return err
	}

	return nil
}
