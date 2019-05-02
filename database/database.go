package database

import (
	"database/sql"

	. "../structs"
)

type DB struct {
	*sql.DB
}

type Datastore interface {
	// Comment
	AddCommentToDB(Comment) error
	GetCommentFromDB(int64) (Comment, error)
	// Ticket
	UpdateTicketToDB(Ticket) error
	AddTicketToDB(Ticket) (Ticket, error)
	GetTicketFromDB(int64) (Ticket, error)
	GetNext10TicketsFromDB(int64, StatusType, ...IssueType) ([]Ticket, error)
	GetMCTicketsFromDB(int64, int64) ([]Ticket, error)
}

func (d *DB) NewDB(dataSourceName string) (err error) {
	d.DB, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	return nil
}
