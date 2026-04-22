package codec

import (
	"errors"
	"reflect"
	"sync"
)

type fieldOp int

type collectionOp int

const (
	fieldOpIgnore fieldOp = iota
	fieldOpFile
	fieldOpBitCount
	fieldOpRecursive
	fieldOpSliceOrArray
	fieldOpScalar
)

const (
	collectionOpNone collectionOp = iota
	collectionOpRecursive
	collectionOpSubBitCount
	collectionOpScalar
)

type collectionMeta struct {
	op collectionOp
}

type encodeTypeMeta struct {
	typ    reflect.Type
	fields []encodeFieldMeta
}

type decodeTypeMeta struct {
	typ    reflect.Type
	fields []decodeFieldMeta
}

type encodeFieldMeta struct {
	index         int
	tags          fieldTag
	op            fieldOp
	needsNilCheck bool
	collection    collectionMeta
}

type decodeFieldMeta struct {
	index        int
	tags         fieldTag
	op           fieldOp
	needsPtrInit bool
	collection   collectionMeta
}

var encodeTypeMetaCache sync.Map
var decodeTypeMetaCache sync.Map

func normalizeMetaType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func getEncodeTypeMeta(t reflect.Type) (*encodeTypeMeta, error) {
	t = normalizeMetaType(t)
	if cached, ok := encodeTypeMetaCache.Load(t); ok {
		return cached.(*encodeTypeMeta), nil
	}

	meta, err := buildEncodeTypeMeta(t)
	if err != nil {
		return nil, err
	}
	actual, _ := encodeTypeMetaCache.LoadOrStore(t, meta)
	return actual.(*encodeTypeMeta), nil
}

func getDecodeTypeMeta(t reflect.Type) (*decodeTypeMeta, error) {
	t = normalizeMetaType(t)
	if cached, ok := decodeTypeMetaCache.Load(t); ok {
		return cached.(*decodeTypeMeta), nil
	}

	meta, err := buildDecodeTypeMeta(t)
	if err != nil {
		return nil, err
	}
	actual, _ := decodeTypeMetaCache.LoadOrStore(t, meta)
	return actual.(*decodeTypeMeta), nil
}

func buildEncodeTypeMeta(t reflect.Type) (*encodeTypeMeta, error) {
	if t == nil || t.Kind() != reflect.Struct {
		return nil, errors.New("encode metadata requires struct type")
	}

	fields := make([]encodeFieldMeta, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		fieldMeta, err := buildEncodeFieldMeta(t.Field(i), i)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fieldMeta)
	}
	return &encodeTypeMeta{typ: t, fields: fields}, nil
}

func buildDecodeTypeMeta(t reflect.Type) (*decodeTypeMeta, error) {
	if t == nil || t.Kind() != reflect.Struct {
		return nil, errors.New("decode metadata requires struct type")
	}

	fields := make([]decodeFieldMeta, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		fieldMeta, err := buildDecodeFieldMeta(t.Field(i), i)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fieldMeta)
	}
	return &decodeTypeMeta{typ: t, fields: fields}, nil
}

func buildEncodeFieldMeta(field reflect.StructField, index int) (encodeFieldMeta, error) {
	tags, err := parseEncodeTag(field.Tag)
	if err != nil {
		return encodeFieldMeta{}, err
	}
	if tags.Ignore {
		return encodeFieldMeta{index: index, tags: *tags, op: fieldOpIgnore}, nil
	}
	if err := validateFieldTag(tags, field.Type); err != nil {
		return encodeFieldMeta{}, err
	}

	return encodeFieldMeta{
		index:         index,
		tags:          *tags,
		op:            classifyEncodeFieldOp(field.Type.Kind(), tags),
		needsNilCheck: needsEncodeNilCheck(field.Type.Kind()),
		collection:    buildEncodeCollectionMeta(field.Type, tags),
	}, nil
}

func buildDecodeFieldMeta(field reflect.StructField, index int) (decodeFieldMeta, error) {
	tags, err := parseDecodeTag(field.Tag)
	if err != nil {
		return decodeFieldMeta{}, err
	}
	if tags.Ignore {
		return decodeFieldMeta{index: index, tags: *tags, op: fieldOpIgnore}, nil
	}
	if err := validateFieldTag(tags, field.Type); err != nil {
		return decodeFieldMeta{}, err
	}

	return decodeFieldMeta{
		index:        index,
		tags:         *tags,
		op:           classifyDecodeFieldOp(field.Type.Kind(), tags),
		needsPtrInit: field.Type.Kind() == reflect.Ptr,
		collection:   buildDecodeCollectionMeta(field.Type, tags),
	}, nil
}

func classifyEncodeFieldOp(kind reflect.Kind, tags *fieldTag) fieldOp {
	switch {
	case tags.Ignore:
		return fieldOpIgnore
	case tags.File:
		return fieldOpFile
	case tags.BitCount != 0:
		return fieldOpBitCount
	case kind == reflect.Struct || kind == reflect.Ptr || kind == reflect.Interface:
		return fieldOpRecursive
	case kind == reflect.Array || kind == reflect.Slice:
		return fieldOpSliceOrArray
	default:
		return fieldOpScalar
	}
}

func classifyDecodeFieldOp(kind reflect.Kind, tags *fieldTag) fieldOp {
	switch {
	case tags.Ignore:
		return fieldOpIgnore
	case tags.File:
		return fieldOpFile
	case tags.BitCount != 0:
		return fieldOpBitCount
	case kind == reflect.Ptr || kind == reflect.Interface:
		return fieldOpRecursive
	case kind == reflect.Array || kind == reflect.Slice:
		return fieldOpSliceOrArray
	default:
		return fieldOpScalar
	}
}

func buildEncodeCollectionMeta(fieldType reflect.Type, tags *fieldTag) collectionMeta {
	if fieldType.Kind() != reflect.Array && fieldType.Kind() != reflect.Slice {
		return collectionMeta{}
	}

	elemKind := fieldType.Elem().Kind()
	switch {
	case elemKind == reflect.Struct || elemKind == reflect.Array || elemKind == reflect.Slice:
		return collectionMeta{op: collectionOpRecursive}
	case tags != nil && tags.SubBitCount != 0:
		return collectionMeta{op: collectionOpSubBitCount}
	default:
		return collectionMeta{op: collectionOpScalar}
	}
}

func buildDecodeCollectionMeta(fieldType reflect.Type, tags *fieldTag) collectionMeta {
	if fieldType.Kind() != reflect.Array && fieldType.Kind() != reflect.Slice {
		return collectionMeta{}
	}

	elemKind := fieldType.Elem().Kind()
	switch {
	case elemKind == reflect.Ptr || elemKind == reflect.Array:
		return collectionMeta{op: collectionOpRecursive}
	case tags != nil && tags.SubBitCount != 0:
		return collectionMeta{op: collectionOpSubBitCount}
	default:
		return collectionMeta{op: collectionOpScalar}
	}
}

func needsEncodeNilCheck(kind reflect.Kind) bool {
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return true
	default:
		return false
	}
}
