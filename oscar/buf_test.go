package oscar

import "testing"

func fail(t *testing.T, e error, method string) {
	if e != nil {
		t.Errorf("invalid read from %s: %s", method, e.Error())
	}
}

func TestBuffer(t *testing.T) {
	b := Buffer{}

	b.WriteUint8(uint8(1))
	b.WriteUint16(uint16(2))
	b.WriteUint32(uint32(3))
	b.WriteUint64(uint64(4))

	x1, err := b.ReadUint8()
	fail(t, err, "ReadUint8")
	if x1 != 1 {
		t.Errorf("expected ReadUint8 to read 1, got %d", x1)
	}

	x2, err := b.ReadUint16()
	fail(t, err, "ReadUint16")
	if x2 != 2 {
		t.Errorf("expected ReadUint16 to read 2, got %d", x2)
	}

	x3, err := b.ReadUint32()
	fail(t, err, "ReadUint32")
	if x3 != 3 {
		t.Errorf("expected ReadUint32 to read 3, got %d", x3)
	}

	x4, err := b.ReadUint64()
	fail(t, err, "ReadUint64")
	if x4 != 4 {
		t.Errorf("expected ReadUint64 to read 4, got %d", x4)
	}
}

func TestBufferLPString(t *testing.T) {
	b := Buffer{}

	expectedStr := "This is a long string"
	b.WriteLPString(expectedStr)

	str, err := b.ReadLPString()
	fail(t, err, "ReadLPString")
	if str != expectedStr {
		t.Errorf("expected to read %s, got %s", expectedStr, str)
	}
}
