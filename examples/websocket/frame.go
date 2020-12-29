package main

// A frame is a byte array that represents a WebSocket frame as defined in
// [RFC 6455](https://tools.ietf.org/html/rfc6455).
//
//      0                   1                   2                   3
//      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//     +-+-+-+-+-------+-+-------------+-------------------------------+
//     |F|R|R|R| opcode|M| Payload len |    Extended payload length    |
//     |I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
//     |N|V|V|V|       |S|             |   (if payload len==126/127)   |
//     | |1|2|3|       |K|             |                               |
//     +-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
//     |     Extended payload length continued, if payload len == 127  |
//     + - - - - - - - - - - - - - - - +-------------------------------+
//     |                               |Masking-key, if MASK set to 1  |
//     +-------------------------------+-------------------------------+
//     | Masking-key (continued)       |          Payload Data         |
//     +-------------------------------- - - - - - - - - - - - - - - - +
//     :                     Payload Data continued ...                :
//     + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
//     |                     Payload Data continued ...                |
//     +---------------------------------------------------------------+
type frame []byte

const (
	opClose = 0x08
	opText  = 0x01
	opPing  = 0x09
	opPong  = 0x0A
)

func (f frame) size() int {
	// First byte is flags and opcode. Second is mask indicator and payload
	// length.
	masked := (0x80&f[1] != 0)
	var plen uint64
	plen = uint64(f[1] & 0x7f)
	switch plen {
	case 126:
		plen = uint64(f[2])<<8 | uint64(f[3])
		plen += 4
	case 127:
		plen = 0
		for i := 8; 0 < i; i-- {
			plen = (plen << 8) | uint64(f[2+i])
		}
		plen += 10
	default:
		plen += 2
	}
	if masked {
		plen += 4
	}
	return int(plen)
}

func (f frame) op() int {
	return int(0x0F & f[0])
}

func (f frame) ready() bool {
	return 2 < len(f) && f.size() <= len(f)
}

func (f frame) payload() []byte {
	masked := (0x80&f[1] != 0)
	var plen uint64
	plen = uint64(f[1] & 0x7f)
	off := 0
	switch plen {
	case 126:
		plen = uint64(f[2])<<8 | uint64(f[3])
		off = 4
	case 127:
		plen = 0
		for i := 8; 0 < i; i-- {
			plen = (plen << 8) | uint64(f[2+i])
		}
		off = 10
	default:
		off = 2
	}
	if !masked {
		return f[off : int(plen)+off]
	}
	// masked
	off += 4
	mask := make([]byte, 4)
	for i, b := range f[off-4 : off] {
		mask[i] = b
	}
	payload := make([]byte, 0, plen)
	for i, b := range f[off : int(plen)+off] {
		payload = append(payload, b^mask[i%4])
	}
	return payload
}

func newFrame(payload []byte) frame {
	f := make(frame, 0, len(payload)+10)
	f = append(f, byte(0x80|opText)) // FIN and text opcode
	plen := uint64(len(payload))
	switch {
	case plen < 126:
		f = append(f, byte(plen))
	case plen <= 0xFFFF:
		f = append(f, byte(0x7E))
		f = append(f, byte((plen>>8)&0xFF))
		f = append(f, byte((plen & 0xFF)))
	default:
		f = append(f, byte(0x7F))
		for i := 56; 0 <= i; i -= 8 {
			f = append(f, byte((plen>>i)&0xFF))
		}
	}
	f = append(f, payload...)

	return f
}
