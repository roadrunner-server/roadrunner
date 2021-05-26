package pubsub

import (
	json "github.com/json-iterator/go"
)

type Msg struct {
	// Topic message been pushed into.
	T []string `json:"topic"`

	// Command (join, leave, headers)
	C string `json:"command"`

	// Broker (redis, memory)
	B string `json:"broker"`

	// Payload to be broadcasted
	P []byte `json:"payload"`
}

//func (m Msg) UnmarshalBinary(data []byte) error {
//	//Use default gob decoder
//	reader := bytes.NewReader(data)
//	dec := gob.NewDecoder(reader)
//	if err := dec.Decode(&m); err != nil {
//		return err
//	}
//
//	return nil
//}

func (m *Msg) MarshalBinary() ([]byte, error) {
	//buf := new(bytes.Buffer)
	//
	//for i := 0; i < len(m.T); i++ {
	//	buf.WriteString(m.T[i])
	//}
	//
	//buf.WriteString(m.C)
	//buf.WriteString(m.B)
	//buf.Write(m.P)

	return json.Marshal(m)

}

// Payload in raw bytes
func (m *Msg) Payload() []byte {
	return m.P
}

// Command for the connection
func (m *Msg) Command() string {
	return m.C
}

// Topics to subscribe
func (m *Msg) Topics() []string {
	return m.T
}

func (m *Msg) Broker() string {
	return m.B
}
