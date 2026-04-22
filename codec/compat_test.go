package codec

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type encodeSubBitCountFixture struct {
	Flags [8]uint8 `encode:"subBitCount:1"`
}

type decodeSubBitCountFixture struct {
	Flags [8]uint8 `decode:"subBitCount:1"`
}

type trailingBitFieldFixture struct {
	A uint8 `encode:"bitCount:1"`
	B uint8 `encode:"bitCount:1"`
	C uint8 `encode:"bitCount:1"`
}

type invalidEncodeTagFixture struct {
	Value uint8 `encode:"bitCount:"`
}

type invalidDecodeTagFixture struct {
	Value uint8 `decode:"bitCount:"`
}

type invalidEncodeBitCountTypeFixture struct {
	Value string `encode:"bitCount:3"`
}

type invalidDecodeBitCountTypeFixture struct {
	Value string `decode:"bitCount:3"`
}

type invalidEncodeSubBitCountTypeFixture struct {
	Flags string `encode:"subBitCount:1"`
}

type invalidDecodeSubBitCountTypeFixture struct {
	Flags string `decode:"subBitCount:1"`
}

type invalidEncodeByteCountTypeFixture struct {
	Value uint8 `encode:"byteCount:4"`
}

type invalidDecodeByteCountTypeFixture struct {
	Value uint8 `decode:"byteCount:4"`
}

type invalidEncodeFileTypeFixture struct {
	Value uint8 `encode:"file"`
}

type invalidDecodeFileTypeFixture struct {
	Value uint8 `decode:"file"`
}

type mixedBitFieldEncodeFixture struct {
	Prefix uint8
	A      uint8 `encode:"bitCount:3"`
	B      uint8 `encode:"bitCount:3"`
	C      uint8 `encode:"bitCount:2"`
	Suffix uint8
}

type mixedBitFieldDecodeFixture struct {
	Prefix uint8
	A      uint8 `decode:"bitCount:3"`
	B      uint8 `decode:"bitCount:3"`
	C      uint8 `decode:"bitCount:2"`
	Suffix uint8
}

type subBitBoundaryEncodeFixture struct {
	Flags [10]uint8 `encode:"subBitCount:1"`
}

type subBitBoundaryDecodeFixture struct {
	Flags [10]uint8 `decode:"subBitCount:1"`
}

type fixedStringEncodeFixture struct {
	Name string `encode:"byteCount:5"`
}

type fixedStringDecodeFixture struct {
	Name string `decode:"byteCount:5"`
}

type encodeFileFixture struct {
	Path string `encode:"file"`
}

type decodeFileFixture struct {
	Path string `decode:"file"`
}

type endianFixture struct {
	Value16 uint16
	Value32 uint32
}

type bitThenByteEncodeFixture struct {
	A uint8 `encode:"bitCount:3"`
	B uint8
}

type bitThenByteDecodeFixture struct {
	A uint8 `decode:"bitCount:3"`
	B uint8
}

type subBitThenByteEncodeFixture struct {
	Flags [3]uint8 `encode:"subBitCount:1"`
	B     uint8
}

type subBitThenByteDecodeFixture struct {
	Flags [3]uint8 `decode:"subBitCount:1"`
	B     uint8
}

type sequentialScalarFixture struct {
	Value uint16
}

func TestEncodeSubBitCountArray(t *testing.T) {
	data, err := Marshal(encodeSubBitCountFixture{Flags: [8]uint8{1, 0, 1, 0, 1, 0, 1, 0}})
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xAA}, data)
}

func TestDecodeSubBitCountArray(t *testing.T) {
	var result decodeSubBitCountFixture
	err := Unmarshal([]byte{0xAA}, &result)
	assert.NoError(t, err)
	assert.Equal(t, [8]uint8{1, 0, 1, 0, 1, 0, 1, 0}, result.Flags)
}

func TestEncodeTrailingBitFieldFlush(t *testing.T) {
	data, err := Marshal(trailingBitFieldFixture{A: 1, B: 0, C: 1})
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xA0}, data)
}

