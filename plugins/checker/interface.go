package checker

// Status consists of status code from the service
type Status struct {
	Code int
}

// Checker interface used to get latest status from plugin
type Checker interface {
	Status() Status
}
