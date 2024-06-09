package modbus

import (
	"bytes"
	"encoding/binary"
)

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
