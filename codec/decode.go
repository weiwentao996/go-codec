package codec

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
)

func Unmarshal(data []byte, v interface{}, opts ...Option) error {
	cfg := buildConfig(opts...)
	return decodeValueTree(bytes.NewBuffer(data), v, cfg)
}

func decodeValueTree(buf *bytes.Buffer, v interface{}, cfg Config) error {
	vv, err := getReflectAndInitObj(v)
	if err != nil {
		return err
	}

	reader := newBitDecoder(buf, cfg.BitLayout)
	if vv.Kind() != reflect.Struct {
		if vv.Kind() == reflect.Array || vv.Kind() == reflect.Slice {
			l := calLen(buf, vv, nil)
			if err := decodeSliceOrArray(buf, reader, l, vv, nil, cfg); err != nil {
				return err
			}
		} else {
			if err := decodeValue(buf, vv, nil, cfg); err != nil {
				return err
			}
		}
		return nil
	}

	for i := 0; i < vv.NumField(); i++ {
		fieldValue := vv.Field(i)
		fieldType := vv.Type().Field(i)
		tags, err := parseDecodeTag(fieldType.Tag)
		if err != nil {
			return err
		}
		if tags.Ignore {
			continue
		}
		if tags.File {
			filePath, err := writeFile(vv.Type().Name(), buf.Bytes(), cfg)
			if err != nil {
				return err
			}
			fieldValue.SetString(filePath)
			buf.Reset()
			continue
		}
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() && fieldValue.CanAddr() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
		}
		if tags.BitCount != 0 {
			result, err := reader.readBits(tags.BitCount)
			if err != nil {
				return err
			}
			fieldValue.SetUint(result)
			continue
		}
		switch fieldValue.Kind() {
		case reflect.Ptr, reflect.Interface:
			if err := decodeValueTree(buf, fieldValue, cfg); err != nil {
				return err
			}
		case reflect.Array, reflect.Slice:
			l := calLen(buf, fieldValue, tags)
			if err := decodeSliceOrArray(buf, reader, l, fieldValue, tags, cfg); err != nil {
				return err
			}
		default:
			if err := decodeValue(buf, fieldValue, tags, cfg); err != nil {
				return err
			}
		}
	}
	return nil
}

func calLen(buf *bytes.Buffer, fieldValue reflect.Value, tags *fieldTag) int {
	var l int
	if fieldValue.Kind() == reflect.Slice {
		l = fieldValue.Cap()
		if fieldValue.Cap() == 0 {
			if tags == nil || tags.ByteCount == 0 {
				l = buf.Len() / int(fieldValue.Type().Elem().Size())
			} else {
				l = buf.Len() / tags.ByteCount
			}
			fieldValue.Set(reflect.MakeSlice(fieldValue.Type(), l, l))
		}
	}
	if fieldValue.Kind() == reflect.Array {
		l = fieldValue.Cap()
	}
	return l
}

func decodeSliceOrArray(buf *bytes.Buffer, reader *bitDecoder, l int, fieldValue reflect.Value, tags *fieldTag, cfg Config) error {
	for j := 0; j < l; j++ {
		currentValue := fieldValue.Index(j)
		switch {
		case currentValue.Kind() == reflect.Ptr:
			if err := decodeValueTree(buf, currentValue, cfg); err != nil {
				return err
			}
		case currentValue.Kind() == reflect.Array:
			if err := decodeValueTree(buf, currentValue, cfg); err != nil {
				return err
			}
		case tags != nil && tags.SubBitCount != 0:
			result, err := reader.readBits(tags.SubBitCount)
			if err != nil {
				return err
			}
			currentValue.SetUint(result)
		default:
			if err := decodeValue(buf, currentValue, nil, cfg); err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeValue(buf *bytes.Buffer, v reflect.Value, tag *fieldTag, cfg Config) error {
	order := cfg.byteOrder()
	if tag == nil {
		tag = &fieldTag{}
	}
	switch v.Kind() {
	case reflect.Bool:
		var b bool
		if err := binary.Read(buf, order, &b); err != nil { return err }
		v.SetBool(b)
	case reflect.Int8:
		var b int8
		if err := binary.Read(buf, order, &b); err != nil { return err }
		v.SetInt(int64(b))
	case reflect.Uint8:
		var b uint8
		if err := binary.Read(buf, order, &b); err != nil { return err }
		v.SetUint(uint64(b))
	case reflect.Int16:
		var b int16
		if err := binary.Read(buf, order, &b); err != nil { return err }
		v.SetInt(int64(b))
	case reflect.Uint16:
		var b uint16
		if err := binary.Read(buf, order, &b); err != nil { return err }
		v.SetUint(uint64(b))
	case reflect.Int32:
		var b int32
		if err := binary.Read(buf, order, &b); err != nil { return err }
		v.SetInt(int64(b))
	case reflect.Uint32:
		var b uint32
		if err := binary.Read(buf, order, &b); err != nil { return err }
		v.SetUint(uint64(b))
	case reflect.Int64:
		var b int64
		if err := binary.Read(buf, order, &b); err != nil { return err }
		v.SetInt(b)
	case reflect.Uint64:
		var b uint64
		if err := binary.Read(buf, order, &b); err != nil { return err }
		v.SetUint(b)
	case reflect.Float32:
		var b float32
		if err := binary.Read(buf, order, &b); err != nil { return err }
		if math.IsNaN(float64(b)) { b = 0 }
		v.SetFloat(float64(b))
	case reflect.Float64:
		var b float64
		if err := binary.Read(buf, order, &b); err != nil { return err }
		if math.IsNaN(b) { b = 0 }
		v.SetFloat(b)
	case reflect.String:
		strBytes := make([]byte, tag.ByteCount)
		if err := binary.Read(buf, order, &strBytes); err != nil { return err }
		v.SetString(decodeFixedString(strBytes))
	}
	return nil
}
