package immudb

import (
	"encoding/binary"
	"strings"
)

func initBinaryCounter() []byte {
	var sizeValue = make([]byte, 8)
	binary.BigEndian.PutUint64(sizeValue, 0)
	return sizeValue
}

func incBinaryCounter(raw []byte) []byte {
	var size = binary.BigEndian.Uint64(raw)
	size++
	var sizeValue = make([]byte, 8)
	binary.BigEndian.PutUint64(sizeValue, size)
	return sizeValue
}

func binaryCounter(raw []byte) uint64 {
	return binary.BigEndian.Uint64(raw)
}

// cleanKey cleans empty runes from byte array
func cleanKey(k []byte) string {
	return strings.Replace(string(k), "\x00", "", -1)
}