func TestBitFieldBeforeByteFieldAlignsToNextByte(t *testing.T) {
	data, err := Marshal(bitThenByteEncodeFixture{A: 5, B: 0x12})
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xA0, 0x12}, data)

	var decoded bitThenByteDecodeFixture
	err = Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, uint8(5), decoded.A)
	assert.Equal(t, uint8(0x12), decoded.B)
}

func TestSubBitArrayBeforeByteFieldAlignsToNextByte(t *testing.T) {
	data, err := Marshal(subBitThenByteEncodeFixture{Flags: [3]uint8{1, 0, 1}, B: 0x12})
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xA0, 0x12}, data)

	var decoded subBitThenByteDecodeFixture
	err = Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, [3]uint8{1, 0, 1}, decoded.Flags)
	assert.Equal(t, uint8(0x12), decoded.B)
}

func TestParseEncodeTagReturnsErrorOnInvalidNumber(t *testing.T) {
	_, err := Marshal(invalidEncodeTagFixture{Value: 1})
	assert.Error(t, err)
}

func TestParseDecodeTagReturnsErrorOnInvalidNumber(t *testing.T) {
	var result invalidDecodeTagFixture
	err := Unmarshal([]byte{0x01}, &result)
	assert.Error(t, err)
}

func TestEncodeBitCountRejectsNonUnsignedIntegerField(t *testing.T) {
	_, err := Marshal(invalidEncodeBitCountTypeFixture{Value: "abc"})
	assert.EqualError(t, err, "bitCount tag requires unsigned integer field, got string")
}

func TestDecodeBitCountRejectsNonUnsignedIntegerField(t *testing.T) {
	var result invalidDecodeBitCountTypeFixture
	err := Unmarshal([]byte{0xA0}, &result)
	assert.EqualError(t, err, "bitCount tag requires unsigned integer field, got string")
}

func TestEncodeSubBitCountRejectsNonArrayOrSliceField(t *testing.T) {
	_, err := Marshal(invalidEncodeSubBitCountTypeFixture{Flags: "abc"})
	assert.EqualError(t, err, "subBitCount tag requires array or slice field, got string")
}

func TestDecodeSubBitCountRejectsNonArrayOrSliceField(t *testing.T) {
	var result invalidDecodeSubBitCountTypeFixture
	err := Unmarshal([]byte{0xA0}, &result)
	assert.EqualError(t, err, "subBitCount tag requires array or slice field, got string")
}

func TestEncodeByteCountRejectsNonStringField(t *testing.T) {
	_, err := Marshal(invalidEncodeByteCountTypeFixture{Value: 1})
	assert.EqualError(t, err, "byteCount tag requires string field, got uint8")
}

func TestDecodeByteCountRejectsNonStringField(t *testing.T) {
	var result invalidDecodeByteCountTypeFixture
	err := Unmarshal([]byte{0x01, 0x02, 0x03, 0x04}, &result)
	assert.EqualError(t, err, "byteCount tag requires string field, got uint8")
}

func TestEncodeFileRejectsNonStringField(t *testing.T) {
	_, err := Marshal(invalidEncodeFileTypeFixture{Value: 1})
	assert.EqualError(t, err, "file tag requires string field, got uint8")
}

func TestDecodeFileRejectsNonStringField(t *testing.T) {
	var result invalidDecodeFileTypeFixture
	err := Unmarshal([]byte{0x01}, &result)
	assert.EqualError(t, err, "file tag requires string field, got uint8")
}

func TestMixedBitFieldsEncodeAndDecode(t *testing.T) {
	input := mixedBitFieldEncodeFixture{Prefix: 0x12, A: 5, B: 2, C: 3, Suffix: 0x34}
	data, err := Marshal(input)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x12, 0xAB, 0x34}, data)

	var decoded mixedBitFieldDecodeFixture
	err = Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, uint8(0x12), decoded.Prefix)
	assert.Equal(t, uint8(5), decoded.A)
	assert.Equal(t, uint8(2), decoded.B)
	assert.Equal(t, uint8(3), decoded.C)
	assert.Equal(t, uint8(0x34), decoded.Suffix)
}

