package pmachine

import "fmt"

type SCI struct {
	Heap [0xFFFF]uint8

	Acc uint16
	IP  uint16
	SP  uint16
}

func (sci *SCI) pop() uint16 {
	sci.SP -= 2
	base := 0xFFFF - sci.SP
	val := uint16(sci.Heap[base-1]<<8) | uint16(sci.Heap[base-2])
	return val
}

func (sci *SCI) push(v uint16) {
	base := 0xFFFF - sci.SP
	sci.Heap[base-1] = uint8(v >> 8)
	sci.Heap[base-2] = uint8(v & 0xFF)
	sci.SP += 2
}

const (
	bFALSE = 0
	bTRUE  = 1
)

type op uint8

const (
	opNOT  op = iota
	opADD
	opSUB
	opMUL
	opDIV
	opMOD
	opSHR
	opSHL
	opXOR
	opAND
	opOR
	opNEG
	opBNOT
	opEQ
	opNE
	opGT
	opGTE
	opLT
	opLTE
)

var opHandlers = [128]func(*SCI, bool){
	opNOT: func(sci *SCI, _ bool) { sci.Acc = ^sci.Acc },
	opADD: func(sci *SCI, _ bool) { sci.Acc = sci.pop() + sci.Acc },
	opSUB: func(sci *SCI, _ bool) { sci.Acc = sci.pop() - sci.Acc },
	opMUL: func(sci *SCI, _ bool) { sci.Acc = sci.pop() * sci.Acc },
	opDIV: func(sci *SCI, _ bool) {
		if sci.Acc != 0 {
			sci.Acc = sci.pop() / sci.Acc
		}
	},
	opMOD: func(sci *SCI, _ bool) { sci.Acc = sci.pop() % sci.Acc },
	opSHR: func(sci *SCI, _ bool) { sci.Acc = sci.pop() >> sci.Acc },
	opSHL: func(sci *SCI, _ bool) { sci.Acc = sci.pop() << sci.Acc },
	opXOR: func(sci *SCI, _ bool) { sci.Acc = sci.Acc ^ sci.pop() },
	opAND: func(sci *SCI, _ bool) { sci.Acc = sci.Acc & sci.pop() },
	opOR:  func(sci *SCI, _ bool) { sci.Acc = sci.Acc | sci.pop() },
	opNEG: func(sci *SCI, _ bool) { sci.Acc = -sci.Acc },
	opBNOT: func(sci *SCI, _ bool) {
		if sci.Acc == bFALSE {
			sci.Acc = bTRUE
		} else {
			sci.Acc = bFALSE
		}
	},
	opEQ: func(sci *SCI, _ bool) {
		if sci.Acc == sci.pop() {
			sci.Acc = bTRUE
		} else {
			sci.Acc = bFALSE
		}
	},
	opNE: func(sci *SCI, _ bool) {
		if sci.Acc != sci.pop() {
			sci.Acc = bTRUE
		} else {
			sci.Acc = bFALSE
		}
	},
	opGT: func(sci *SCI, _ bool) {
		if sci.pop() > sci.Acc {
			sci.Acc = bTRUE
		} else {
			sci.Acc = bFALSE
		}
	},
	opGTE: func(sci *SCI, _ bool) {
		if sci.pop() >= sci.Acc {
			sci.Acc = bTRUE
		} else {
			sci.Acc = bFALSE
		}
	},
	opLT: func(sci *SCI, _ bool) {
		if sci.pop() < sci.Acc {
			sci.Acc = bTRUE
		} else {
			sci.Acc = bFALSE
		}
	},
	opLTE: func(sci *SCI, _ bool) {
		if sci.pop() <= sci.Acc {
			sci.Acc = bTRUE
		} else {
			sci.Acc = bFALSE
		}
	},
}

func (sci *SCI) ExecuteStep() {
	code := sci.Heap[sci.IP]
	sci.IP += 1
	op := op(code >> 1)
	if fn := opHandlers[op]; fn != nil {
		size := code & 0x1 == 0x1
		fn(sci, size)
	} else {
		panic(fmt.Sprintf("op %x not supported", code))
	}
}