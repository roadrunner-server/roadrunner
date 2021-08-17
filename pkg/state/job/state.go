package job

// State represents job's state
type State struct {
	// Pipeline name
	Pipeline string
	// Driver name
	Driver string
	// Queue name (tube for the beanstalk)
	Queue string
	// Active jobs which are consumed from the driver but not handled by the PHP worker yet
	Active int64
	// Delayed jobs
	Delayed int64
	// Reserved jobs which are in the driver but not consumed yet
	Reserved int64
}
