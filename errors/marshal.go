package errors

import (
	"encoding/binary"
	"log"
)

func (e *Error) MarshalAppend(b []byte) []byte {
	if e == nil {
		return b
	}

	b = appendString(b, string(e.Op))

	var tmp [16]byte
	N := binary.PutVarint(tmp[:], int64(e.Kind))
	b = append(b, tmp[:N]...)
	b = MarshalErrorAppend(e.Err, b)
	return b
}

func (e *Error) MarshalBinary() ([]byte, error) {
	return e.MarshalAppend(nil), nil
}

func MarshalErrorAppend(err error, b []byte) []byte {
	if err == nil {
		return b
	}
	if e, ok := err.(*Error); ok {
		b = append(b, 'E')
		return e.MarshalAppend(b)
	}
	// Ordinary error.
	b = append(b, 'e')
	b = appendString(b, err.Error())
	return b
}

func MarshalError(err error) []byte {
	return MarshalErrorAppend(err, nil)
}

func (e *Error) UnmarshalBinary(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	data, b := getBytes(b)
	if data != nil {
		e.Op = Op(data)
	}
	k, N := binary.Varint(b)
	e.Kind = Kind(k)
	b = b[N:]
	e.Err = UnmarshalError(b)
	return nil
}

func UnmarshalError(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	code := b[0]
	b = b[1:]
	switch code {
	case 'e':
		var data []byte
		data, b = getBytes(b)
		if len(b) != 0 {
			log.Printf("Unmarshal error: trailing bytes")
		}
		return Str(string(data))
	case 'E':
		var err Error
		err.UnmarshalBinary(b)
		return &err
	default:
		log.Printf("Unmarshal error: corrupt data %q", b)
		return Str(string(b))
	}
}

func appendString(b []byte, str string) []byte {
	var tmp [16]byte
	N := binary.PutUvarint(tmp[:], uint64(len(str)))
	b = append(b, tmp[:N]...)
	b = append(b, str...)
	return b
}

func getBytes(b []byte) (data, remaining []byte) {
	u, N := binary.Uvarint(b)
	if len(b) < N+int(u) {
		log.Printf("Unmarshal error: bad encoding")
		return nil, nil
	}
	if N == 0 {
		log.Printf("Unmarshal error: bad encoding")
		return nil, b
	}
	return b[N : N+int(u)], b[N+int(u):]
}
