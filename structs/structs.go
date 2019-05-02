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
	Refund = iota
	Terminated
	DNAFP
	Extension
	MCContacts
	Discount
	Downgrade
	UndoDowngrade
)

func (i IssueType) String() string {
	switch i {
	case Refund:
		return "Refund"
	case Terminated:
		return "Billing Terminated"
	case DNAFP:
		return "DNA FP Reactivation"
	case Extension:
		return "Extension"
	case MCContacts:
		return "MC Contacts"
	case Discount:
		return "Discount"
	case Downgrade:
		return "Force Downgrade"
	case UndoDowngrade:
		return "Undo Downgrade"
	default:
		return "[undefined]"
	}
}

type StatusType int

const (
	StatusOpen = iota
	StatusSolved
)

func (s StatusType) String() string {
	switch s {
	case StatusOpen:
		return "Open"
	case StatusSolved:
		return "Solved"
	default:
		return "[undefined]"
	}
}

// Tickets page structure for paginating
type TicketsPage struct {
	Type       string
	Tickets    []Ticket
	NextButton bool
	NextPage   int64
	PrevPage   int64
	PrevButton bool
	Status     StatusType
}
