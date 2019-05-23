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
	ZDTicket  int        `schema:"zdticket,required"`
	UserID    int        `schema:"userid,required"`
	Issue     IssueType  `schema:"issue,required"`
	Initials  string     `schema:"initials,required"`
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
	ForceDowngrade
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
	case ForceDowngrade:
		return "Force Downgrade/Cancellation"
	case UndoDowngrade:
		return "Undo Downgrade/Cancellation"
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
	Type     string
	Tickets  []Ticket
	NextPage int64
	PrevPage int64
	Status   StatusType
}

func (tp *TicketsPage) SetPages(page int64) {
	if len(tp.Tickets) == 10 {
		tp.NextPage = page + 1
	}

	// Set previous page
	tp.PrevPage = page - 1
}