func TestSubBitCountCrossesByteBoundary(t *testing.T) {
	input := [10]uint8{1, 0, 1, 0, 1, 0, 1, 0, 1, 1}
	data, err := Marshal(subBitBoundaryEncodeFixture{Flags: input})
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xAA, 0xC0}, data)

	var decoded subBitBoundaryDecodeFixture
	err = Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, input, decoded.Flags)
}

func TestDecodeSupportsInterfacePointerAndDirectPointer(t *testing.T) {
	var asInterface interface{} = &decodeSubBitCountFixture{}
	err := Unmarshal([]byte{0xAA}, &asInterface)
	assert.NoError(t, err)
	assert.Equal(t, [8]uint8{1, 0, 1, 0, 1, 0, 1, 0}, asInterface.(*decodeSubBitCountFixture).Flags)

	direct := &decodeSubBitCountFixture{}
	err = Unmarshal([]byte{0xAA}, direct)
	assert.NoError(t, err)
	assert.Equal(t, [8]uint8{1, 0, 1, 0, 1, 0, 1, 0}, direct.Flags)
}

func TestFixedLengthStringCompatibility(t *testing.T) {
	data, err := Marshal(fixedStringEncodeFixture{Name: "ab"})
	assert.NoError(t, err)
	assert.Equal(t, []byte{'a', 'b', 0, 0, 0}, data)

	var decoded fixedStringDecodeFixture
	err = Unmarshal([]byte{'a', 'b', 0, 0, 0}, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "ab", decoded.Name)
}

func TestEncodeFileTagRequiresConfiguredReader(t *testing.T) {
	_, err := Marshal(encodeFileFixture{Path: "demo.bin"})
	assert.ErrorIs(t, err, ErrFileReaderNotConfigured)
}

func TestDecodeFileTagRequiresConfiguredWriter(t *testing.T) {
	var decoded decodeFileFixture
	err := Unmarshal([]byte{1, 2, 3}, &decoded)
	assert.ErrorIs(t, err, ErrFileWriterNotConfigured)
}

func TestFileTagUsesHooksWhenConfigured(t *testing.T) {
	data, err := Marshal(encodeFileFixture{Path: "demo.bin"}, WithFileReader(func(path string) ([]byte, error) {
		assert.Equal(t, "demo.bin", path)
		return []byte{1, 2, 3}, nil
	}))
	assert.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3}, data)

	var decoded decodeFileFixture
	err = Unmarshal([]byte{4, 5, 6}, &decoded, WithFileWriter(func(typeName string, data []byte) (string, error) {
		assert.Equal(t, "decodeFileFixture", typeName)
		assert.Equal(t, []byte{4, 5, 6}, data)
		return "saved/path", nil
	}))
	assert.NoError(t, err)
	assert.Equal(t, "saved/path", decoded.Path)
}

func TestMarshalAndUnmarshalWithLegacyPreset(t *testing.T) {
	data, err := Marshal(encodeSubBitCountFixture{Flags: [8]uint8{1, 0, 1, 0, 1, 0, 1, 0}}, WithLegacyHKimPreset())
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xAA}, data)

	var decoded decodeSubBitCountFixture
	err = Unmarshal(data, &decoded, WithLegacyHKimPreset())
	assert.NoError(t, err)
	assert.Equal(t, [8]uint8{1, 0, 1, 0, 1, 0, 1, 0}, decoded.Flags)
}

func TestFileReaderPropagatesError(t *testing.T) {
	expected := errors.New("read failed")
	_, err := Marshal(encodeFileFixture{Path: "broken.bin"}, WithFileReader(func(path string) ([]byte, error) {
		return nil, expected
	}))
	assert.ErrorIs(t, err, expected)
}

func TestFileWriterConsumesRemainingBytes(t *testing.T) {
	var decoded decodeFileFixture
	err := Unmarshal(bytes.NewBuffer([]byte{7, 8, 9}).Bytes(), &decoded, WithFileWriter(func(typeName string, data []byte) (string, error) {
		assert.Equal(t, []byte{7, 8, 9}, data)
		return "ok", nil
	}))
	assert.NoError(t, err)
	assert.Equal(t, "ok", decoded.Path)
}

