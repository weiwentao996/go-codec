package codec

import (
	"encoding/binary"
	"errors"
)

type ByteOrderMode int

type BitLayoutMode int

type Option func(*Config)

type Config struct {
	ByteOrder  ByteOrderMode
	BitLayout  BitLayoutMode
	FileReader func(path string) ([]byte, error)
	FileWriter func(typeName string, data []byte) (string, error)
}

const (
	BigEndian ByteOrderMode = iota
	LittleEndian
)

const (
	LegacySameDevice BitLayoutMode = iota
	LSBFirstLowToHigh
)

var ErrFileReaderNotConfigured = errors.New("codec file reader is not configured")
var ErrFileWriterNotConfigured = errors.New("codec file writer is not configured")

func defaultConfig() Config {
	return Config{
		ByteOrder: BigEndian,
		BitLayout: LegacySameDevice,
	}
}

func (cfg Config) byteOrder() binary.ByteOrder {
	if cfg.ByteOrder == LittleEndian {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func WithByteOrder(order ByteOrderMode) Option {
	return func(cfg *Config) {
		cfg.ByteOrder = order
	}
}

func WithBitLayout(layout BitLayoutMode) Option {
	return func(cfg *Config) {
		cfg.BitLayout = layout
	}
}

func WithFileReader(reader func(path string) ([]byte, error)) Option {
	return func(cfg *Config) {
		cfg.FileReader = reader
	}
}

func WithFileWriter(writer func(typeName string, data []byte) (string, error)) Option {
	return func(cfg *Config) {
		cfg.FileWriter = writer
	}
}

func WithLegacyHKimPreset() Option {
	return func(cfg *Config) {
		cfg.ByteOrder = BigEndian
		cfg.BitLayout = LegacySameDevice
	}
}

func buildConfig(opts ...Option) Config {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}
