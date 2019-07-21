package resource

import (
	"compress/lzw"
	"errors"
	"io"
)

type CompressionMethod uint16

type DecompressionFn = func(r io.Reader, dest []byte, compressedSize, decompressedSize uint16) error

type DecompressorLUT map[CompressionMethod]DecompressionFn

func DecompressNone(r io.Reader, dst []byte, compressedSize, decompressedSize uint16) error {
	buf := dst[:decompressedSize]
	lr := io.LimitReader(r, int64(compressedSize))
	_, err := io.ReadFull(lr, buf)
	return err
}

func DecompressLZW(r io.Reader, dst []byte, compressedSize, decompressedSize uint16) error {
	buf := dst[:decompressedSize]
	lr := io.LimitReader(r, int64(compressedSize))
	lzwr := lzw.NewReader(lr, lzw.LSB, 8)
	_, err := io.ReadFull(lzwr, buf)
	return err
}

func DecompressHuffman(r io.Reader, dst []byte, compressedSize, decompressedSize uint16) error {
	buf := dst[:decompressedSize]
	lr := io.LimitReader(r, int64(compressedSize))
	err := huffman(lr, buf)
	return err
}

func DecompressLZW1(r io.Reader, dst []byte, compressedSize, decompressedSize uint16) error {
	return errors.New("not implemented: SCI01 LZW decompression")
}

func DecompressCOMP3(r io.Reader, dst []byte, compressedSize uint16, decompressedSize uint16) error {
	return errors.New("not implemented: SCI01 COMP3 decompression")
}

var Decompressors = struct {
	SCI0  DecompressorLUT
	SCI01 DecompressorLUT
}{
	SCI0: DecompressorLUT{
		0: DecompressNone,
		1: DecompressLZW,
		2: DecompressHuffman,
	},
	SCI01: DecompressorLUT{
		0: DecompressNone,
		1: DecompressLZW1,
		2: DecompressCOMP3,
		3: DecompressHuffman,
	},
}
