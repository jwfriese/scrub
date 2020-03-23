package scrub

type Selector struct {
	Table  string
	Column string
	Wheres string
}

type Config struct {
	Selectors []Selector
	Method    Method
}
