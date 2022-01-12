package main

import (
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
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

	C2SSimplePingV1 := []byte{
		byte((then >> 24) & 0xff),
		byte((then >> 16) & 0xff),
		byte((then >> 8) & 0xff),
		byte(then & 0xff),
	}

	if debug {
		logbytes.LogPrefix(C2SSimplePingV1, "C2S |")
	}

	if _, err := conn.Write(C2SSimplePingV1); err != nil {
		return resp, errors.Wrap(err, "conn.Write")
	}

	respBytes := make([]byte, 8)
	if _, err := conn.Read(respBytes); err != nil {
		return resp, errors.Wrap(err, "conn.Read")
	}
	if debug {
		logbytes.LogPrefix(respBytes, "S2C |")
	}

	now := uint32(time.Now().UnixMilli())

	resp.PlayerCount = uint32(respBytes[0]) + uint32(respBytes[1])<<8 + uint32(respBytes[2])<<16 + uint32(respBytes[3])<<24
	resp.ClientTime = uint32(respBytes[4]) + uint32(respBytes[5])<<8 + uint32(respBytes[6])<<16 + uint32(respBytes[7])<<24

	resp.Lag = now - then

	return resp, nil
}
