package services

import (
	"bytes"
	"testing"
)

func TestRoast(t *testing.T) {
	expected := []byte{0x83, 0x47, 0xF2, 0xB7, 0x4E, 0xE9, 0xA9, 0xF6}
	result := roast("password")
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %+v, but got %+v", expected, result)
	}
}
