package decompression

import (
	"fmt"
	"github.com/32bitkid/bitreader"
	"io"
)

type lzwToken struct {
	data uint8
	next uint16
}

func lzw1(dst io.Writer, r io.Reader, max int) error {
	br := bitreader.NewReader(r)

	stack := make([]uint8, 0x1014)
	tokens := make([]lzwToken, 0x1014)

	const (
		DefaultEndToken     uint16 = 0x1ff
		DefaultCurrentToken uint16 = 0x102
		EndOfDataToken      uint16 = 0x101
		ResetToken          uint16 = 0x100
	)

	var (
		len int
		n int

		numBits      uint
		currentToken uint16
		endToken     uint16


		lastByte   uint8
		stackDepth uint16
		lastBits   uint16

		token uint16
		bits  uint16
		err   error
	)

reset:
	numBits = 9
	currentToken = DefaultCurrentToken
	endToken = DefaultEndToken

	bits, err = br.Read16(numBits)
	if err != nil {
		return err
	}
	if bits == EndOfDataToken {
		goto done
	}
	lastByte = uint8(bits & 0xff)
	n, err = dst.Write([]byte{lastByte})
	if err != nil {
		return err
	}

	len += n
	lastBits = bits

next:
	bits, err = br.Read16(numBits)
	if err != nil {
		return err
	}

	if bits == EndOfDataToken {
		goto done
	}

	if bits == ResetToken {
		goto reset
	}

	token = bits
	if token >= currentToken {
		token = lastBits
		stack[stackDepth] = lastByte
		stackDepth++
	}
	for (token > 0xff) && (token < 0x1004) {
		stack[stackDepth] = tokens[token].data
		stackDepth++
		token = tokens[token].next
	}

	lastByte = uint8(token & 0xff)
	stack[stackDepth] = lastByte
	stackDepth++

	for stackDepth > 0 {
		stackDepth--
		n, err := dst.Write([]byte{stack[stackDepth]})
		if err != nil {
			return err
		}
		len += n

		if max == len {
			goto done
		}
	}

	if currentToken <= endToken {
		tokens[currentToken].data = lastByte
		tokens[currentToken].next = lastBits
		currentToken++
		if currentToken == endToken && numBits < 12 {
			numBits++
			endToken = (endToken << 1) + 1
		}
	}
	lastBits = bits
	goto next

done:
	if len != max {
		return fmt.Errorf("decompression error: expected %d bytes got %d bytes", max, len)
	}

	return nil
}
