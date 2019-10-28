package structs

import "time"

type Comment struct {
	ID           int64     `schema:"-"`
	Timestamp    time.Time `schema:"-"`
	Text         string    `schema:"text"`
	TicketNumber int64     `schema:"-"`
}

// Ticket structure
type Ticket struct {
	Number    int64
	ZDTicket  int        `schema:"zdticket,required"`
	UserID    int        `schema:"userid,required"`
	Issue     IssueType  `schema:"issue,required"`
	Initials  string     `schema:"initials,required"`
	Status    StatusType `schema:"-"`
	Submitted time.Time  `schema:"-"`
	Comment   Comment    `schema:"comment"`
}

func (t *Ticket) Patch(updatedTicket Ticket) {
	t.ZDTicket = updatedTicket.ZDTicket
	t.UserID = updatedTicket.UserID
	t.Issue = updatedTicket.Issue
	t.Initials = updatedTicket.Initials
	t.Status = updatedTicket.Status
	t.Comment.Text = updatedTicket.Comment.Text
}

type IssueType int

const (
	Refund         = iota // 0
	Terminated            // 1
	DNAFP                 // 2
	Extension             // 3
	MCContacts            // 4 (DEPRECATED)
	Discount              // 5
	ForceDowngrade        // 6
	UndoDowngrade         // 7
	LegacyContacts        // 8
	TNEContacts           // 9
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
		return "MC Contacts (deprecated)"
	case Discount:
		return "Discount"
	case ForceDowngrade:
		return "Force Downgrade/Cancellation"
	case UndoDowngrade:
		return "Undo Downgrade/Cancellation"
	case LegacyContacts:
		return "Legacy Contacts"
	case TNEContacts:
		return "TNE Contacts"
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

// Tickets page struct for paginating
type TicketsPage struct {
	SolvedTicket int64
	Type         string
	Tickets      []Ticket
	NextPage     int64
	PrevPage     int64
	Status       StatusType
}

func (tp *TicketsPage) SetPrevAndNextPage(page int64) {
	if len(tp.Tickets) == 10 {
		tp.NextPage = page + 1
	}

	// Set previous page
	tp.PrevPage = page - 1
}

var CXIssues = []IssueType{Refund, Terminated, DNAFP, Extension}
var LeadIssues = []IssueType{Discount, ForceDowngrade, UndoDowngrade}
var AllIssues = []IssueType{Refund, Terminated, DNAFP, Extension, Discount, ForceDowngrade, UndoDowngrade}
