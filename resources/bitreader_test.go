package resources

import (
	"bytes"
	"testing"
)

func TestBitreader(t *testing.T) {
	br := newBitReader(bytes.NewReader([]byte{0x80}))
	if b, err := br.read1(); err != nil {
		t.Fatal(err)
	} else if b != true {
		t.Fatal("expected true")
	}
	_, err := br.read8(8)
	if err == nil {
		t.Fatal(err)
	}
}

type brTestCase struct {
	len   uint
	value uint64
}

func runBrTest(t *testing.T, data []byte, cases []brTestCase) {
	br := newBitReader(bytes.NewReader(data))

	for i, tCase := range cases {
		val, err := br.read(tCase.len)
		if err != nil {
			t.Fatal(err)
		}
		if tCase.value != val {
			t.Fatalf("%d: expected(%d) != actual(%d)", i, tCase.value, val)
		}
	}
}

func TestBitreader2(t *testing.T) {
	runBrTest(
		t,
		[]byte{
			0x80, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x01,
		},
		[]brTestCase{
			{1, 1},
			{3, 0},
			{4, 0},
			{8, 255},
			{64, 256},
			{4, 0},
			{4, 1},
		},
	)
}

func TestBitreader3(t *testing.T) {
	runBrTest(
		t,
		[]byte{
			0x80, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x01, 0x00, 0x00, 0x1f, 0x2f, 0x3d,
		},
		[]brTestCase{
			{32, 0x80ff0000},
			{64, 0x0000000001000100},
			{32, 0x001f2f3d},
		},
	)
}
