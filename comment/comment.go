package comment

import (
	"time"

	"../database"
)

type Comment struct {
	Timestamp    time.Time
	Text         string `schema:"text"`
	TicketNumber int64
}

func (c Comment) AddToDB() (err error) {
	query := `INSERT INTO comments (timestamp, text, ticket_id)
		VALUES (?, ?, ?)`
	_, err = database.DBCon.Exec(query, c.Timestamp.Unix(), c.Text, c.TicketNumber)
	if err != nil {
		return err
	}
	return nil
}

func (c *Comment) GetFromDB(num int64) (err error) {
	query := `SELECT timestamp, text, ticket_id FROM comments
		WHERE ticket_id=?`

	r := database.DBCon.QueryRow(query, num)
	var ts int64
	err = r.Scan(&ts, &c.Text, &c.TicketNumber)
	if err != nil {
		return err
	}

	c.Timestamp = time.Unix(ts, 0)

	return nil
}