func TestDefaultMatchesLegacyPreset(t *testing.T) {
	fixture := mixedBitFieldEncodeFixture{Prefix: 0x12, A: 5, B: 2, C: 3, Suffix: 0x34}
	defaultBytes, err := Marshal(fixture)
	assert.NoError(t, err)

	legacyBytes, err := Marshal(fixture, WithLegacyHKimPreset())
	assert.NoError(t, err)
	assert.Equal(t, legacyBytes, defaultBytes)
}

func TestDecodeMatchesUnmarshalForSingleValue(t *testing.T) {
	data, err := Marshal(mixedBitFieldEncodeFixture{Prefix: 0x12, A: 5, B: 2, C: 3, Suffix: 0x34})
	assert.NoError(t, err)

	var fromBytes mixedBitFieldDecodeFixture
	err = Unmarshal(data, &fromBytes)
	assert.NoError(t, err)

	buf := bytes.NewBuffer(data)
	var fromBuffer mixedBitFieldDecodeFixture
	err = Decode(buf, &fromBuffer)
	assert.NoError(t, err)
	assert.Equal(t, fromBytes, fromBuffer)
	assert.Equal(t, 0, buf.Len())
}

func TestDecodeConsumesBufferSequentially(t *testing.T) {
	first := sequentialScalarFixture{Value: 0x1234}
	second := sequentialScalarFixture{Value: 0x5678}

	firstBytes, err := Marshal(first)
	assert.NoError(t, err)
	secondBytes, err := Marshal(second)
	assert.NoError(t, err)

	buf := bytes.NewBuffer(append(firstBytes, secondBytes...))
	assert.Equal(t, len(firstBytes)+len(secondBytes), buf.Len())

	var firstDecoded sequentialScalarFixture
	err = Decode(buf, &firstDecoded)
	assert.NoError(t, err)
	assert.Equal(t, first, firstDecoded)
	assert.Equal(t, len(secondBytes), buf.Len())

	var secondDecoded sequentialScalarFixture
	err = Decode(buf, &secondDecoded)
	assert.NoError(t, err)
	assert.Equal(t, second, secondDecoded)
	assert.Equal(t, 0, buf.Len())
}

func TestLittleEndianChangesScalarEncodingAndDecoding(t *testing.T) {
	fixture := endianFixture{Value16: 0x1234, Value32: 0x01020304}

	bigBytes, err := Marshal(fixture)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x12, 0x34, 0x01, 0x02, 0x03, 0x04}, bigBytes)

	littleBytes, err := Marshal(fixture, WithByteOrder(LittleEndian))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x34, 0x12, 0x04, 0x03, 0x02, 0x01}, littleBytes)

	var decoded endianFixture
	err = Unmarshal(littleBytes, &decoded, WithByteOrder(LittleEndian))
	assert.NoError(t, err)
	assert.Equal(t, fixture, decoded)
}

func TestAlternativeBitLayoutChangesBitCountEncoding(t *testing.T) {
	fixture := mixedBitFieldEncodeFixture{Prefix: 0x12, A: 5, B: 2, C: 3, Suffix: 0x34}

	legacyBytes, err := Marshal(fixture)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x12, 0xAB, 0x34}, legacyBytes)

	altBytes, err := Marshal(fixture, WithBitLayout(LSBFirstLowToHigh))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x12, 0xD5, 0x34}, altBytes)
	assert.NotEqual(t, legacyBytes, altBytes)

	var decoded mixedBitFieldDecodeFixture
	err = Unmarshal(altBytes, &decoded, WithBitLayout(LSBFirstLowToHigh))
	assert.NoError(t, err)
	assert.Equal(t, uint8(0x12), decoded.Prefix)
	assert.Equal(t, uint8(5), decoded.A)
	assert.Equal(t, uint8(2), decoded.B)
	assert.Equal(t, uint8(3), decoded.C)
	assert.Equal(t, uint8(0x34), decoded.Suffix)
}

