package transformation

import (
	"encoding/json"
	"testing"

	"github.com/twmb/franz-go/pkg/kgo"
)

func TestNewTransformer(t *testing.T) {
	rules := []*TransformationRule{
		{Type: TransformationMask, Field: "credit_card"},
	}

	transformer := NewTransformer(rules)

	if transformer == nil {
		t.Fatal("Expected non-nil transformer")
	}

	if len(transformer.rules) != 1 {
		t.Error("Expected 1 rule")
	}
}

func TestTransformer_MaskField(t *testing.T) {
	transformer := NewTransformer([]*TransformationRule{
		{Type: TransformationMask, Field: "credit_card"},
	})

	data := map[string]interface{}{
		"credit_card": "1234567890123456",
		"name":        "John Doe",
	}

	jsonData, _ := json.Marshal(data)
	record := &kgo.Record{Value: jsonData}

	transformed, err := transformer.Transform(record)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(transformed.Value, &result)

	maskedCard := result["credit_card"].(string)
	if maskedCard == "1234567890123456" {
		t.Error("Expected credit card to be masked")
	}

	// Should show last 4 digits
	if maskedCard[len(maskedCard)-4:] != "3456" {
		t.Error("Expected last 4 digits to be visible")
	}

	// Name should not be masked
	if result["name"] != "John Doe" {
		t.Error("Expected name to remain unchanged")
	}
}

func TestTransformer_RedactField(t *testing.T) {
	transformer := NewTransformer([]*TransformationRule{
		{Type: TransformationRedact, Field: "ssn"},
	})

	data := map[string]interface{}{
		"ssn":  "123-45-6789",
		"name": "John Doe",
	}

	jsonData, _ := json.Marshal(data)
	record := &kgo.Record{Value: jsonData}

	transformed, err := transformer.Transform(record)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(transformed.Value, &result)

	if _, exists := result["ssn"]; exists {
		t.Error("Expected SSN field to be removed")
	}

	if result["name"] != "John Doe" {
		t.Error("Expected name to remain")
	}
}

func TestTransformer_EnrichField(t *testing.T) {
	transformer := NewTransformer([]*TransformationRule{
		{Type: TransformationEnrich, Field: "timestamp", Value: "2025-11-05T12:00:00Z"},
	})

	data := map[string]interface{}{
		"id": "123",
	}

	jsonData, _ := json.Marshal(data)
	record := &kgo.Record{Value: jsonData}

	transformed, err := transformer.Transform(record)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(transformed.Value, &result)

	if result["timestamp"] != "2025-11-05T12:00:00Z" {
		t.Error("Expected timestamp to be added")
	}
}

func TestTransformer_HashField(t *testing.T) {
	transformer := NewTransformer([]*TransformationRule{
		{Type: TransformationHash, Field: "email"},
	})

	data := map[string]interface{}{
		"email": "user@example.com",
	}

	jsonData, _ := json.Marshal(data)
	record := &kgo.Record{Value: jsonData}

	transformed, err := transformer.Transform(record)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(transformed.Value, &result)

	hashedEmail := result["email"].(string)
	if hashedEmail == "user@example.com" {
		t.Error("Expected email to be hashed")
	}

	// Should be a hex string
	if len(hashedEmail) != 64 { // SHA256 produces 64 hex chars
		t.Errorf("Expected 64 character hash, got %d", len(hashedEmail))
	}
}

func TestTransformer_MultipleRules(t *testing.T) {
	transformer := NewTransformer([]*TransformationRule{
		{Type: TransformationMask, Field: "credit_card"},
		{Type: TransformationHash, Field: "email"},
		{Type: TransformationRedact, Field: "ssn"},
		{Type: TransformationEnrich, Field: "processed", Value: true},
	})

	data := map[string]interface{}{
		"credit_card": "1234567890123456",
		"email":       "user@example.com",
		"ssn":         "123-45-6789",
		"name":        "John Doe",
	}

	jsonData, _ := json.Marshal(data)
	record := &kgo.Record{Value: jsonData}

	transformed, err := transformer.Transform(record)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(transformed.Value, &result)

	// Verify all transformations applied
	if result["credit_card"] == "1234567890123456" {
		t.Error("Expected credit card to be masked")
	}

	if result["email"] == "user@example.com" {
		t.Error("Expected email to be hashed")
	}

	if _, exists := result["ssn"]; exists {
		t.Error("Expected SSN to be redacted")
	}

	if result["processed"] != true {
		t.Error("Expected processed field to be added")
	}

	if result["name"] != "John Doe" {
		t.Error("Expected name to remain unchanged")
	}
}

