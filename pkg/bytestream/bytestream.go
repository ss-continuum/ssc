package bytestream

import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
)

type ByteStream struct {
	*bytes.Reader
	binary.ByteOrder
}

func New(data []byte, endian binary.ByteOrder) *ByteStream {
	r := ByteStream{
		Reader:    bytes.NewReader(data),
		ByteOrder: endian,
	}

	if r.ByteOrder == nil {
		r.ByteOrder = binary.LittleEndian
	}

	return &r
}

func (in *ByteStream) ReadUint32() (uint32, error) {
	data := make([]byte, 4)
	if _, err := in.Read(data); err != nil {
		return 0, errors.Wrap(err, "in.Read")
	}

	return in.ByteOrder.Uint32(data), nil
}

func (in *ByteStream) ReadUint32Var(out *uint32) error {
	n, err := in.ReadUint32()
	if err != nil {
		return errors.Wrap(err, "ReadUint32")
	}

	*out = n

	return nil
}

func (in *ByteStream) ReadUint16() (uint16, error) {
	data := make([]byte, 2)
	if _, err := in.Read(data); err != nil {
		return 0, errors.Wrap(err, "in.Read")
	}

	return in.ByteOrder.Uint16(data), nil
}

func (in *ByteStream) ReadUint16Var(out *uint16) error {
	n, err := in.ReadUint16()
	if err != nil {
		return errors.Wrap(err, "ReadUint16")
	}

	*out = n

	return nil
}

// ReadZeroString reads a zero-terminated string from the given reader.
func (in *ByteStream) ReadZeroString() (string, error) {
	var ret string
	var c byte
	var err error

	for c, err = in.ReadByte(); err == nil && c != 0; c, err = in.ReadByte() {
		ret += string(c)
	}

	if err != nil {
		return "", errors.Wrap(err, "in.ReadByte")
	}

	return ret, nil
}
