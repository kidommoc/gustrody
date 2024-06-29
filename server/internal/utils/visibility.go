package utils

type Vsb uint

const (
	Vsb_PUBLIC Vsb = iota
	Vsb_FOLLOWER
	Vsb_DIRECT
)

func GetVsb(literal string) (vsb Vsb, ok bool) {
	switch literal {
	case "public":
		return Vsb_PUBLIC, true
	case "follower":
		return Vsb_FOLLOWER, true
	case "direct":
		return Vsb_DIRECT, true
	default:
		return Vsb_PUBLIC, false
	}
}

func (v Vsb) String() string {
	switch v {
	case Vsb_PUBLIC:
		return "public"
	case Vsb_FOLLOWER:
		return "follower"
	case Vsb_DIRECT:
		return "direct"
	}
	return "public"
}
