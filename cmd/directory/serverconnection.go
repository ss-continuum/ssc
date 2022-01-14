package main

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/ss-continuum/ssc/pkg/logbytes"
	"log"
	"net"
	"time"
)

type ServerConnection struct {
	net.Conn
	Debug bool
}

func (s *ServerConnection) Write(b []byte) (int, error) {
	if s.Debug {
		logbytes.LogPrefix(b, "C2S |")
	}
	return s.Conn.Write(b)
}

func (s *ServerConnection) Login(key uint32) error {
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

func (s *ServerConnection) Ack(packetID uint32) error {
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

func (s *ServerConnection) DirectoryHello(pack uint32) error {
	out := bytes.NewBuffer([]byte{})
	out.Write([]byte{0x00, 0x01})
	uint32Bytes := make([]byte, 4)
	endian.PutUint32(uint32Bytes, pack)
	out.Write(uint32Bytes)
	payload := out.Bytes()

	_, err := s.Write(payload)
	if err != nil {
		return errors.Wrap(err, "s.Write")
	}

	return nil
}

func (s *ServerConnection) DirectoryListRequest(minPlayers uint32) error {
	payload := []byte{
		0x00, 0x03,
		0, 0, 0, 0, 0x01,
		0, 0, 0, 0,
	}
	endian.PutUint32(payload[7:11], minPlayers)

	_, err := s.Write(payload)
	if err != nil {
		return errors.Wrap(err, "s.Write")
	}

	return nil
}

func (s *ServerConnection) Disconnect() error {
	payload := []byte{
		0x00, 0x07,
	}

	_, err := s.Write(payload)
	if err != nil {
		return errors.Wrap(err, "s.Write")
	}

	return nil
}

func (s *ServerConnection) ReadWithDeadline(duration time.Duration) ([]byte, error) {
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

// connect to addr in the format ip:port
func connect(addr string) (*ServerConnection, error) {
	log.Printf("Connecting to %s...\n", addr)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "net.Dial")
	}

	log.Printf("Connected.\n")

	return &ServerConnection{Conn: conn}, nil
}
