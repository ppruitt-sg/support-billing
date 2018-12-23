package ticket

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
