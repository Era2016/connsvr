package model

type Demo struct {
	Key   string
	Value string
}

func NewDemo() *Demo {
	s := &Demo{}
	return s
}
