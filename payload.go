package roadrunner

type Payload struct {
	Head, Body []byte
}

func (p *Payload) HeadString() {

}

// String returns payload body as string
func (p *Payload) String() string {
	return string(p.Body)
}
