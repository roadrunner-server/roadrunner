package grpc

import "google.golang.org/grpc/encoding"

type rawMessage []byte

const rm string = "rawMessage"

func (r rawMessage) Reset()       {}
func (rawMessage) ProtoMessage()  {}
func (rawMessage) String() string { return rm }

type codec struct{ base encoding.Codec }

// Marshal returns the wire format of v. rawMessages would be returned without encoding.
func (c *codec) Marshal(v interface{}) ([]byte, error) {
	if raw, ok := v.(rawMessage); ok {
		return raw, nil
	}

	return c.base.Marshal(v)
}

// Unmarshal parses the wire format into v. rawMessages would not be unmarshalled.
func (c *codec) Unmarshal(data []byte, v interface{}) error {
	if raw, ok := v.(*rawMessage); ok {
		*raw = data
		return nil
	}

	return c.base.Unmarshal(data, v)
}

// String return codec name.
func (c *codec) String() string {
	return "raw:" + c.base.Name()
}
