package athenz

type SettingType uint8

const (
	EXPIRATION = iota
	REVIEW
)

func (s SettingType) String() string {
	switch s {
	case EXPIRATION:
		return "expiration"
	case REVIEW:
		return "review"
	default:
		return "Invalid setting type"
	}
}