func TestTransformer_NonJSONData(t *testing.T) {
	transformer := NewTransformer([]*TransformationRule{
		{Type: TransformationMask, Field: "field"},
	})

	// Non-JSON data
	record := &kgo.Record{Value: []byte("not json")}

	transformed, err := transformer.Transform(record)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return original record for non-JSON
	if string(transformed.Value) != "not json" {
		t.Error("Expected non-JSON data to pass through unchanged")
	}
}

func TestNewFilter(t *testing.T) {
	rules := []*FilterRule{
		{Field: "status", Operator: "eq", Value: "active"},
	}

	filter := NewFilter(rules)

	if filter == nil {
		t.Fatal("Expected non-nil filter")
	}

	if len(filter.rules) != 1 {
		t.Error("Expected 1 rule")
	}
}

func TestFilter_ShouldFilter(t *testing.T) {
	filter := NewFilter([]*FilterRule{
		{Field: "status", Operator: "eq", Value: "active"},
	})

	t.Run("Matching message", func(t *testing.T) {
		data := map[string]interface{}{
			"status": "active",
		}
		jsonData, _ := json.Marshal(data)
		record := &kgo.Record{Value: jsonData}

		if filter.ShouldFilter(record) {
			t.Error("Expected message to pass filter")
		}
	})

	t.Run("Non-matching message", func(t *testing.T) {
		data := map[string]interface{}{
			"status": "inactive",
		}
		jsonData, _ := json.Marshal(data)
		record := &kgo.Record{Value: jsonData}

		if !filter.ShouldFilter(record) {
			t.Error("Expected message to be filtered")
		}
	})

	t.Run("Missing field", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "test",
		}
		jsonData, _ := json.Marshal(data)
		record := &kgo.Record{Value: jsonData}

		if !filter.ShouldFilter(record) {
			t.Error("Expected message with missing field to be filtered")
		}
	})
}

func TestFilter_Operators(t *testing.T) {
	tests := []struct {
		name       string
		operator   string
		value      interface{}
		testValue  interface{}
		shouldPass bool
	}{
		{"Equal - match", "eq", "test", "test", true},
		{"Equal - no match", "eq", "test", "other", false},
		{"Not equal - match", "ne", "test", "other", true},
		{"Not equal - no match", "ne", "test", "test", false},
		{"Contains - match", "contains", "world", "hello world", true},
		{"Contains - no match", "contains", "world", "hello", false},
		{"Regex - match", "regex", "^test.*", "test123", true},
		{"Regex - no match", "regex", "^test.*", "abc123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter([]*FilterRule{
				{Field: "field", Operator: tt.operator, Value: tt.value},
			})

			data := map[string]interface{}{
				"field": tt.testValue,
			}
			jsonData, _ := json.Marshal(data)
			record := &kgo.Record{Value: jsonData}

			shouldFilter := filter.ShouldFilter(record)
			shouldPass := !shouldFilter

			if shouldPass != tt.shouldPass {
				t.Errorf("Expected shouldPass=%v, got %v", tt.shouldPass, shouldPass)
			}
		})
	}
}

func TestFilter_NoRules(t *testing.T) {
	filter := NewFilter([]*FilterRule{})

	data := map[string]interface{}{
		"status": "active",
	}
	jsonData, _ := json.Marshal(data)
	record := &kgo.Record{Value: jsonData}

	if filter.ShouldFilter(record) {
		t.Error("Expected message to pass with no filter rules")
	}
}

func TestFilter_InvalidJSON(t *testing.T) {
	filter := NewFilter([]*FilterRule{
		{Field: "status", Operator: "eq", Value: "active"},
	})

	record := &kgo.Record{Value: []byte("not json")}

	if !filter.ShouldFilter(record) {
		t.Error("Expected invalid JSON to be filtered")
	}
}
