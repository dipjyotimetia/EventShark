package avro

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/linkedin/goavro/v2"
)

// Package avro provides functions to encode and decode messages using the Avro serialization format.

// EncodeAvroMessage reads an Avro schema from a file and encodes a sample Customer object
// into a binary Avro message using the provided schema.
// It returns the encoded binary Avro message.
func EncodeAvroMessage(schemaFilePath string, message any) ([]byte, error) {
	pwd, _ := os.Getwd()
	// Read the Avro schema from the file.
	schemaBytes, err := os.ReadFile(pwd + schemaFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Avro schema file: %v", err)
	}

	codec, err := goavro.NewCodec(string(schemaBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Avro codec: %v", err)
	}

	dat, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to customer: %v", err)
	}

	native, _, err := codec.NativeFromTextual(dat)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to native: %v", err)
	}

	// Convert native Go form to binary Avro data
	binary, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, fmt.Errorf("failed to convert native to binary: %v", err)
	}
	return binary, nil
}

// DecodeAvroMessage reads an Avro schema from a file and decodes a binary Avro message
// using the provided schema. It returns a pointer to a Customer object.
func DecodeAvroMessage(schemaFilePath string, binary []byte) ([]byte, error) {
	pwd, _ := os.Getwd()
	// Read the Avro schema from the file.
	schemaBytes, err := os.ReadFile(pwd + schemaFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Avro schema file: %v", err)
	}

	// Create an Avro codec from the schema.
	codec, err := goavro.NewCodec(string(schemaBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Avro codec: %v", err)
	}

	// Decode the Avro message.
	native, _, err := codec.NativeFromBinary(binary)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Avro message: %v", err)
	}

	// Convert native Go form to textual Avro data
	textual, err := codec.TextualFromNative(nil, native)
	if err != nil {
		return nil, fmt.Errorf("failed to convert native to textual: %v", err)
	}

	return textual, nil
}
