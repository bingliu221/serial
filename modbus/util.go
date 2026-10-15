package modbus

import (
	"encoding/binary"
)

func boolu16(v bool) uint16 {
	if v {
		return 0xFF00
	}
	return 0x0000
}

func putBoolList(dst []byte, values []bool) {
	for i, v := range values {
		if v {
			dst[i/8] |= 1 << (i % 8)
		}
	}
}

func beboollist(b []byte) []bool {
	values := make([]bool, 0, len(b)*8)
	for _, v := range b {
		values = append(values, v&0x01 != 0, v&0x02 != 0, v&0x04 != 0, v&0x08 != 0,
			v&0x10 != 0, v&0x20 != 0, v&0x40 != 0, v&0x80 != 0)
	}
	return values
}

func u16bePair(a, b uint16) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[0:2], a)
	binary.BigEndian.PutUint16(buf[2:4], b)
	return buf
}

func putU16ListBE(dst []byte, values []uint16) {
	for i, v := range values {
		binary.BigEndian.PutUint16(dst[i*2:], v)
	}
}

func beu16list(b []byte) []uint16 {
	values := make([]uint16, 0, len(b)/2)
	for i := 0; i < len(b); i += 2 {
		v := binary.BigEndian.Uint16(b[i:])
		values = append(values, v)
	}
	return values
}
