package codec

import (
	"errors"
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
