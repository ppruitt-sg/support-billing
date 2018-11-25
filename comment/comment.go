package comment

import (
	"database/sql"
	"os"
	"time"
)

type Comment struct {
	Timestamp    time.Time
	Text         string `schema:"text"`
	TicketNumber int64
}

func (c Comment) AddToDB() error {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	query := `INSERT INTO comments (timestamp, text, ticket_id)
		VALUES (?, ?, ?)`
	_, err = db.Exec(query, c.Timestamp.Unix(), c.Text, c.TicketNumber)
	if err != nil {
		return err
	}
	return nil
}

func GetFromDB(num int64) (c Comment, err error) {
	db, err := sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/supportbilling")
	defer db.Close()
	if err != nil {
		return c, err
	}

	query := `SELECT timestamp, text, ticket_id FROM comments
		WHERE ticket_id=?`

	r := db.QueryRow(query, num)
	var ts int64
	err = r.Scan(&ts, &c.Text, &c.TicketNumber)
	c.Timestamp = time.Unix(ts, 0)
	if err != nil {
		return c, err
	}

	return c, nil
}
