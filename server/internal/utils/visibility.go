package utils

type Vsb string

const (
	Vsb_PUBLIC   Vsb = "public"
	Vsb_FOLLOWER Vsb = "follower"
	Vsb_DIRECT   Vsb = "direct"
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
