// Package serialization provides multi-format serialization support
package serialization

import (
	"fmt"

	"github.com/dipjyotimetia/event-shark/pkg/errors"
	"github.com/goccy/go-json"
	"github.com/hamba/avro/v2"
	"github.com/twmb/franz-go/pkg/sr"
)

// Format represents serialization format
type Format string

const (
	FormatAvro     Format = "avro"
	FormatJSON     Format = "json"
	FormatProtobuf Format = "protobuf"
)

// Serializer interface for different serialization formats
type Serializer interface {
	Serialize(data interface{}, schema interface{}) ([]byte, error)
	Deserialize(data []byte, schema interface{}, target interface{}) error
	GetFormat() Format
}

// AvroSerializer implements Avro serialization
type AvroSerializer struct {
	schemaRegistry *sr.Client
}

// NewAvroSerializer creates a new Avro serializer
func NewAvroSerializer(schemaRegistry *sr.Client) *AvroSerializer {
	return &AvroSerializer{
		schemaRegistry: schemaRegistry,
	}
}

// Serialize serializes data using Avro
func (s *AvroSerializer) Serialize(data interface{}, schema interface{}) ([]byte, error) {
	avroSchema, ok := schema.(avro.Schema)
	if !ok {
		return nil, errors.NewSerializationError("avro", fmt.Errorf("invalid Avro schema type"))
	}

	serialized, err := avro.Marshal(avroSchema, data)
	if err != nil {
		return nil, errors.NewSerializationError("avro", err)
	}

	return serialized, nil
}

// Deserialize deserializes Avro data
func (s *AvroSerializer) Deserialize(data []byte, schema interface{}, target interface{}) error {
	avroSchema, ok := schema.(avro.Schema)
	if !ok {
		return errors.NewSerializationError("avro", fmt.Errorf("invalid Avro schema type"))
	}

	if err := avro.Unmarshal(avroSchema, data, target); err != nil {
		return errors.NewSerializationError("avro", err)
	}

	return nil
}

// GetFormat returns the serialization format
func (s *AvroSerializer) GetFormat() Format {
	return FormatAvro
}

// JSONSerializer implements JSON serialization
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSON serializer
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Serialize serializes data using JSON
func (s *JSONSerializer) Serialize(data interface{}, schema interface{}) ([]byte, error) {
	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, errors.NewSerializationError("json", err)
	}

	return serialized, nil
}

// Deserialize deserializes JSON data
func (s *JSONSerializer) Deserialize(data []byte, schema interface{}, target interface{}) error {
	if err := json.Unmarshal(data, target); err != nil {
		return errors.NewSerializationError("json", err)
	}

	return nil
}

// GetFormat returns the serialization format
func (s *JSONSerializer) GetFormat() Format {
	return FormatJSON
}

// SerializerFactory creates serializers based on format
type SerializerFactory struct {
	avroSerializer *AvroSerializer
	jsonSerializer *JSONSerializer
}

// NewSerializerFactory creates a new serializer factory
func NewSerializerFactory(schemaRegistry *sr.Client) *SerializerFactory {
	return &SerializerFactory{
		avroSerializer: NewAvroSerializer(schemaRegistry),
		jsonSerializer: NewJSONSerializer(),
	}
}

// GetSerializer returns a serializer for the given format
func (f *SerializerFactory) GetSerializer(format Format) (Serializer, error) {
	switch format {
	case FormatAvro:
		return f.avroSerializer, nil
	case FormatJSON:
		return f.jsonSerializer, nil
	case FormatProtobuf:
		return nil, fmt.Errorf("protobuf serialization not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported serialization format: %s", format)
	}
}

// DetectFormat detects the serialization format from content type
func DetectFormat(contentType string) Format {
	switch contentType {
	case "application/json":
		return FormatJSON
	case "application/avro", "application/vnd.apache.avro+binary":
		return FormatAvro
	case "application/protobuf", "application/x-protobuf":
		return FormatProtobuf
	default:
		return FormatAvro // default
	}
}
