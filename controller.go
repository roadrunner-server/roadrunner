package roadrunner

// Controller observes pool state and decides if any worker must be destroyed.
type Controller interface {
	// Lock controller on given pool instance.
	Attach(p Pool) Controller

	// Detach pool watching.
	Detach()
}

// Attacher defines the ability to attach rr controller.
type Attacher interface {
	// Attach attaches controller to the service.
	Attach(c Controller)
}
