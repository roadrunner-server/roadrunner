package codec

import "google.golang.org/grpc/encoding"

type RawMessage []byte

const cName string = "proto"
const rm string = "rawMessage"

func (r RawMessage) Reset()       {}
func (RawMessage) ProtoMessage()  {}
func (RawMessage) String() string { return rm }

type Codec struct{ base encoding.Codec }

// Marshal returns the wire format of v. rawMessages would be returned without encoding.
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	if raw, ok := v.(RawMessage); ok {
		return raw, nil
	}

	return c.base.Marshal(v)
}

// Unmarshal parses the wire format into v. rawMessages would not be unmarshalled.
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if raw, ok := v.(*RawMessage); ok {
		*raw = data
		return nil
	}

	return c.base.Unmarshal(data, v)
}

func (c *Codec) Name() string {
	return cName
}

// String return codec name.
func (c *Codec) String() string {
	return "raw:" + c.base.Name()
}
