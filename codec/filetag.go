package codec

func readFile(path string, cfg Config) ([]byte, error) {
	if cfg.FileReader == nil {
		return nil, ErrFileReaderNotConfigured
	}
	return cfg.FileReader(path)
}

func writeFile(typeName string, data []byte, cfg Config) (string, error) {
	if cfg.FileWriter == nil {
		return "", ErrFileWriterNotConfigured
	}
	return cfg.FileWriter(typeName, data)
}
