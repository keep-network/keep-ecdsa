package local2

type localIdentifier string

func (li localIdentifier) String() string {
	return string(li)
}
