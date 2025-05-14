package modbus

import (
	"bytes"
	"encoding/binary"
)

func boolbe(v bool) []byte {
	if v {
		return []byte{0xFF, 0x00}
	}
	return []byte{0x00, 0x00}
}

func boollistbe(values []bool) []byte {
	b := make([]byte, (len(values)+7)/8)
	for i, v := range values {
		if v {
			b[i/8] |= 1 << (7 - (i % 8))
		}
	}
	return b
}

func beboollist(b []byte) []bool {
	values := make([]bool, 0, len(b)*8)
	for _, v := range b {
		values = append(values, v&0x80 != 0, v&0x40 != 0, v&0x20 != 0, v&0x10 != 0,
			v&0x08 != 0, v&0x04 != 0, v&0x02 != 0, v&0x01 != 0)
	}
	return values
}

func u16be(v uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	return b
}

func u16listbe(values []uint16) []byte {
	b := make([]byte, 0, len(values)*2)
	for _, v := range values {
		binary.BigEndian.AppendUint16(b, v)
	}
	return b
}

func bytesJoin(b ...[]byte) []byte {
	return bytes.Join(b, nil)
}

func beu16list(b []byte) []uint16 {
	values := make([]uint16, 0, len(b)/2)
	for i := 0; i < len(b); i += 2 {
		v := binary.BigEndian.Uint16(b[i:])
		values = append(values, v)
	}
	return values
}
