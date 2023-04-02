package athenz

type MemberType uint8

const (
	USER = iota
	GROUP
	SERVICE
)

func (s MemberType) String() string {
	switch s {
	case USER:
		return "user"
	case GROUP:
		return "group"
	case SERVICE:
		return "service"
	default:
		return "Invalid member type"
	}
}
