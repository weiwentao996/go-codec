# go-code

A reusable binary codec extracted from `hkim-data-transfer/lib/data`.

## Scope

This repository currently contains only the core codec layer:

- reflection-based marshal / unmarshal
- `encode` / `decode` struct tags
- configurable scalar byte order
- configurable bitfield layout
- fixed-length string handling
- optional file tag hooks

It intentionally does **not** include protocol framing, command mapping, MQ wrappers, or project-specific storage behavior.

## API

```go
b, err := codec.Marshal(v)
err := codec.Unmarshal(b, &v)
```

## Configuration

### Legacy compatibility preset

```go
b, err := codec.Marshal(v, codec.WithLegacyHKimPreset())
err := codec.Unmarshal(b, &v, codec.WithLegacyHKimPreset())
```

This preset keeps the historical behavior from `hkim-data-transfer`:

- `ByteOrder = BigEndian`
- `BitLayout = LegacySameDevice`

### Byte order

Supported global byte-order modes:

- `codec.BigEndian`
- `codec.LittleEndian`

Example:

```go
b, err := codec.Marshal(v, codec.WithByteOrder(codec.LittleEndian))
err := codec.Unmarshal(b, &v, codec.WithByteOrder(codec.LittleEndian))
```

Notes:

- byte order affects multi-byte scalar values such as `uint16`, `uint32`, `uint64`, `float32`, and `float64`
- `uint8`, `int8`, and `bool` are effectively unaffected by endian choice

### Bitfield layout

Supported global bitfield layouts:

- `codec.LegacySameDevice`
- `codec.LSBFirstLowToHigh`

Example:

```go
b, err := codec.Marshal(v, codec.WithBitLayout(codec.LSBFirstLowToHigh))
err := codec.Unmarshal(b, &v, codec.WithBitLayout(codec.LSBFirstLowToHigh))
```

#### `LegacySameDevice`

This is the compatibility mode used by the old project.

Behavior summary:

- bitfields use the existing historical packing order
- current protocol compatibility is preserved
- default configuration uses this mode

#### `LSBFirstLowToHigh`

This is an alternative layout for testing and future protocol support.

Behavior summary:

- lower bits are emitted first
- bits fill each byte from low bit to high bit
- encoded bytes differ from `LegacySameDevice`, so encode/decode must use the same layout

## Tags

Currently supported tags:

- `encode:"-"` / `decode:"-"`
- `bitCount:N`
- `subBitCount:N`
- `byteCount:N`
- `file`

Examples:

```go
type Packet struct {
    Flags [8]uint8 `encode:"subBitCount:1" decode:"subBitCount:1"`
    Name  string   `encode:"byteCount:8" decode:"byteCount:8"`
    A     uint8    `encode:"bitCount:3" decode:"bitCount:3"`
}
```

## File tag hooks

The core library does not perform project-specific file IO by itself.
If you use `file` tags, you must provide hooks.

### Encode-side file reader

```go
b, err := codec.Marshal(v, codec.WithFileReader(func(path string) ([]byte, error) {
    return os.ReadFile(path)
}))
```

### Decode-side file writer

```go
err := codec.Unmarshal(data, &v, codec.WithFileWriter(func(typeName string, data []byte) (string, error) {
    path := filepath.Join("out", typeName)
    return path, os.WriteFile(path, data, 0o644)
}))
```

If hooks are not configured:

- encode with `file` tag returns `ErrFileReaderNotConfigured`
- decode with `file` tag returns `ErrFileWriterNotConfigured`

## Compatibility notes

The default configuration preserves the current legacy behavior from `hkim-data-transfer`:

- big-endian scalar encoding / decoding
- current bitfield order (`LegacySameDevice`)
- fixed-length string zero-padding on encode
- fixed-length string zero-to-space + trim behavior on decode
- pointer / interface decode compatibility

## Status

Current phase focuses on:

- behavior compatibility
- explicit configuration for byte order and bitfield layout
- reusable core extraction

Not included yet:

- field-level endian override
- field-level bit layout override
- protocol header/body/frame abstractions
- command registries and transport adapters
