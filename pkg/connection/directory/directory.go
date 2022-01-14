package directory

import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/ss-continuum/ssc/pkg/bytestream"
	"github.com/ss-continuum/ssc/pkg/connection/server"
	"github.com/ss-continuum/ssc/pkg/directory"
	"github.com/ss-continuum/ssc/pkg/packetmap"
	"io"
	"log"
	"time"
)

var endian = binary.LittleEndian

type Connection struct {
	*server.Connection

	x08Map packetmap.PacketMap
	x0aMap packetmap.PacketMap

	x08ChunkComplete bool
	x0aChunkComplete bool
}

func Dial(addr string) (*Connection, error) {
	conn, err := server.Dial(addr)
	if err != nil {
		return nil, err
	}
	return &Connection{
		Connection: conn,
	}, nil
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

func (s *Connection) handleSmallChunk(id uint32, chunk []byte) {
	s.x08Map.Add(id, chunk)

	if chunk[0] == 0x00 && chunk[1] == 0x09 {
		s.x08ChunkComplete = true
	}
}

func (s *Connection) handleBigChunk(id uint32, chunk []byte) {
	s.x0aMap.Add(id, chunk)

	expectedLen := int(endian.Uint32(chunk[2:6]))

	if s.x0aMap.Size() >= expectedLen {
		s.x0aChunkComplete = true
	}
}

func (s *Connection) handle0x03(data []byte) error {
	packetID := endian.Uint32(data[2:6])
	packetData := data[6:]

	if err := s.Ack(packetID); err != nil {
		return errors.Errorf("cannot ack packet %d", packetID)
	}

	if packetData[0] == 0x00 && (packetData[1] == 0x08 || packetData[1] == 0x09) {
		s.handleSmallChunk(packetID, packetData)
	} else if packetData[0] == 0x00 && packetData[1] == 0x0a {
		s.handleBigChunk(packetID, packetData)
	} else {
		return errors.Errorf("I don't know what to do with packet 0x%02x 0x%02x inside a 0x00 0x03 packet\n", packetData[0], packetData[1])
	}

	return nil
}

func (s *Connection) RequestList(minPlayers uint32) error {
	payload := []byte{
		0x00, 0x03,
		0, 0, 0, 0, 0x01,
		0, 0, 0, 0,
	}
	endian.PutUint32(payload[7:11], minPlayers)

	if _, err := s.Write(payload); err != nil {
		return errors.Wrap(err, "s.Write")
	}

	return nil
}

func (s *Connection) Directory(minPlayers uint32) (directory.Directory, error) {
	s.x08Map = packetmap.New()
	s.x0aMap = packetmap.New()

	var ret directory.Directory
	timeoutCount := 0

	for {
		data, err := s.ReadWithDeadline(5 * time.Second)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return ret, errors.Wrap(err, "conn.ReadWithDeadline")
			}
			log.Println(err)
			timeoutCount++

			if timeoutCount >= 5 {
				return ret, errors.New("timeout x 5")
			}

			continue
		}

		if data[0] == 0x00 && data[1] == 0x02 {
			if err := s.RequestList(minPlayers); err != nil {
				return directory.Directory{}, errors.Wrap(err, "s.RequestList")
			}
		}

		if data[0] == 0x00 && data[1] == 0x03 {
			if err := s.handle0x03(data); err != nil {
				log.Println(err)
				//return ret, errors.Wrap(err, "failed to handle 0x03 packet")
			}
		}

		if s.x08ChunkComplete || s.x0aChunkComplete {
			_ = s.Disconnect()
			break
		}

		if data[0] == 0x00 && data[1] == 0x07 {
			return ret, errors.New("server requested disconnection")
		}
	}

	var consolidatedData []byte
	if s.x08ChunkComplete {
		consolidatedData = s.x08Map.Bytes()
		s.x08Map.Clear()
	} else if s.x0aChunkComplete {
		consolidatedData = s.x0aMap.Bytes()
		s.x0aMap.Clear()
	}

	entryList, err := directory.NewFromStream(bytestream.New(consolidatedData, endian))
	if err != nil {
		log.Println("decodeDirectoryPayload:", err)
	}

	return entryList, err
}
