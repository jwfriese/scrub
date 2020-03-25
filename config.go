package scrub

type Selector struct {
	// if empty string, no database will be appended to tables in scrubber's queries
	Database string

	Table    string
	Column   string

	// if empty string, no wheres will be used in scrubber's queries
	Wheres   string
}

type Config struct {
	Selectors []Selector
	Method    Method
}