func TestAlternativeBitLayoutChangesSubBitEncoding(t *testing.T) {
	fixture := subBitBoundaryEncodeFixture{Flags: [10]uint8{1, 0, 1, 0, 1, 0, 1, 0, 1, 1}}

	legacyBytes, err := Marshal(fixture)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xAA, 0xC0}, legacyBytes)

	altBytes, err := Marshal(fixture, WithBitLayout(LSBFirstLowToHigh))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x55, 0x03}, altBytes)
	assert.NotEqual(t, legacyBytes, altBytes)

	var decoded subBitBoundaryDecodeFixture
	err = Unmarshal(altBytes, &decoded, WithBitLayout(LSBFirstLowToHigh))
	assert.NoError(t, err)
	assert.Equal(t, [10]uint8{1, 0, 1, 0, 1, 0, 1, 0, 1, 1}, decoded.Flags)
}

func TestMixedConfigurationWorksTogether(t *testing.T) {
	fixture := struct {
		Value uint16
		A     uint8 `encode:"bitCount:3"`
		B     uint8 `encode:"bitCount:5"`
	}{
		Value: 0x1234,
		A:     5,
		B:     17,
	}

	data, err := Marshal(fixture, WithByteOrder(LittleEndian), WithBitLayout(LSBFirstLowToHigh))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x34, 0x12, 0x8D}, data)
}

func TestMismatchedBitLayoutDoesNotDecodeLegacyBytesTheSame(t *testing.T) {
	legacyBytes, err := Marshal(mixedBitFieldEncodeFixture{Prefix: 0x12, A: 5, B: 2, C: 3, Suffix: 0x34})
	assert.NoError(t, err)

	var decoded mixedBitFieldDecodeFixture
	err = Unmarshal(legacyBytes, &decoded, WithBitLayout(LSBFirstLowToHigh))
	assert.NoError(t, err)
	assert.NotEqual(t, uint8(5), decoded.A)
}

func TestMetadataCacheRepeatedMarshalProducesSameBytes(t *testing.T) {
	fixture := mixedBitFieldEncodeFixture{Prefix: 0x12, A: 5, B: 2, C: 3, Suffix: 0x34}

	first, err := Marshal(fixture)
	assert.NoError(t, err)

	second, err := Marshal(fixture)
	assert.NoError(t, err)
	assert.Equal(t, first, second)
}

func TestMetadataCacheRepeatedUnmarshalProducesSameResult(t *testing.T) {
	data := []byte{0x12, 0xAB, 0x34}

	var first mixedBitFieldDecodeFixture
	err := Unmarshal(data, &first)
	assert.NoError(t, err)

	var second mixedBitFieldDecodeFixture
	err = Unmarshal(data, &second)
	assert.NoError(t, err)
	assert.Equal(t, first, second)
}

func TestMetadataCachePreservesConfigSpecificEncoding(t *testing.T) {
	fixture := mixedBitFieldEncodeFixture{Prefix: 0x12, A: 5, B: 2, C: 3, Suffix: 0x34}

	legacyBytes, err := Marshal(fixture)
	assert.NoError(t, err)

	altBytes, err := Marshal(fixture, WithBitLayout(LSBFirstLowToHigh))
	assert.NoError(t, err)
	assert.NotEqual(t, legacyBytes, altBytes)

	bigEndian, err := Marshal(endianFixture{Value16: 0x1234, Value32: 0x01020304})
	assert.NoError(t, err)

	littleEndian, err := Marshal(endianFixture{Value16: 0x1234, Value32: 0x01020304}, WithByteOrder(LittleEndian))
	assert.NoError(t, err)
	assert.NotEqual(t, bigEndian, littleEndian)
}

func TestMetadataCacheIsConcurrentSafe(t *testing.T) {
	fixture := mixedBitFieldEncodeFixture{Prefix: 0x12, A: 5, B: 2, C: 3, Suffix: 0x34}
	expected, err := Marshal(fixture)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			data, err := Marshal(fixture)
			assert.NoError(t, err)
			assert.Equal(t, expected, data)

			var decoded mixedBitFieldDecodeFixture
			err = Unmarshal(expected, &decoded)
			assert.NoError(t, err)
			assert.Equal(t, uint8(0x12), decoded.Prefix)
			assert.Equal(t, uint8(5), decoded.A)
			assert.Equal(t, uint8(2), decoded.B)
			assert.Equal(t, uint8(3), decoded.C)
			assert.Equal(t, uint8(0x34), decoded.Suffix)
		}()
	}
	wg.Wait()
}
