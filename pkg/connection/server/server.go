package server

import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/ss-continuum/ssc/pkg/logbytes"
	"log"
	"net"
	"time"
)

var endian = binary.LittleEndian

// Connection is a helper struct for handling udp connections to ssc ping, directory, billing and game servers.
type Connection struct {
	net.Conn
	Debug bool
}

// Dial -- connect to addr in the format ip:port
func Dial(addr string) (*Connection, error) {
	log.Printf("Connecting to %s...\n", addr)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "net.Dial")
	}

	log.Printf("Connected.\n")

	return &Connection{Conn: conn}, nil
}

func (s *Connection) Write(b []byte) (int, error) {
	if s.Debug {
		logbytes.LogPrefix(b, "C2S |")
	}
	return s.Conn.Write(b)
}

func (s *Connection) Login(key uint32) error {
	out := bytes.NewBuffer([]byte{})
	out.Write([]byte{0x00, 0x01})

	uint32Bytes := make([]byte, 4)
	// write key
	endian.PutUint32(uint32Bytes, key)
	out.Write(uint32Bytes)

	// protocol version
	out.Write([]byte{0x01, 0x00}) // vie
	//out.Write([]byte{0x11, 0x00}) // continuum

	_, err := s.Write(out.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to write login")
	}

	return nil
}

func (s *Connection) Ack(packetID uint32) error {
	out := bytes.NewBuffer([]byte{})
	out.Write([]byte{0x00, 0x04})

	uint32Bytes := make([]byte, 4)
	// write key
	endian.PutUint32(uint32Bytes, packetID)
	out.Write(uint32Bytes)

	_, err := s.Write(out.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to write ack")
	}

	return nil
}

func (s *Connection) Disconnect() error {
	payload := []byte{
		0x00, 0x07,
	}

	_, err := s.Write(payload)
	if err != nil {
		return errors.Wrap(err, "s.Write")
	}

	return nil
}

func (s *Connection) ReadWithDeadline(duration time.Duration) ([]byte, error) {
	err := s.SetReadDeadline(time.Now().Add(duration))
	if err != nil {
		return nil, errors.Wrap(err, "failed to set read deadline")
	}

	buf := make([]byte, 1024)
	n, err := s.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read")
	}

	data := buf[:n]
	if s.Debug {
		logbytes.LogPrefix(data, "S2C |")
	}

	return data, nil
}
