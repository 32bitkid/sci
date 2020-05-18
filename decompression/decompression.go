package decompression

import (
	"compress/lzw"
	"io"
)

type Method uint16

type Decompressor = func(src io.Reader, dest io.Writer, compressedSize, decompressedSize uint16) error

type LUT map[Method]Decompressor

func DecompressNone(src io.Reader, dst io.Writer, compressedSize, decompressedSize uint16) error {
	lr := io.LimitReader(src, int64(compressedSize))
	_, err := io.Copy(dst, lr)
	return err
}

func DecompressLZW(src io.Reader, dst io.Writer, compressedSize, decompressedSize uint16) error {
	lr := io.LimitReader(src, int64(compressedSize))
	lzwr := lzw.NewReader(lr, lzw.LSB, 8)
	_, err := io.Copy(dst, lzwr)
	return err
}

func DecompressHuffman(src io.Reader, dst io.Writer, compressedSize, decompressedSize uint16) error {
	lr := io.LimitReader(src, int64(compressedSize))
	return huffman(dst, lr, int(decompressedSize))
}

func DecompressLZW1(src io.Reader, dst io.Writer, compressedSize, decompressedSize uint16) error {
	lr := io.LimitReader(src, int64(compressedSize))
	return lzw1(dst, lr, int(decompressedSize))
}

var Decompressors = struct {
	SCI0  LUT
	SCI01 LUT
}{
	SCI0: LUT{
		0: DecompressNone,
		1: DecompressLZW,
		2: DecompressHuffman,
	},
	SCI01: LUT{
		0: DecompressNone,
		1: DecompressHuffman,
		2: DecompressLZW1,
	},
}
