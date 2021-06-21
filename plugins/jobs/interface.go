package jobs


// todo naming
type Consumer interface {
	Push()
	Stat()
}
