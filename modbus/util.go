package modbus

import (
	"encoding/binary"
	"fmt"
)

func encodeBool(v bool) uint16 {
	if v {
		return 0xFF00
	}
	return 0x0000
}

func encodeBools(values []bool) []byte {
	buf := make([]byte, (len(values)+7)/8)
	for i, v := range values {
		if v {
			buf[i/8] |= 1 << (i % 8)
		}
	}
	return buf
}

func decodeBools(b []byte, count uint16) ([]bool, error) {
	if len(b)*8 < int(count) {
		return nil, fmt.Errorf("expected %d bits, got %d", count, len(b)*8)
	}
	values := make([]bool, int(count))
	for i := range values {
		values[i] = b[i/8]&(1<<(i%8)) != 0
	}
	return values, nil
}

func encodeUint16s(values ...uint16) []byte {
	buf := make([]byte, len(values)*2)
	for i, v := range values {
		binary.BigEndian.PutUint16(buf[i*2:], v)
	}
	return buf
}

func decodeUint16s(b []byte, count uint16) ([]uint16, error) {
	if len(b) != int(count)*2 {
		return nil, fmt.Errorf("expected %d bytes, got %d", int(count)*2, len(b))
	}
	values := make([]uint16, int(count))
	for i := range values {
		values[i] = binary.BigEndian.Uint16(b[i*2:])
	}
	return values, nil
}

func encodeWithPayload(payload []byte, header ...uint16) []byte {
	buf := make([]byte, len(header)*2+1+len(payload))
	for i, v := range header {
		binary.BigEndian.PutUint16(buf[i*2:], v)
	}
	buf[len(header)*2] = byte(len(payload))
	copy(buf[len(header)*2+1:], payload)
	return buf
}
