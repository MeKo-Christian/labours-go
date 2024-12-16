package readers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// DetectAndReadInput detects the format (if "auto"), creates the appropriate Reader, and reads the input.
func DetectAndReadInput(input string, format string) (Reader, error) {
	// Open input source
	var file io.Reader
	if input == "-" {
		file = os.Stdin
	} else {
		f, err := os.Open(input)
		if err != nil {
			return nil, fmt.Errorf("error opening file %s: %v", input, err)
		}
		defer f.Close()
		file = f
	}

	// Detect format if set to "auto"
	if format == "auto" {
		var err error
		format, file, err = detectFormat(file)
		if err != nil {
			return nil, err
		}
	}

	// Create the appropriate Reader
	reader, err := createReader(format)
	if err != nil {
		return nil, err
	}

	// Read the input using the Reader
	if err := reader.Read(file); err != nil {
		return nil, fmt.Errorf("error reading input with %s reader: %v", format, err)
	}

	return reader, nil
}

// detectFormat inspects the input to determine the format (YAML or Protobuf).
func detectFormat(file io.Reader) (string, io.Reader, error) {
	buffer := make([]byte, 16)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", nil, fmt.Errorf("error reading input for format detection: %v", err)
	}

	// Rewind the file for further reading
	file = io.MultiReader(bytes.NewReader(buffer), file)

	if isYAML(buffer) {
		return "yaml", file, nil
	}
	return "pb", file, nil
}

// isYAML checks if the buffer contains YAML-specific patterns.
func isYAML(buffer []byte) bool {
	return bytes.Contains(buffer, []byte("hercules"))
}

// createReader initializes the correct Reader implementation.
func createReader(format string) (Reader, error) {
	switch strings.ToLower(format) {
	case "yaml":
		return &YamlReader{}, nil
	case "pb":
		return &ProtobufReader{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
