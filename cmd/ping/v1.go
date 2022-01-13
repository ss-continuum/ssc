package main

import (
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/ss-continuum/ssc/pkg/bytestream"
	"github.com/ss-continuum/ssc/pkg/logbytes"
)

type PingV1Resp struct {
	PlayerCount uint32
	ClientTime  uint32 // This is but an echo of what was sent to the server

	Lag uint32 // in milliseconds
}

func (p PingV1Resp) String() string {
	return fmt.Sprintf("PlayerCount: %d, Lag: %dms", p.PlayerCount, p.Lag)
}

func PingV1(ip string, port int, debug bool) (PingV1Resp, error) {
	var resp PingV1Resp
	addr := fmt.Sprintf("%s:%d", ip, port)

	conn, err := net.Dial("udp", addr)
	if err != nil {
		return resp, errors.Wrap(err, "net.Dial")
	}
	defer conn.Close()

	then := uint32(time.Now().UnixMilli())

	C2SSimplePingV1 := make([]byte, 4)
	endian.PutUint32(C2SSimplePingV1, then)

	if debug {
		logbytes.LogPrefix(C2SSimplePingV1, "C2S |")
	}

	if _, err := conn.Write(C2SSimplePingV1); err != nil {
		return resp, errors.Wrap(err, "conn.Write")
	}

	respBytes := make([]byte, 8)
	n, err := conn.Read(respBytes)
	if err != nil {
		return resp, errors.Wrap(err, "conn.Read")
	}
	if debug {
		logbytes.LogPrefix(respBytes, "S2C |")
	}
	in := bytestream.New(respBytes[:n], endian)

	now := uint32(time.Now().UnixMilli())

	if err := in.ReadUint32Var(&resp.PlayerCount); err != nil {
		return resp, errors.Wrap(err, "in.ReadUint32Var")
	}
	if err := in.ReadUint32Var(&resp.ClientTime); err != nil {
		return resp, errors.Wrap(err, "in.ReadUint32Var")
	}

	resp.Lag = now - then

	return resp, nil
}
