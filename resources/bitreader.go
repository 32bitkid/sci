package resources

import (
	"errors"
	"io"
)

type bitreader struct {
	r         io.Reader
	buffer    uint64
	remaining uint
	raw       [8]uint8
}

func (br *bitreader) fill() error {
	total := (64 - br.remaining) >> 3

	n, err := br.r.Read(br.raw[:total])
	if err != nil {
		return err
	}

	ir := br.remaining
	for i := 0; i < n; i++ {
		pos := 64 - 8 - (uint(i) << 3) - ir
		br.buffer |= uint64(br.raw[i]) << pos
		br.remaining += 8
	}

	return nil
}

func (br *bitreader) read1() (bool, error) {
	val, err := br.peek1()
	if err != nil {
		return false, nil
	}
	return val, br.skip(1)
}

func (br *bitreader) read8(n uint) (uint8, error) {
	if n > 8 {
		return 0, errors.New("overflow")
	}

	val, err := br.read(n)
	return uint8(val), err
}

func (br *bitreader) read16(n uint) (uint16, error) {
	if n > 16 {
		return 0, errors.New("overflow")
	}
	val, err := br.read(n)
	return uint16(val), err
}

func (br *bitreader) read32(n uint) (uint32, error) {
	if n > 32 {
		return 0, errors.New("overflow")
	}
	val, err := br.read(n)
	return uint32(val), err
}

func (br *bitreader) read(n uint) (uint64, error) {
	val, err := br.peek(n)
	if err != nil {
		return 0, err
	}
	return val, br.skip(n)
}

func (br *bitreader) peek1() (bool, error) {
	val, err := br.peek(1)
	return err == nil && val == 1, err
}

func (br *bitreader) peek(n uint) (uint64, error) {
	if n > 64 {
		return 0, errors.New("overflow")
	}

	if n > 56 && br.remaining&0x7 != 0 {
		return 0, errors.New("offset mismatch, can't fill the buffer with leftover-bytes")
	}

	for br.remaining < n {
		if err := br.fill(); err != nil {
			return 0, err
		}
	}

	dist := 64 - n
	mask := ^uint64(0) << dist
	result := (br.buffer & mask) >> dist
	return result, nil
}

func (br *bitreader) skip(n uint) error {
	for n > 0 {
		if n > br.remaining {
			if err := br.fill(); err != nil {
				return err
			}
		}

		len := n
		if len > br.remaining {
			len = br.remaining
		}

		br.buffer <<= len
		br.remaining -= len
		n -= len
	}

	return nil
}

func newBitReader(reader io.Reader) *bitreader {
	return &bitreader{
		r: reader,
	}
}
