package structs

import "time"

type Comment struct {
	Timestamp    time.Time
	Text         string `schema:"text"`
	TicketNumber int64
}

// Ticket structure
type Ticket struct {
	Number    int64      `schema:"-"`
	ZDTicket  int        `schema:"zdticket"`
	UserID    int        `schema:"userid"`
	Issue     IssueType  `schema:"issue"`
	Initials  string     `schema:"initials"`
	Status    StatusType `schema:"-"`
	Submitted time.Time
	Comment   Comment `schema:"comment"`
}

type IssueType int

const (
	Refund     IssueType = 0
	Terminated IssueType = 1
	DNAFP      IssueType = 2
	Extension  IssueType = 3
	MCContacts IssueType = 4
)

func (i IssueType) ToString() string {
	switch i {
	case 0:
		return "Refund"
	case 1:
		return "Billing Terminated"
	case 2:
		return "DNA FP"
	case 3:
		return "Extension"
	case 4:
		return "MC Contacts"
	default:
		return ""
	}
}

type StatusType int

const (
	StatusOpen   StatusType = 0
	StatusSolved StatusType = 1
)

func (s StatusType) ToString() string {
	switch s {
	case 0:
		return "Open"
	case 1:
		return "Solved"
	default:
		return ""
	}
}

// Tickets page structure for paginating
type TicketsPage struct {
	Tickets    []Ticket
	NextButton bool
	NextPage   int64
	PrevPage   int64
	PrevButton bool
	Status     StatusType
}
