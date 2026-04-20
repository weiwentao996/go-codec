package codec

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

func Marshal(v interface{}, opts ...Option) ([]byte, error) {
	cfg := buildConfig(opts...)
	buf := new(bytes.Buffer)
	if err := encodeValueTree(buf, v, cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeValueTree(buf *bytes.Buffer, v interface{}, cfg Config) error {
	vv, err := getReflectAndInitObj(v)
	if err != nil {
		return err
	}

	writer := newBitEncoder(buf, cfg.BitLayout)
	if vv.Kind() != reflect.Struct {
		if vv.Kind() == reflect.Array || vv.Kind() == reflect.Slice {
			if err := encodeSliceOrArray(writer, vv, nil, cfg); err != nil {
				return err
			}
		} else {
			if err := encodeValue(buf, vv, nil, cfg); err != nil {
				return err
			}
		}
		return writer.flushPending()
	}

	for i := 0; i < vv.NumField(); i++ {
		fieldValue := vv.Field(i)
		fieldType := vv.Type().Field(i)
		switch fieldValue.Kind() {
		case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
			if fieldValue.IsNil() {
				continue
			}
		}

		tags, err := parseEncodeTag(fieldType.Tag)
		if err != nil {
			return err
		}
		if tags.Ignore {
			continue
		}
		if tags.File {
			fileBytes, err := readFile(fieldValue.String(), cfg)
			if err != nil {
				return err
			}
			if _, err := buf.Write(fileBytes); err != nil {
				return err
			}
			continue
		}
		if tags.BitCount != 0 {
			if err := writer.writeValueBits(fieldValue, tags.BitCount, cfg); err != nil {
				return err
			}
			continue
		}

		switch fieldValue.Kind() {
		case reflect.Struct, reflect.Ptr, reflect.Interface:
			if err := encodeValueTree(buf, fieldValue, cfg); err != nil {
				return err
			}
		case reflect.Array, reflect.Slice:
			if err := encodeSliceOrArray(writer, fieldValue, tags, cfg); err != nil {
				return err
			}
		default:
			if err := encodeValue(buf, fieldValue, tags, cfg); err != nil {
				return err
			}
		}
	}
	return writer.flushPending()
}

func encodeSliceOrArray(writer *bitEncoder, fieldValue reflect.Value, tags *fieldTag, cfg Config) error {
	for i := 0; i < fieldValue.Len(); i++ {
		currentValue := fieldValue.Index(i)
		if currentValue.Kind() == reflect.Struct || currentValue.Kind() == reflect.Array || currentValue.Kind() == reflect.Slice {
			if err := encodeValueTree(writer.buf, currentValue, cfg); err != nil {
				return err
			}
			continue
		}
		if tags == nil || tags.SubBitCount == 0 {
			if err := encodeValue(writer.buf, currentValue, nil, cfg); err != nil {
				return err
			}
			continue
		}
		if err := writer.writeValueBits(currentValue, tags.SubBitCount, cfg); err != nil {
			return err
		}
	}
	return nil
}

func encodeValue(buf *bytes.Buffer, v reflect.Value, tags *fieldTag, cfg Config) error {
	order := cfg.byteOrder()
	if tags == nil {
		tags = &fieldTag{}
	}
	switch v.Kind() {
	case reflect.Bool:
		return binary.Write(buf, order, v.Bool())
	case reflect.Int8:
		return binary.Write(buf, order, int8(v.Int()))
	case reflect.Uint8:
		return binary.Write(buf, order, uint8(v.Uint()))
	case reflect.Int16:
		return binary.Write(buf, order, int16(v.Int()))
	case reflect.Uint16:
		return binary.Write(buf, order, uint16(v.Uint()))
	case reflect.Int32:
		return binary.Write(buf, order, int32(v.Int()))
	case reflect.Uint32:
		return binary.Write(buf, order, uint32(v.Uint()))
	case reflect.Int64:
		return binary.Write(buf, order, v.Int())
	case reflect.Uint64:
		return binary.Write(buf, order, v.Uint())
	case reflect.Float32:
		return binary.Write(buf, order, float32(v.Float()))
	case reflect.Float64:
		return binary.Write(buf, order, v.Float())
	case reflect.String:
		return binary.Write(buf, order, lengthAlignRightPadding([]byte(v.String()), tags.ByteCount))
	default:
		return nil
	}
}
