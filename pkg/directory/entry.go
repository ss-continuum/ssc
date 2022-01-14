package directory

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/ss-continuum/ssc/pkg/bytestream"
	"net"
	"strings"
)

type Entry struct {
	Name        string
	Description string
	IP          string
	Port        uint16

	ScoreKeeping uint16
	Players      uint16
	Version      uint32
}

func NewEntry(stream *bytestream.ByteStream) (Entry, error) {
	ipAddress := make([]byte, 4)
	if n, err := stream.Read(ipAddress); err != nil {
		return Entry{}, errors.Wrap(err, "stream.Read")
	} else if n != len(ipAddress) {
		return Entry{}, errors.Errorf("failed to read ipAddress: expected %d, got %d", len(ipAddress), n)
	}
	serverPort, err := stream.ReadUint16()
	if err != nil {
		return Entry{}, errors.Wrap(err, "stream.ReadUint16")
	}
	playerCount, err := stream.ReadUint16()
	if err != nil {
		return Entry{}, errors.Wrap(err, "stream.ReadUint16")
	}
	scoreKeeping, err := stream.ReadUint16()
	if err != nil {
		return Entry{}, errors.Wrap(err, "stream.ReadUint16")
	}
	serverVersion, err := stream.ReadUint32()
	if err != nil {
		return Entry{}, errors.Wrap(err, "stream.ReadUint32")
	}

	serverName := make([]byte, 64)
	currentOffset := stream.Size() - int64(stream.Len())
	if n, err := stream.Read(serverName); err != nil {
		return Entry{}, errors.Wrap(err, "stream.Read")
	} else if n != len(serverName) {
		return Entry{}, errors.Errorf("failed to read serverName: expected %d, got %d (offset %v)", len(serverName), n, currentOffset)
	}
	serverDescription, err := stream.ReadZeroString()
	if err != nil {
		return Entry{}, errors.Wrap(err, "stream.ReadZeroString")
	}

	return Entry{
		Name:         string(serverName),
		Description:  serverDescription,
		IP:           net.IP(ipAddress).String(),
		Port:         serverPort,
		ScoreKeeping: scoreKeeping,
		Players:      playerCount,
		Version:      serverVersion,
	}, nil
}

func (d Entry) String() string {
	pieces := []string{
		d.Name,
		fmt.Sprintf("ss://%s:%d", d.IP, d.Port),
		d.Description,
		fmt.Sprintf("%d players", d.Players),
		fmt.Sprintf("%d score keeping", d.ScoreKeeping),
		fmt.Sprintf("%d server version", d.Version),
	}

	return strings.Join(pieces, "\n")
}
