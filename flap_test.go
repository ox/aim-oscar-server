package main

import (
	"reflect"
	"testing"
)

func TestFLAP(t *testing.T) {
	session := Session{}
	hello := NewFLAP(&session, 1)
	hello.Data.Write([]byte{0, 0, 0, 1})

	b, err := hello.MarshalBinary()
	if err != nil {
		t.Errorf("FLAP failed to marshal: %x", err)
	}

	expected := []byte{0x2a, 1, 0, 1, 0, 4, 0, 0, 0, 1}
	if !reflect.DeepEqual(b, expected) {
		t.Errorf("unexpected marshaled bytes. expected: %v, got: %v", expected, b)
	}
}

func TestUnmarshalFLAP(t *testing.T) {
	b := []byte{0x2a, 1, 0, 1, 0, 4, 0, 0, 0, 1}

	hello := FLAP{}
	if err := hello.UnmarshalBinary(b); err != nil {
		t.Errorf("FLAP failed to unmarshal from bytes: %x", err)
	}

	if hello.Header.Channel != 1 {
		t.Errorf("FLAP channel should be %d, got %d", 1, hello.Header.Channel)
	}
	if hello.Header.SequenceNumber != 1 {
		t.Errorf("FLAP sequence number should be %d, got %d", 1, hello.Header.SequenceNumber)
	}
	if hello.Header.DataLength != 4 {
		t.Errorf("FLAP data length should be %d, got %d", 4, hello.Header.DataLength)
	}
	if !reflect.DeepEqual(hello.Data.Bytes(), b[6:]) {
		t.Errorf("FLAP body should be %x, got %x", b[6:], hello.Data.Bytes())
	}
}
