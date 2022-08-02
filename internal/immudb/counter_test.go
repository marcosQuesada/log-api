package immudb

import (
	"fmt"
	"testing"
)

func TestItInitsBinaryLogLinesCounter(t *testing.T) {
	raw := initBinaryCounter()

	if expected, got := uint64(0), binaryCounter(raw); expected != got {
		t.Fatalf("Values do not match, expected %d got %d", expected, got)
	}
}

func TestItIncrementsBinaryLogLinesCounter(t *testing.T) {
	raw := initBinaryCounter()
	raw = incBinaryCounter(raw)

	if expected, got := uint64(1), binaryCounter(raw); expected != got {
		t.Fatalf("Values do not match, expected %d got %d", expected, got)
	}
}

func TestItCleansBinaryStringWithEmptyRunes(t *testing.T) {
	raw := []byte{0, 102, 111, 111, 95, 120}

	key := cleanKey(raw)
	if expected, got := len(raw)-1, len(key); expected != got {
		t.Fatalf("Values do not match, expected %d got %d", expected, got)
	}

	fmt.Println(key)
	firstRune := ""
	for i, _ := range key {
		if i == 0 {
			continue
		}
		firstRune = key[:i]
		break
	}

	if expected, got := "f", firstRune; expected != got {
		t.Fatalf("Values do not match, expected %s got %s", expected, got)
	}
}
