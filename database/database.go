package database

import (
	"database/sql"

	. "github.com/ppruitt-sg/support-billing/structs"
)

type DB struct {
	*sql.DB
}

type Datastore interface {
	// Comment
	AddComment(Comment) error
	GetComment(int64) (Comment, error)
	UpdateComment(Comment) (err error)
	// Ticket
	UpdateTicket(Ticket) error
	AddTicket(Ticket) (Ticket, error)
	GetTicket(int64) (Ticket, error)
	GetNext10Tickets(int64, StatusType, ...IssueType) ([]Ticket, error)
	GetMCTickets(int64, int64) ([]Ticket, []Ticket, error)
}

func (d *DB) NewDB(dataSourceName string) (err error) {
	d.DB, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	return nil
}
