package types

type MgmtMessage uint8

const (
	List MgmtMessage = iota
	Flush
)

func (d MgmtMessage) String() string {
	return [...]string{"List", "Flush"}[d]
}
