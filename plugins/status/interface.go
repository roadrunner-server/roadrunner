package status

// Status consists of status code from the service
type Status struct {
	Code int
}

// Checker interface used to get latest status from plugin
type Checker interface {
	Status() Status
}

// Readiness interface used to get readiness status from the plugin
// that means, that worker poll inside the plugin has 1+ plugins which are ready to work
// at the particular moment
type Readiness interface {
	Ready() Status
}
