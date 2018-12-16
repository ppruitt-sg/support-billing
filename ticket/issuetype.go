package ticket

type IssueType int

const (
	Refund     IssueType = 0
	Terminated IssueType = 1
	DNAFP      IssueType = 2
	Extension  IssueType = 3
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
	default:
		return ""
	}
}
