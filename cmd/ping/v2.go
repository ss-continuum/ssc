package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/ss-continuum/ssc/pkg/bytestream"
	"github.com/ss-continuum/ssc/pkg/logbytes"
)

const PingGlobalSummary = 0x01
const PingArenaSummary = 0x02

type PingV2Resp struct {
	ClientTime uint32 // This is but an echo of what was sent to the server
	Options    uint32

	Lag uint32 // in milliseconds

	GlobalSummary *PingV2GlobalSummary
	ArenaSummary  []PingV2ArenaSummary
}

func (p PingV2Resp) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("Lag: %dms", p.Lag))
	if p.GlobalSummary != nil {
		lines = append(lines, fmt.Sprintf("GlobalSummary: %s", p.GlobalSummary))
	}

	for i, summary := range p.ArenaSummary {
		lines = append(lines, fmt.Sprintf("Arena %d: %s", i, summary))
	}

	return strings.Join(lines, "\n")
}

type PingV2GlobalSummary struct {
	Total   uint32
	Playing uint32
}

func (g PingV2GlobalSummary) String() string {
	return fmt.Sprintf("Total: %d, Playing: %d", g.Total, g.Playing)
}

type PingV2ArenaSummary struct {
	Name    string
	Total   uint16
	Playing uint16
}

func (a PingV2ArenaSummary) String() string {
	return fmt.Sprintf("%s: %d/%d", a.Name, a.Playing, a.Total)
}

var endian = binary.LittleEndian

func PingV2(ip string, port int, debug bool, options uint32) (PingV2Resp, error) {
	var resp PingV2Resp
	addr := fmt.Sprintf("%s:%d", ip, port)

	conn, err := net.Dial("udp", addr)
	if err != nil {
		return resp, errors.Wrap(err, "net.Dial")
	}
	defer conn.Close()

	then := uint32(time.Now().UnixMilli())

	C2SSimplePingV2 := make([]byte, 8)
	endian.PutUint32(C2SSimplePingV2[0:4], then)
	endian.PutUint32(C2SSimplePingV2[4:8], options)

	if debug {
		logbytes.LogPrefix(C2SSimplePingV2, "C2S |")
	}
	if _, err := conn.Write(C2SSimplePingV2); err != nil {
		return resp, errors.Wrap(err, "conn.Write")
	}

	respBytes := make([]byte, 2048)
	n, err := conn.Read(respBytes)
	if err != nil {
		return resp, errors.Wrap(err, "conn.Read")
	}
	if debug {
		logbytes.LogPrefix(respBytes[:n], "S2C |")
	}

	in := bytestream.New(respBytes[:n], endian)

	if err := in.ReadUint32Var(&resp.ClientTime); err != nil {
		return resp, errors.Wrap(err, "in.ReadUint32Var")
	}
	if err := in.ReadUint32Var(&resp.Options); err != nil {
		return resp, errors.Wrap(err, "in.ReadUint32Var")
	}

	now := uint32(time.Now().UnixMilli())
	resp.Lag = now - resp.ClientTime

	// Global summary
	if options&PingGlobalSummary != 0 {
		var g PingV2GlobalSummary
		if err := in.ReadUint32Var(&g.Total); err != nil {
			return resp, errors.Wrap(err, "in.ReadUint32Var")
		}
		if err := in.ReadUint32Var(&g.Playing); err != nil {
			return resp, errors.Wrap(err, "in.ReadUint32Var")
		}

		resp.GlobalSummary = &g
	}

	// Arena summary
	if options&PingArenaSummary != 0 {
		for in.Len() > 0 {
			var arena PingV2ArenaSummary
			name, err := in.ReadZeroString()
			if err != nil {
				return resp, errors.Wrap(err, "ReadZeroString")
			}
			if name == "" {
				break
			}
			arena.Name = name
			if err := in.ReadUint16Var(&arena.Total); err != nil {
				return resp, errors.Wrap(err, "ReadUint16Var")
			}
			if err := in.ReadUint16Var(&arena.Playing); err != nil {
				return resp, errors.Wrap(err, "ReadUint16Var")
			}

			resp.ArenaSummary = append(resp.ArenaSummary, arena)
		}
	}

	return resp, nil
}
