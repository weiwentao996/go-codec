package codec

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var tagIntPattern = regexp.MustCompile("[0-9]+")

type fieldTag struct {
	Ignore      bool
	BitCount    int
	SubBitCount int
	ByteCount   int
	File        bool
}

func parseEncodeTag(tag reflect.StructTag) (*fieldTag, error) {
	return parseTag(tag, "encode")
}

func parseDecodeTag(tag reflect.StructTag) (*fieldTag, error) {
	return parseTag(tag, "decode")
}

func parseTag(tag reflect.StructTag, tagName string) (*fieldTag, error) {
	result := fieldTag{}
	for _, value := range strings.Split(tag.Get(tagName), ",") {
		switch {
		case value == "-":
			result.Ignore = true
		case strings.Contains(value, "bitCount:"):
			bitCount, err := parseTagInt(value)
			if err != nil {
				return nil, err
			}
			result.BitCount = bitCount
		case strings.Contains(value, "subBitCount:"):
			subBitCount, err := parseTagInt(value)
			if err != nil {
				return nil, err
			}
			result.SubBitCount = subBitCount
		case strings.Contains(value, "byteCount:"):
			byteCount, err := parseTagInt(value)
			if err != nil {
				return nil, err
			}
			result.ByteCount = byteCount
		case strings.Contains(value, "file"):
			result.File = true
		}
	}
	return &result, nil
}

func parseTagInt(tagValue string) (int, error) {
	matched := tagIntPattern.FindString(tagValue)
	if matched == "" {
		return 0, errors.New("parse tag fail")
	}
	value, err := strconv.Atoi(matched)
	if err != nil {
		return 0, errors.New("parse tag fail")
	}
	return value, nil
}

func validateFieldTag(tags *fieldTag, fieldType reflect.Type) error {
	if tags == nil || tags.Ignore {
		return nil
	}

	kind := fieldType.Kind()
	if kind == reflect.Pointer {
		kind = fieldType.Elem().Kind()
	}

	if tags.File && kind != reflect.String {
		return fmt.Errorf("file tag requires string field, got %s", fieldType.Kind())
	}
	if tags.BitCount != 0 && !isUnsignedIntegerKind(kind) {
		return fmt.Errorf("bitCount tag requires unsigned integer field, got %s", fieldType.Kind())
	}
	if tags.SubBitCount != 0 {
		if kind != reflect.Array && kind != reflect.Slice {
			return fmt.Errorf("subBitCount tag requires array or slice field, got %s", fieldType.Kind())
		}
		elemKind := fieldType.Elem().Kind()
		if !isUnsignedIntegerKind(elemKind) {
			return fmt.Errorf("subBitCount tag requires unsigned integer elements, got %s", fieldType.Elem().Kind())
		}
	}
	if tags.ByteCount != 0 && kind != reflect.String {
		return fmt.Errorf("byteCount tag requires string field, got %s", fieldType.Kind())
	}
	return nil
}

func isUnsignedIntegerKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return true
	default:
		return false
	}
}
