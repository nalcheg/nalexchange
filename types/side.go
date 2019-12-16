package types

type Side uint8

const (
	Sell Side = iota
	Buy
)

func (d Side) String() string {
	return [...]string{"Sell", "Buy"}[d]
}
