// Package transformation provides message filtering and transformation capabilities
package transformation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
)

// TransformationType represents the type of transformation
type TransformationType string

const (
	TransformationMask      TransformationType = "mask"
	TransformationRedact    TransformationType = "redact"
	TransformationEnrich    TransformationType = "enrich"
	TransformationHash      TransformationType = "hash"
	TransformationFilter    TransformationType = "filter"
)

// TransformationRule defines a transformation rule
type TransformationRule struct {
	Type       TransformationType     `json:"type"`
	Field      string                 `json:"field"`
	Pattern    string                 `json:"pattern"`
	Replacement string                `json:"replacement"`
	Condition  string                 `json:"condition"`
	Value      interface{}            `json:"value"`
	Options    map[string]interface{} `json:"options"`
}

// Transformer handles message transformation
type Transformer struct {
	rules []*TransformationRule
}

// NewTransformer creates a new transformer
func NewTransformer(rules []*TransformationRule) *Transformer {
	return &Transformer{
		rules: rules,
	}
}

// Transform applies transformation rules to a message
func (t *Transformer) Transform(record *kgo.Record) (*kgo.Record, error) {
	// Parse the message value as JSON
	var data map[string]interface{}
	if err := json.Unmarshal(record.Value, &data); err != nil {
		// If not JSON, return original record
		return record, nil
	}

	// Apply each transformation rule
	for _, rule := range t.rules {
		if err := t.applyRule(data, rule); err != nil {
			return nil, fmt.Errorf("failed to apply transformation rule: %w", err)
		}
	}

	// Serialize back to JSON
	transformed, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transformed data: %w", err)
	}

	// Create new record with transformed data
	transformedRecord := &kgo.Record{
		Topic:     record.Topic,
		Partition: record.Partition,
		Key:       record.Key,
		Value:     transformed,
		Headers:   record.Headers,
	}

	// Add transformation metadata header
	transformedRecord.Headers = append(transformedRecord.Headers, kgo.RecordHeader{
		Key:   "x-transformed",
		Value: []byte("true"),
	})

	return transformedRecord, nil
}

// applyRule applies a single transformation rule
func (t *Transformer) applyRule(data map[string]interface{}, rule *TransformationRule) error {
	switch rule.Type {
	case TransformationMask:
		return t.maskField(data, rule)
	case TransformationRedact:
		return t.redactField(data, rule)
	case TransformationEnrich:
		return t.enrichField(data, rule)
	case TransformationHash:
		return t.hashField(data, rule)
	case TransformationFilter:
		return t.filterField(data, rule)
	default:
		return fmt.Errorf("unsupported transformation type: %s", rule.Type)
	}
}

// maskField masks sensitive data in a field
func (t *Transformer) maskField(data map[string]interface{}, rule *TransformationRule) error {
	value, exists := data[rule.Field]
	if !exists {
		return nil
	}

	strValue, ok := value.(string)
	if !ok {
		return nil
	}

	// Mask all but last 4 characters
	if len(strValue) > 4 {
		masked := strings.Repeat("*", len(strValue)-4) + strValue[len(strValue)-4:]
		data[rule.Field] = masked
	} else {
		data[rule.Field] = strings.Repeat("*", len(strValue))
	}

	return nil
}

// redactField removes a field completely
func (t *Transformer) redactField(data map[string]interface{}, rule *TransformationRule) error {
	delete(data, rule.Field)
	return nil
}

// enrichField adds or updates a field with a value
func (t *Transformer) enrichField(data map[string]interface{}, rule *TransformationRule) error {
	data[rule.Field] = rule.Value
	return nil
}

// hashField hashes a field value
func (t *Transformer) hashField(data map[string]interface{}, rule *TransformationRule) error {
	value, exists := data[rule.Field]
	if !exists {
		return nil
	}

	strValue := fmt.Sprintf("%v", value)
	hash := sha256.Sum256([]byte(strValue))
	data[rule.Field] = hex.EncodeToString(hash[:])

	return nil
}

// filterField filters based on a condition
func (t *Transformer) filterField(data map[string]interface{}, rule *TransformationRule) error {
	// Pattern-based filtering
	if rule.Pattern != "" {
		value, exists := data[rule.Field]
		if !exists {
			return nil
		}

		strValue := fmt.Sprintf("%v", value)
		matched, err := regexp.MatchString(rule.Pattern, strValue)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}

		if !matched {
			return fmt.Errorf("filter condition not met")
		}
	}

	return nil
}

// Filter checks if a message should be filtered based on rules
type Filter struct {
	rules []*FilterRule
}

// FilterRule defines a filtering rule
type FilterRule struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, regex
	Value    interface{} `json:"value"`
}

// NewFilter creates a new filter
func NewFilter(rules []*FilterRule) *Filter {
	return &Filter{
		rules: rules,
	}
}

// ShouldFilter checks if a message should be filtered
func (f *Filter) ShouldFilter(record *kgo.Record) bool {
	if len(f.rules) == 0 {
		return false // No rules means pass all messages
	}

	// Parse the message value as JSON
	var data map[string]interface{}
	if err := json.Unmarshal(record.Value, &data); err != nil {
		return true // Filter out invalid JSON
	}

	// Check all filter rules (AND logic)
	for _, rule := range f.rules {
		if !f.matchesRule(data, rule) {
			return true // Filter out if any rule doesn't match
		}
	}

	return false // Pass the message
}

// matchesRule checks if data matches a single filter rule
func (f *Filter) matchesRule(data map[string]interface{}, rule *FilterRule) bool {
	value, exists := data[rule.Field]
	if !exists {
		return false
	}

	switch rule.Operator {
	case "eq":
		return value == rule.Value
	case "ne":
		return value != rule.Value
	case "contains":
		strValue := fmt.Sprintf("%v", value)
		strRuleValue := fmt.Sprintf("%v", rule.Value)
		return strings.Contains(strValue, strRuleValue)
	case "regex":
		strValue := fmt.Sprintf("%v", value)
		strPattern := fmt.Sprintf("%v", rule.Value)
		matched, err := regexp.MatchString(strPattern, strValue)
		return err == nil && matched
	default:
		return false
	}
}
