package serialization

import (
	"testing"

	"github.com/twmb/franz-go/pkg/sr"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		contentType    string
		expectedFormat Format
	}{
		{"application/json", FormatJSON},
		{"application/avro", FormatAvro},
		{"application/vnd.apache.avro+binary", FormatAvro},
		{"application/protobuf", FormatProtobuf},
		{"application/x-protobuf", FormatProtobuf},
		{"", FormatAvro}, // default
		{"text/plain", FormatAvro}, // unknown defaults to Avro
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			format := DetectFormat(tt.contentType)
			if format != tt.expectedFormat {
				t.Errorf("Expected format %s for content-type %s, got %s",
					tt.expectedFormat, tt.contentType, format)
			}
		})
	}
}

func TestNewJSONSerializer(t *testing.T) {
	serializer := NewJSONSerializer()

	if serializer == nil {
		t.Fatal("Expected non-nil serializer")
	}

	if serializer.GetFormat() != FormatJSON {
		t.Error("Expected JSON format")
	}
}

func TestJSONSerializer_Serialize(t *testing.T) {
	serializer := NewJSONSerializer()

	data := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}

	serialized, err := serializer.Serialize(data, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(serialized) == 0 {
		t.Error("Expected non-empty serialized data")
	}

	// Check it's valid JSON
	expected := `{"name":"test","value":123}`
	if string(serialized) != expected {
		t.Errorf("Expected %s, got %s", expected, string(serialized))
	}
}

func TestJSONSerializer_Deserialize(t *testing.T) {
	serializer := NewJSONSerializer()

	jsonData := []byte(`{"name":"test","value":123}`)

	var target map[string]interface{}
	err := serializer.Deserialize(jsonData, nil, &target)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if target["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", target["name"])
	}

	if target["value"].(float64) != 123 {
		t.Errorf("Expected value 123, got %v", target["value"])
	}
}

func TestJSONSerializer_RoundTrip(t *testing.T) {
	serializer := NewJSONSerializer()

	original := map[string]interface{}{
		"id":     "abc123",
		"amount": 100.50,
		"active": true,
	}

	// Serialize
	serialized, err := serializer.Serialize(original, nil)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	// Deserialize
	var decoded map[string]interface{}
	err = serializer.Deserialize(serialized, nil, &decoded)
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}

	// Verify
	if decoded["id"] != original["id"] {
		t.Error("ID mismatch after round trip")
	}

	if decoded["amount"] != original["amount"] {
		t.Error("Amount mismatch after round trip")
	}

	if decoded["active"] != original["active"] {
		t.Error("Active mismatch after round trip")
	}
}

func TestSerializerFactory_GetSerializer(t *testing.T) {
	factory := NewSerializerFactory(nil)

	t.Run("JSON Serializer", func(t *testing.T) {
		serializer, err := factory.GetSerializer(FormatJSON)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if serializer == nil {
			t.Fatal("Expected non-nil serializer")
		}

		if serializer.GetFormat() != FormatJSON {
			t.Error("Expected JSON format")
		}
	})

	t.Run("Avro Serializer", func(t *testing.T) {
		serializer, err := factory.GetSerializer(FormatAvro)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if serializer == nil {
			t.Fatal("Expected non-nil serializer")
		}

		if serializer.GetFormat() != FormatAvro {
			t.Error("Expected Avro format")
		}
	})

	t.Run("Protobuf Serializer (Not Implemented)", func(t *testing.T) {
		serializer, err := factory.GetSerializer(FormatProtobuf)
		if err == nil {
			t.Error("Expected error for unimplemented Protobuf")
		}

		if serializer != nil {
			t.Error("Expected nil serializer for unimplemented format")
		}
	})

	t.Run("Unsupported Format", func(t *testing.T) {
		serializer, err := factory.GetSerializer(Format("unsupported"))
		if err == nil {
			t.Error("Expected error for unsupported format")
		}

		if serializer != nil {
			t.Error("Expected nil serializer for unsupported format")
		}
	})
}

func TestNewAvroSerializer(t *testing.T) {
	// Create a mock schema registry client (nil is okay for unit tests)
	serializer := NewAvroSerializer(nil)

	if serializer == nil {
		t.Fatal("Expected non-nil serializer")
	}

	if serializer.GetFormat() != FormatAvro {
		t.Error("Expected Avro format")
	}
}

func TestNewSerializerFactory(t *testing.T) {
	var client *sr.Client
	factory := NewSerializerFactory(client)

	if factory == nil {
		t.Fatal("Expected non-nil factory")
	}

	if factory.jsonSerializer == nil {
		t.Error("Expected JSON serializer to be initialized")
	}

	if factory.avroSerializer == nil {
		t.Error("Expected Avro serializer to be initialized")
	}
}

func TestFormat_String(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{FormatAvro, "avro"},
		{FormatJSON, "json"},
		{FormatProtobuf, "protobuf"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.format) != tt.expected {
				t.Errorf("Expected format string %s, got %s", tt.expected, string(tt.format))
			}
		})
	}
}
