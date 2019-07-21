package resource

import (
	"encoding/binary"
	"fmt"
	"github.com/32bitkid/bitreader"
	"io"
)

// Huffman decoding

type huffmanNodes struct {
	Value    uint8
	Siblings uint8
}

type huffmanState struct {
	nodes []huffmanNodes
	br    bitreader.BitReader8
}

func (h *huffmanState) next(idx int) (uint8, bool, error) {
	node := h.nodes[idx]
	value := node.Value
	siblings := node.Siblings

	if siblings == 0 {
		return value, false, nil
	}

	bit, err := h.br.Read1()
	if err != nil {
		return 0, false, err
	}

	var next int
	if bit {
		next = int(siblings & 0x0f)
	} else {
		next = int(siblings & 0xf0 >> 4)
	}

	if next == 0 {
		literal, err := h.br.Read8(8)
		return literal, true, err
	}

	return h.next(idx + next)
}

func huffman(src io.Reader, dest []uint8) error {
	var nodeCount uint8
	if err := binary.Read(src, binary.LittleEndian, &nodeCount); err != nil {
		return err
	}

	var term uint8
	if err := binary.Read(src, binary.LittleEndian, &term); err != nil {
		return err
	}

	nodes := make([]huffmanNodes, nodeCount)
	if err := binary.Read(src, binary.LittleEndian, &nodes); err != nil {
		return err
	}

	huffman := huffmanState{
		br:    bitreader.NewReader(src),
		nodes: nodes,
	}

	i := 0
	for {
		c, ok, err := huffman.next(0)
		if err != nil {
			return err
		}
		if ok && c == term {
			break
		}
		dest[i] = c
		i++
	}

	if i != len(dest) {
		return fmt.Errorf("read aborted early. expected(%d) != actual(%d)", len(dest), i)
	}

	return nil
}
