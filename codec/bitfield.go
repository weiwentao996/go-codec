package codec

import (
	"bytes"
	"fmt"
	"reflect"
)

type bitBuffer struct {
	buf *bytes.Buffer
}

func newBitBuffer() *bitBuffer {
	return &bitBuffer{buf: bytes.NewBuffer(nil)}
}

func (bb *bitBuffer) reset() {
	bb.buf.Reset()
}

func (bb *bitBuffer) write(b bool) {
	if b {
		bb.buf.WriteByte(1)
	} else {
		bb.buf.WriteByte(0)
	}
}

func (bb *bitBuffer) bytes(layout BitLayoutMode) []byte {
	data := bb.buf.Bytes()
	length := (len(data) + 7) / 8
	result := make([]byte, length)
	for i, datum := range data {
		index := i / 8
		var offset int
		if layout == LSBFirstLowToHigh {
			offset = i % 8
		} else {
			offset = 7 - i%8
		}
		result[index] = datum<<offset | result[index]
	}
	return result
}

func bytesToBits(data []byte, layout BitLayoutMode) []bool {
	bits := make([]bool, 0, len(data)*8)
	if layout == LSBFirstLowToHigh {
		for i := 0; i < len(data); i++ {
			for j := 0; j < 8; j++ {
				bit := (data[i] >> uint(j)) & 1
				bits = append(bits, bit == 1)
			}
		}
		return bits
	}

	for i := len(data) - 1; i >= 0; i-- {
		for j := 0; j < 8; j++ {
			bit := (data[i] >> uint(j)) & 1
			bits = append(bits, bit == 1)
		}
	}
	return bits
}

func encodeBitField(writeData *bitBuffer, bs []bool, size int, layout BitLayoutMode) {
	if layout == LSBFirstLowToHigh {
		for i := 0; i < size; i++ {
			writeData.write(bs[i])
		}
		return
	}

	for i := 0; i < size; i++ {
		writeData.write(bs[len(bs)-1-i])
	}
}

func decodeBitField(data []byte, offset, size int, layout BitLayoutMode) (result uint64, index int) {
	if layout == LSBFirstLowToHigh {
		for i := 0; i < size; i++ {
			byteIndex := offset / 8
			currentByte := data[byteIndex]
			currentBit := (currentByte >> (offset % 8)) & 1
			result = result | uint64(currentBit)<<i
			offset++
		}
		return result, offset
	}

	for i := size - 1; i >= 0; i-- {
		byteIndex := offset / 8
		currentByte := data[byteIndex]
		currentBit := (currentByte >> (7 - (offset % 8))) & 1
		result = result | uint64(currentBit)<<i
		offset++
	}
	return result, offset
}

type bitEncoder struct {
	buf     *bytes.Buffer
	pending *bitBuffer
	layout  BitLayoutMode
}

func newBitEncoder(buf *bytes.Buffer, layout BitLayoutMode) *bitEncoder {
	return &bitEncoder{buf: buf, pending: newBitBuffer(), layout: layout}
}

func (e *bitEncoder) writeValueBits(v reflect.Value, bitCount int, cfg Config) error {
	tmp := new(bytes.Buffer)
	if err := encodeValue(tmp, v, nil, cfg); err != nil {
		return err
	}
	dataBits := bytesToBits(tmp.Bytes(), e.layout)
	if len(dataBits) < bitCount {
		return fmt.Errorf("encode bitCount %d exceeds field size %d", bitCount, len(dataBits))
	}
	encodeBitField(e.pending, dataBits[:bitCount], bitCount, e.layout)
	return e.flushFullBytes()
}

func (e *bitEncoder) flushFullBytes() error {
	if e.pending.buf.Len() == 0 || e.pending.buf.Len()%8 != 0 {
		return nil
	}
	if _, err := e.buf.Write(e.pending.bytes(e.layout)); err != nil {
		return err
	}
	e.pending.reset()
	return nil
}

func (e *bitEncoder) flushPending() error {
	if e.pending.buf.Len() == 0 {
		return nil
	}
	if _, err := e.buf.Write(e.pending.bytes(e.layout)); err != nil {
		return err
	}
	e.pending.reset()
	return nil
}

type bitDecoder struct {
	buf          *bytes.Buffer
	currentBytes []byte
	index        int
	layout       BitLayoutMode
}

func newBitDecoder(buf *bytes.Buffer, layout BitLayoutMode) *bitDecoder {
	return &bitDecoder{buf: buf, layout: layout}
}

func (d *bitDecoder) readBits(bitCount int) (uint64, error) {
	for (d.index + bitCount) > len(d.currentBytes)*8 {
		readByte, err := d.buf.ReadByte()
		if err != nil {
			return 0, err
		}
		d.currentBytes = append(d.currentBytes, readByte)
	}
	result, newIndex := decodeBitField(d.currentBytes, d.index, bitCount, d.layout)
	d.index = newIndex
	return result, nil
}
