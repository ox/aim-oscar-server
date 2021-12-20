package util

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestLPString(t *testing.T) {
	expected := []byte{4, 0x74, 0x6f, 0x6f, 0x66}
	result := LPString("toof")
	if !bytes.Equal(expected, result) {
		t.Errorf("expected bytes to look like %+v, got %+v", expected, result)
	}
}

func TestLPUin16String(t *testing.T) {
	expected := []byte{0, 4, 0x74, 0x6f, 0x6f, 0x66}
	result := LPUint16String("toof")
	if !bytes.Equal(expected, result) {
		t.Errorf("expected bytes to look like %+v, got %+v", expected, result)
	}
}

func TestLPUStringLong(t *testing.T) {
	str := `<HTML><BODY BGCOLOR="#ffffff"><FONT>TEST of profile</FONT></BODY></HTML>`
	result := LPString(str)

	resultLength := uint8(result[0])
	if int(resultLength) != len(str) {
		t.Errorf("expected length prefix to be %x but got %x", len(str), resultLength)
	}
}

func TestLPUint16StringLong(t *testing.T) {
	str := `<HTML><BODY BGCOLOR="#ffffff"><FONT>TEST of profile</FONT></BODY></HTML>`
	result := LPUint16String(str)

	resultLength := binary.BigEndian.Uint16(result[:2])
	if int(resultLength) != len(str) {
		t.Errorf("expected length prefix to be %x but got %x", len(str), resultLength)
	}
}
