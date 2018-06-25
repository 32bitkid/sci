package pmachine

import (
	"testing"
)

func TestNotOp(t *testing.T) {
	sci := SCI{
		IP:  0x0000,
		Acc: 0xFF00,
	}

	sci.Heap[0] = 0x00
	sci.ExecuteStep()
	if sci.Acc != 255 {
		t.Error("unexpected")
	}
}

func TestAddOp(t *testing.T) {
	sci := SCI{
		IP:  0x0000,
		Acc: 0x0010,
	}

	sci.push(0x10)
	sci.Heap[0] = 0x02
	sci.ExecuteStep()
	if sci.Acc != 0x20 {
		t.Error("unexpected")
	}
}

func TestSubOp(t *testing.T) {
	sci := SCI{
		IP:  0x0000,
		Acc: 0x0008,
	}

	sci.Heap[0] = 0x04
	sci.push(0x10)
	sci.ExecuteStep()
	if sci.Acc != 0x08 {
		t.Errorf("unexpected %d", sci.Acc)
	}
}
