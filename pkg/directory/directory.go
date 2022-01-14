package directory

import (
	"github.com/pkg/errors"
	"github.com/ss-continuum/ssc/pkg/bytestream"
	"log"
)

type Directory struct {
	Entries []Entry
}

func NewFromStream(stream *bytestream.ByteStream) (Directory, error) {
	if h, err := stream.ReadByte(); err != nil {
		return Directory{}, errors.Wrap(err, "failed to read header")
	} else if h != 0x01 {
		log.Printf("unexpected header: 0x%02x\n", h)
	}

	var list []Entry

	for stream.Len() > 0 {
		entry, err := NewEntry(stream)
		if err != nil {
			return Directory{}, errors.Wrap(err, "NewEntry")
		}

		list = append(list, entry)
	}

	return Directory{Entries: list}, nil
}
