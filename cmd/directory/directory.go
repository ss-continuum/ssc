package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pkg/errors"
	"github.com/ss-continuum/ssc/pkg/bytestream"
	"github.com/ss-continuum/ssc/pkg/packetmap"
)

const directoryServerPort = 4990

var endian = binary.LittleEndian

func requestDirectoryList(addr string, debug bool) ([]DirectoryEntry, error) {
	var list []DirectoryEntry

	conn, err := connect(addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	conn.Debug = debug

	packets := packetmap.New()
	packets0x0a := packetmap.New()

	if err := conn.Login(0); err != nil {
		log.Fatalln(err)
	}

	timeoutCount := 0
	for {
		data, err := conn.ReadWithDeadline(time.Second * 5)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, errors.Wrap(err, "conn.ReadWithDeadline")
			}
			log.Println(err)
			timeoutCount++

			if timeoutCount >= 5 {
				return nil, errors.New("timeout x 5")
			}

			continue
		}

		if data[0] == 0x00 && data[1] == 0x02 {
			if err := conn.DirectoryListRequest(0); err != nil {
				log.Println("DirectoryListRequest:", err)
			}
		}
		if data[0] == 0x00 && data[1] == 0x03 {
			packetID := endian.Uint32(data[2:6])
			packetData := data[6:]
			if packetData[0] == 0x00 && (packetData[1] == 0x08 || packetData[1] == 0x09) {
				packets.Add(packetID, packetData)
			} else if packetData[0] == 0x00 && packetData[1] == 0x0a {
				packets0x0a.Add(packetID, packetData)
			} else {
				log.Printf("I don't know what to do with packet 0x%02x 0x%02x inside a 0x00 0x03 packet\n", packetData[0], packetData[1])
				continue
			}

			if err := conn.Ack(packetID); err != nil {
				log.Println("cannot ack packet", packetID)
				continue
			}

			if packetData[0] == 0x00 && packetData[1] == 0x09 {
				consolidatedData := packets.Bytes()
				packets.Clear()

				entryList, err := NewDirectoryEntryList(bytestream.New(consolidatedData, endian))
				if err != nil {
					log.Println("decodeDirectoryPayload:", err)
				}

				list = entryList

				if err := conn.Disconnect(); err != nil {
					return nil, errors.Wrap(err, "conn.Disconnect")
				}
				break
			} else if packetData[0] == 0x00 && packetData[1] == 0x0a {
				expectedLen := int(endian.Uint32(packetData[2:6]))
				if packets0x0a.Size() >= expectedLen {
					consolidatedData := packets0x0a.Bytes()
					packets0x0a.Clear()

					entryList, err := NewDirectoryEntryList(bytestream.New(consolidatedData, endian))
					if err != nil {
						log.Println("decodeDirectoryPayload:", err)
					}

					list = entryList

					if err := conn.Disconnect(); err != nil {
						return nil, errors.Wrap(err, "conn.Disconnect")
					}
					break
				}
			}
		}
		if data[0] == 0x00 && data[1] == 0x0e {
			log.Println("DirectoryListResponse")
			stream := bytestream.New(data[2:], endian)

			// unknown 7-byte sequence -- always 0x06 0x00 0x04 0x00 0x00 0x00 0x00
			if _, err := stream.Seek(7, io.SeekCurrent); err != nil {
				log.Println("Seek+7:", err)
			}

			// 1-byte packet size
			pktSize, err := stream.ReadByte()
			if err != nil {
				log.Println("pktSize:", err)
			}

			pktType, err := stream.ReadUint16()
			if err != nil {
				log.Println("pktType:", err)
			}

			pktID, err := stream.ReadUint32()
			if err != nil {
				log.Println("pktID:", err)
			}

			fmt.Println(pktSize, pktType, pktID)

			// another 2-byte unknown
			if _, err := stream.Seek(2, io.SeekCurrent); err != nil {
				log.Println("Seek+2:", err)
			}
		}
		if data[0] == 0x00 && data[1] == 0x04 {
			id := endian.Uint32(data[2:6])

			if id == 0 {
				// send disconnect
				err := conn.Disconnect()
				if err != nil {
					log.Println("Disconnect:", err)
				}
			}
		}

		if data[0] == 0x00 && data[1] == 0x07 {
			log.Println("Disconnected requested from server")
			break
		}
	}

	return list, nil
}

func main() {
	fs := flag.NewFlagSet("ssc-directory", flag.ExitOnError)

	var Port int
	var Debug bool

	fs.IntVar(&Port, "port", directoryServerPort, "server port")
	fs.BoolVar(&Debug, "debug", false, "log network packets")

	root := &ffcli.Command{
		ShortUsage: fmt.Sprintf("%s [-debug] [-port <portnumber>] address", os.Args[0]),
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.Errorf("Unexpected number of args. Expected: 1, got: %d", len(args))
			}

			addr := fmt.Sprintf("%s:%d", args[0], Port)

			log.Printf("Requesting directory at %s\n", addr)
			list, err := requestDirectoryList(addr, Debug)
			if err != nil {
				return errors.Wrap(err, "error requesting list")
			}

			for _, entry := range list {
				fmt.Println("---")
				fmt.Println(entry)
			}
			fmt.Println("---")

			return nil
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
