package parser_test

import (
	"errors"
	"testing"

	"github.com/gurch101/gowebutils/pkg/parser"
)

func TestParseJSONMapInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		key     string
		want    int64
		wantErr error
	}{
		{
			name:    "valid int64",
			input:   map[string]any{"count": float64(42)},
			key:     "count",
			want:    42,
			wantErr: nil,
		},
		{
			name:    "invalid type",
			input:   map[string]any{"count": "not a number"},
			key:     "count",
			want:    0,
			wantErr: parser.ErrInvalidMapKey,
		},
		{
			name:    "missing key",
			input:   map[string]any{},
			key:     "count",
			want:    0,
			wantErr: parser.ErrInvalidMapKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseJSONMapInt64(tt.input, tt.key)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseJSONMapInt64() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("ParseJSONMapInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseJSONMapString(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		key     string
		want    string
		wantErr error
	}{
		{
			name:    "valid string",
			input:   map[string]any{"name": "test"},
			key:     "name",
			want:    "test",
			wantErr: nil,
		},
		{
			name:    "invalid type",
			input:   map[string]any{"name": 42},
			key:     "name",
			want:    "",
			wantErr: parser.ErrInvalidMapKey,
		},
		{
			name:    "missing key",
			input:   map[string]any{},
			key:     "name",
			want:    "",
			wantErr: parser.ErrInvalidMapKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseJSONMapString(tt.input, tt.key)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseJSONMapString() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("ParseJSONMapString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructsToFilteredMaps(t *testing.T) {
	t.Parallel()

	// Example usage
	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	}

	type Metadata struct {
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		Tags      []string
	}

	type Person struct {
		Name     string   `json:"name"`
		Age      int      `json:"age"`
		Address  Address  `json:"address"`
		Phone    string   `json:"phone"`
		Metadata Metadata `json:"meta"`
		Friends  []Person `json:"friends"`
	}

	// Setup test data
	people := []Person{
		{
			Name: "John Doe",
			Age:  30,
			Address: Address{
				Street:  "123 Main St",
				City:    "New York",
				Country: "USA",
			},
			Phone: "555-1234",
			Metadata: Metadata{
				CreatedAt: "2023-01-01",
				UpdatedAt: "2023-05-15",
				Tags:      []string{"vip", "developer"},
			},
			//nolint: exhaustruct
			Friends: []Person{
				{
					Name: "Jane Smith",
					Age:  28,
					Address: Address{
						Street:  "456 Oak Ave",
						City:    "Boston",
						Country: "USA",
					},
				},
			},
		},
	}

	includeFields := []string{
		"name",
		"address.city",
		"meta.createdAt",
		"friends.name",
		"friends.address.street",
	}

	// Execute function
	result, err := parser.StructsToFilteredMaps(people, includeFields)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the result structure
	if len(result) != 1 {
		t.Fatalf("Expected 1 person in result, got %d", len(result))
	}

	personMap := result[0]

	verifyTopLevelFields(t, personMap)
	verifyAddressFields(t, personMap)
	verifyMetadataFields(t, personMap)
	verifyFriendsFields(t, personMap)
}

func verifyTopLevelFields(t *testing.T, personMap map[string]interface{}) {
	t.Helper()

	checkFieldExists(t, personMap, "name")
	checkFieldExists(t, personMap, "address")
	checkFieldExists(t, personMap, "meta")
	checkFieldExists(t, personMap, "friends")
	checkFieldMissing(t, personMap, "age") // Should be filtered out
	checkFieldMissing(t, personMap, "phone")

	if name, ok := personMap["name"].(string); !ok || name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %v", personMap["name"])
	}
}

func verifyAddressFields(t *testing.T, personMap map[string]interface{}) {
	t.Helper()

	address, ok := personMap["address"].(map[string]interface{})
	if !ok {
		t.Fatal("address field is not a map")
	}

	if city, ok := address["city"].(string); !ok || city != "New York" {
		t.Errorf("Expected address.city 'New York', got %v", address["city"])
	}

	checkFieldMissing(t, address, "street")
	checkFieldMissing(t, address, "country")
}

func verifyMetadataFields(t *testing.T, personMap map[string]interface{}) {
	t.Helper()

	meta, ok := personMap["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta field is not a map")
	}

	if createdAt, ok := meta["createdAt"].(string); !ok || createdAt != "2023-01-01" {
		t.Errorf("Expected meta.createdAt '2023-01-01', got %v", meta["createdAt"])
	}

	checkFieldMissing(t, meta, "updatedAt")
	checkFieldMissing(t, meta, "tags")
}

func verifyFriendsFields(t *testing.T, personMap map[string]interface{}) {
	t.Helper()

	friends, ok := personMap["friends"].([]interface{})
	if !ok {
		t.Fatal("friends field is not a slice")
	}

	if len(friends) != 1 {
		t.Fatalf("Expected 1 friend, got %d", len(friends))
	}

	friend, ok := friends[0].(map[string]interface{})
	if !ok {
		t.Fatal("friend is not a map")
	}

	if friendName, ok := friend["name"].(string); !ok || friendName != "Jane Smith" {
		t.Errorf("Expected friend.name 'Jane Smith', got %v", friend["name"])
	}

	friendAddress, ok := friend["address"].(map[string]interface{})
	if !ok {
		t.Fatal("friend.address is not a map")
	}

	if street, ok := friendAddress["street"].(string); !ok || street != "456 Oak Ave" {
		t.Errorf("Expected friend.address.street '456 Oak Ave', got %v", friendAddress["street"])
	}

	checkFieldMissing(t, friendAddress, "city")
	checkFieldMissing(t, friend, "age")
}

// Helper functions for testing.
func checkFieldExists(t *testing.T, m map[string]interface{}, field string) {
	t.Helper()

	if _, exists := m[field]; !exists {
		t.Errorf("Expected field %s to exist", field)
	}
}

func checkFieldMissing(t *testing.T, m map[string]interface{}, field string) {
	t.Helper()

	if _, exists := m[field]; exists {
		t.Errorf("Expected field %s to be missing", field)
	}
}
