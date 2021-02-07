package gworker

// Arg ...
type Arg struct {
	Type  string
	Value interface{}
}

// Job ...
type Job struct {
	ID           string
	Name         string
	Args         []Arg
	Retry        uint
	RetryTimeout uint
}
