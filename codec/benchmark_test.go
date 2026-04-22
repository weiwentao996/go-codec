package codec

import "testing"

type benchmarkTaggedFixture struct {
	Prefix uint16
	Flags  [8]uint8 `encode:"subBitCount:1" decode:"subBitCount:1"`
	A      uint8    `encode:"bitCount:3" decode:"bitCount:3"`
	B      uint8    `encode:"bitCount:5" decode:"bitCount:5"`
	Name   string   `encode:"byteCount:8" decode:"byteCount:8"`
	Value  uint32
}

type benchmarkNestedLeaf struct {
	Code uint16
	Name string `encode:"byteCount:4" decode:"byteCount:4"`
}

type benchmarkNestedFixture struct {
	Header uint8
	Left   benchmarkNestedLeaf
	Right  benchmarkNestedLeaf
	Tail   uint32
}

func BenchmarkMarshalStructSmall(b *testing.B) {
	fixture := endianFixture{Value16: 0x1234, Value32: 0x01020304}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Marshal(fixture); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalStructWithTags(b *testing.B) {
	fixture := benchmarkTaggedFixture{
		Prefix: 0x1234,
		Flags:  [8]uint8{1, 0, 1, 0, 1, 0, 1, 0},
		A:      5,
		B:      17,
		Name:   "codec",
		Value:  0x01020304,
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Marshal(fixture); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalNestedStruct(b *testing.B) {
	fixture := benchmarkNestedFixture{
		Header: 1,
		Left: benchmarkNestedLeaf{
			Code: 0x1234,
			Name: "left",
		},
		Right: benchmarkNestedLeaf{
			Code: 0x5678,
			Name: "rght",
		},
		Tail: 0x01020304,
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Marshal(fixture); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalStructSmall(b *testing.B) {
	data, err := Marshal(endianFixture{Value16: 0x1234, Value32: 0x01020304})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var decoded endianFixture
		if err := Unmarshal(data, &decoded); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalStructWithTags(b *testing.B) {
	fixture := benchmarkTaggedFixture{
		Prefix: 0x1234,
		Flags:  [8]uint8{1, 0, 1, 0, 1, 0, 1, 0},
		A:      5,
		B:      17,
		Name:   "codec",
		Value:  0x01020304,
	}
	data, err := Marshal(fixture)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var decoded benchmarkTaggedFixture
		if err := Unmarshal(data, &decoded); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalNestedStruct(b *testing.B) {
	fixture := benchmarkNestedFixture{
		Header: 1,
		Left: benchmarkNestedLeaf{
			Code: 0x1234,
			Name: "left",
		},
		Right: benchmarkNestedLeaf{
			Code: 0x5678,
			Name: "rght",
		},
		Tail: 0x01020304,
	}
	data, err := Marshal(fixture)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var decoded benchmarkNestedFixture
		if err := Unmarshal(data, &decoded); err != nil {
			b.Fatal(err)
		}
	}
}
