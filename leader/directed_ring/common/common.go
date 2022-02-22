package common

type ModeType string

const (
	Pass      ModeType = "pass"
	Unknown            = "unknown"
	NonLeader          = "nonleader"
	Leader             = "leader"
	Relay              = "relay"
)
