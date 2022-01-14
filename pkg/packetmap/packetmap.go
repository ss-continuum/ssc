package packetmap

import (
	"bytes"
	"fmt"
	"github.com/ss-continuum/ssc/pkg/logbytes"
	"sort"
)

// PacketMap is useful when handling a stream of packets that needs to be ordered and bundled together.
type PacketMap map[uint32][]byte

func New() PacketMap {
	return make(map[uint32][]byte)
}

func (p PacketMap) Add(id uint32, data []byte) {
	p[id] = data
}

func (p PacketMap) Clear() {
	for k := range p {
		delete(p, k)
	}
}

func (p PacketMap) Size() int {
	var size int
	for _, packet := range p {
		if packet[0] == 0x00 && packet[1] == 0x08 {
			// remove packet header (2 bytes)
			size += len(packet) - 2
		} else if packet[0] == 0x00 && packet[1] == 0x0a {
			// remove packet header (2 bytes) and size (4 bytes)
			size += len(packet) - 6
		}

	}
	return size
}

func (p PacketMap) Bytes() []byte {
	keys := make([]uint32, len(p))
	i := 0
	for k := range p {
		keys[i] = k
		i++
	}
	sortUint32(keys)

	out := bytes.NewBuffer([]byte{})
	for _, k := range keys {
		data := p[k]
		logbytes.LogPrefix(data, fmt.Sprintf("packet %d", k))
		if data[0] == 0x00 && (data[1] == 0x08 || data[1] == 0x09) {
			// remove packet header (2 bytes)
			out.Write(data[2:])
		} else if data[0] == 0x00 && data[1] == 0x0a {
			// remove packet header (2 bytes) and size (4 bytes)
			out.Write(data[6:])
		} else {
			out.Write(data)
		}
	}

	return out.Bytes()
}

func sortUint32(slice []uint32) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
}
