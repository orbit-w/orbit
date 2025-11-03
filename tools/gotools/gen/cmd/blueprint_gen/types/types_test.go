package blueprint_types

import (
	"testing"
)

func TestParseTypeString_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKind FieldKind
		wantErr  bool
	}{
		{"int32", "int32", FieldKindInt32, false},
		{"int64", "int64", FieldKindInt64, false},
		{"string", "string", FieldKindString, false},
		{"bool", "bool", FieldKindBool, false},
		{"float", "float", FieldKindFloat, false},
		{"double", "double", FieldKindDouble, false},
		{"bytes", "bytes", FieldKindBytes, false},
		{"uint32", "uint32", FieldKindUInt32, false},
		{"uint64", "uint64", FieldKindUInt64, false},
		{"fixed32", "fixed32", FieldKindFixed32, false},
		{"fixed64", "fixed64", FieldKindFixed64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTypeString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Errorf("ParseTypeString() returned nil")
					return
				}
				if got.Kind != tt.wantKind {
					t.Errorf("ParseTypeString() Kind = %v, want %v", got.Kind, tt.wantKind)
				}
				if got.Label != FieldLabelOptional {
					t.Errorf("ParseTypeString() Label = %v, want %v", got.Label, FieldLabelOptional)
				}
				if got.KeyType != nil {
					t.Errorf("ParseTypeString() KeyType should be nil for basic types")
				}
				if got.ValueType != nil {
					t.Errorf("ParseTypeString() ValueType should be nil for basic types")
				}
			}
		})
	}
}

func TestParseTypeString_EnumAndMMEObject(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantKind     FieldKind
		wantTypeName string
	}{
		{"enum", "enum", FieldKindEnum, "enum"},
		{"MMEObject", "MMEObject", FieldKindMMEObject, "MMEObject"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if err != nil {
				t.Errorf("ParseTypeString() error = %v", err)
				return
			}
			if got.Kind != tt.wantKind {
				t.Errorf("ParseTypeString() Kind = %v, want %v", got.Kind, tt.wantKind)
			}
			if got.TypeName != tt.wantTypeName {
				t.Errorf("ParseTypeString() TypeName = %v, want %v", got.TypeName, tt.wantTypeName)
			}
		})
	}
}

func TestParseTypeString_MMEObjectTypes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantKind     FieldKind
		wantTypeName string
	}{
		{"Entity simple", "PlayerEntity", FieldKindMMEObject, "PlayerEntity"},
		{"Entity qualified", "mme.PlayerEntity", FieldKindMMEObject, "mme.PlayerEntity"},
		{"Manager simple", "HeroManager", FieldKindMMEObject, "HeroManager"},
		{"Manager qualified", "mme.HeroManager", FieldKindMMEObject, "mme.HeroManager"},
		{"Module simple", "HeroModule", FieldKindMMEObject, "HeroModule"},
		{"Module qualified", "mme.HeroModule", FieldKindMMEObject, "mme.HeroModule"},
		{"Mechanism simple", "HeroMechanism", FieldKindMMEObject, "HeroMechanism"},
		{"Mechanism qualified", "mme.HeroMechanism", FieldKindMMEObject, "mme.HeroMechanism"},
		{"LevelUpMechanism", "LevelUpMechanism", FieldKindMMEObject, "LevelUpMechanism"},
		{"nested qualified Manager", "core.mme.HeroManager", FieldKindMMEObject, "core.mme.HeroManager"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if err != nil {
				t.Errorf("ParseTypeString() error = %v", err)
				return
			}
			if got.Kind != tt.wantKind {
				t.Errorf("ParseTypeString() Kind = %v, want %v", got.Kind, tt.wantKind)
			}
			if got.TypeName != tt.wantTypeName {
				t.Errorf("ParseTypeString() TypeName = %v, want %v", got.TypeName, tt.wantTypeName)
			}
		})
	}
}

func TestParseTypeString_MessageTypes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantKind     FieldKind
		wantTypeName string
	}{
		{"simple message", "Book", FieldKindMessage, "Book"},
		{"qualified message", "core.Book", FieldKindMessage, "core.Book"},
		{"nested qualified", "core.mme.Book", FieldKindMessage, "core.mme.Book"},
		{"custom type", "CustomType", FieldKindMessage, "CustomType"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if err != nil {
				t.Errorf("ParseTypeString() error = %v", err)
				return
			}
			if got.Kind != tt.wantKind {
				t.Errorf("ParseTypeString() Kind = %v, want %v", got.Kind, tt.wantKind)
			}
			if got.TypeName != tt.wantTypeName {
				t.Errorf("ParseTypeString() TypeName = %v, want %v", got.TypeName, tt.wantTypeName)
			}
		})
	}
}

func TestParseTypeString_MapTypes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantKind      FieldKind
		wantKeyKind   FieldKind
		wantValueKind FieldKind
		wantKeyType   string
		wantValueType string
	}{
		{
			name:          "simple map",
			input:         "map<int32, string>",
			wantKind:      FieldKindMap,
			wantKeyKind:   FieldKindInt32,
			wantValueKind: FieldKindString,
			wantKeyType:   "int32",
			wantValueType: "string",
		},
		{
			name:          "xmap",
			input:         "xmap<int64, string>",
			wantKind:      FieldKindXMap,
			wantKeyKind:   FieldKindInt64,
			wantValueKind: FieldKindString,
			wantKeyType:   "int64",
			wantValueType: "string",
		},
		{
			name:          "map with MME Object value",
			input:         "map<string, HeroManager>",
			wantKind:      FieldKindMap,
			wantKeyKind:   FieldKindString,
			wantValueKind: FieldKindMMEObject,
			wantKeyType:   "string",
			wantValueType: "HeroManager",
		},
		{
			name:          "map with qualified MME Object",
			input:         "map<int32, mme.HeroManager>",
			wantKind:      FieldKindMap,
			wantKeyKind:   FieldKindInt32,
			wantValueKind: FieldKindMMEObject,
			wantKeyType:   "int32",
			wantValueType: "mme.HeroManager",
		},
		{
			name:          "map with message value",
			input:         "map<string, Book>",
			wantKind:      FieldKindMap,
			wantKeyKind:   FieldKindString,
			wantValueKind: FieldKindMessage,
			wantKeyType:   "string",
			wantValueType: "Book",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if err != nil {
				t.Errorf("ParseTypeString() error = %v", err)
				return
			}
			if got == nil {
				t.Errorf("ParseTypeString() returned nil")
				return
			}
			if got.Kind != tt.wantKind {
				t.Errorf("ParseTypeString() Kind = %v, want %v", got.Kind, tt.wantKind)
			}
			if got.KeyType == nil {
				t.Errorf("ParseTypeString() KeyType is nil")
				return
			}
			if got.ValueType == nil {
				t.Errorf("ParseTypeString() ValueType is nil")
				return
			}
			if got.KeyType.Kind != tt.wantKeyKind {
				t.Errorf("ParseTypeString() KeyType.Kind = %v, want %v", got.KeyType.Kind, tt.wantKeyKind)
			}
			if got.ValueType.Kind != tt.wantValueKind {
				t.Errorf("ParseTypeString() ValueType.Kind = %v, want %v", got.ValueType.Kind, tt.wantValueKind)
			}
			if got.KeyType.String() != tt.wantKeyType {
				t.Errorf("ParseTypeString() KeyType.String() = %v, want %v", got.KeyType.String(), tt.wantKeyType)
			}
			if got.ValueType.String() != tt.wantValueType {
				t.Errorf("ParseTypeString() ValueType.String() = %v, want %v", got.ValueType.String(), tt.wantValueType)
			}
		})
	}
}

func TestParseTypeString_RepeatedTypes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantKind      FieldKind
		wantLabel     FieldLabel
		wantValueKind FieldKind
		wantValueType string
	}{
		{
			name:          "repeated int32",
			input:         "repeated int32",
			wantKind:      FieldKindRepeated,
			wantLabel:     FieldLabelRepeated,
			wantValueKind: FieldKindInt32,
			wantValueType: "int32",
		},
		{
			name:          "repeated string",
			input:         "repeated string",
			wantKind:      FieldKindRepeated,
			wantLabel:     FieldLabelRepeated,
			wantValueKind: FieldKindString,
			wantValueType: "string",
		},
		{
			name:          "repeated MME Object",
			input:         "repeated HeroManager",
			wantKind:      FieldKindRepeated,
			wantLabel:     FieldLabelRepeated,
			wantValueKind: FieldKindMMEObject,
			wantValueType: "HeroManager",
		},
		{
			name:          "repeated qualified MME Object",
			input:         "repeated mme.HeroManager",
			wantKind:      FieldKindRepeated,
			wantLabel:     FieldLabelRepeated,
			wantValueKind: FieldKindMMEObject,
			wantValueType: "mme.HeroManager",
		},
		{
			name:          "repeated message",
			input:         "repeated Book",
			wantKind:      FieldKindRepeated,
			wantLabel:     FieldLabelRepeated,
			wantValueKind: FieldKindMessage,
			wantValueType: "Book",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if err != nil {
				t.Errorf("ParseTypeString() error = %v", err)
				return
			}
			if got == nil {
				t.Errorf("ParseTypeString() returned nil")
				return
			}
			if got.Kind != tt.wantKind {
				t.Errorf("ParseTypeString() Kind = %v, want %v", got.Kind, tt.wantKind)
			}
			if got.Label != tt.wantLabel {
				t.Errorf("ParseTypeString() Label = %v, want %v", got.Label, tt.wantLabel)
			}
			if got.KeyType != nil {
				t.Errorf("ParseTypeString() KeyType should be nil for repeated types")
			}
			if got.ValueType == nil {
				t.Errorf("ParseTypeString() ValueType is nil")
				return
			}
			if got.ValueType.Kind != tt.wantValueKind {
				t.Errorf("ParseTypeString() ValueType.Kind = %v, want %v", got.ValueType.Kind, tt.wantValueKind)
			}
			if got.ValueType.String() != tt.wantValueType {
				t.Errorf("ParseTypeString() ValueType.String() = %v, want %v", got.ValueType.String(), tt.wantValueType)
			}
		})
	}
}

func TestParseTypeString_NestedTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKind FieldKind
		validate func(*testing.T, *FieldType)
	}{
		{
			name:     "nested map",
			input:    "map<int32, map<string, int64>>",
			wantKind: FieldKindMap,
			validate: func(t *testing.T, ft *FieldType) {
				if ft.KeyType.Kind != FieldKindInt32 {
					t.Errorf("KeyType.Kind = %v, want %v", ft.KeyType.Kind, FieldKindInt32)
				}
				if ft.ValueType.Kind != FieldKindMap {
					t.Errorf("ValueType.Kind = %v, want %v", ft.ValueType.Kind, FieldKindMap)
				}
				if ft.ValueType.KeyType.Kind != FieldKindString {
					t.Errorf("ValueType.KeyType.Kind = %v, want %v", ft.ValueType.KeyType.Kind, FieldKindString)
				}
				if ft.ValueType.ValueType.Kind != FieldKindInt64 {
					t.Errorf("ValueType.ValueType.Kind = %v, want %v", ft.ValueType.ValueType.Kind, FieldKindInt64)
				}
			},
		},
		{
			name:     "repeated map",
			input:    "repeated map<string, int32>",
			wantKind: FieldKindRepeated,
			validate: func(t *testing.T, ft *FieldType) {
				if ft.ValueType.Kind != FieldKindMap {
					t.Errorf("ValueType.Kind = %v, want %v", ft.ValueType.Kind, FieldKindMap)
				}
				if ft.ValueType.KeyType.Kind != FieldKindString {
					t.Errorf("ValueType.KeyType.Kind = %v, want %v", ft.ValueType.KeyType.Kind, FieldKindString)
				}
				if ft.ValueType.ValueType.Kind != FieldKindInt32 {
					t.Errorf("ValueType.ValueType.Kind = %v, want %v", ft.ValueType.ValueType.Kind, FieldKindInt32)
				}
			},
		},
		{
			name:     "map with repeated value",
			input:    "map<string, repeated int32>",
			wantKind: FieldKindMap,
			validate: func(t *testing.T, ft *FieldType) {
				if ft.KeyType.Kind != FieldKindString {
					t.Errorf("KeyType.Kind = %v, want %v", ft.KeyType.Kind, FieldKindString)
				}
				if ft.ValueType.Kind != FieldKindRepeated {
					t.Errorf("ValueType.Kind = %v, want %v", ft.ValueType.Kind, FieldKindRepeated)
				}
				if ft.ValueType.ValueType.Kind != FieldKindInt32 {
					t.Errorf("ValueType.ValueType.Kind = %v, want %v", ft.ValueType.ValueType.Kind, FieldKindInt32)
				}
			},
		},
		{
			name:     "xmap nested",
			input:    "xmap<int64, map<string, HeroManager>>",
			wantKind: FieldKindXMap,
			validate: func(t *testing.T, ft *FieldType) {
				if ft.KeyType.Kind != FieldKindInt64 {
					t.Errorf("KeyType.Kind = %v, want %v", ft.KeyType.Kind, FieldKindInt64)
				}
				if ft.ValueType.Kind != FieldKindMap {
					t.Errorf("ValueType.Kind = %v, want %v", ft.ValueType.Kind, FieldKindMap)
				}
				if ft.ValueType.KeyType.Kind != FieldKindString {
					t.Errorf("ValueType.KeyType.Kind = %v, want %v", ft.ValueType.KeyType.Kind, FieldKindString)
				}
				if ft.ValueType.ValueType.Kind != FieldKindMMEObject {
					t.Errorf("ValueType.ValueType.Kind = %v, want %v", ft.ValueType.ValueType.Kind, FieldKindMMEObject)
				}
			},
		},
		{
			name:     "triple nested map",
			input:    "map<int32, map<string, map<int64, bool>>>",
			wantKind: FieldKindMap,
			validate: func(t *testing.T, ft *FieldType) {
				if ft.KeyType.Kind != FieldKindInt32 {
					t.Errorf("KeyType.Kind = %v, want %v", ft.KeyType.Kind, FieldKindInt32)
				}
				if ft.ValueType.Kind != FieldKindMap {
					t.Errorf("ValueType.Kind = %v, want %v", ft.ValueType.Kind, FieldKindMap)
				}
				if ft.ValueType.ValueType.Kind != FieldKindMap {
					t.Errorf("ValueType.ValueType.Kind = %v, want %v", ft.ValueType.ValueType.Kind, FieldKindMap)
				}
				if ft.ValueType.ValueType.ValueType.Kind != FieldKindBool {
					t.Errorf("ValueType.ValueType.ValueType.Kind = %v, want %v", ft.ValueType.ValueType.ValueType.Kind, FieldKindBool)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if err != nil {
				t.Errorf("ParseTypeString() error = %v", err)
				return
			}
			if got == nil {
				t.Errorf("ParseTypeString() returned nil")
				return
			}
			if got.Kind != tt.wantKind {
				t.Errorf("ParseTypeString() Kind = %v, want %v", got.Kind, tt.wantKind)
			}
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestParseTypeString_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"empty string", "", true, "empty type string"},
		{"whitespace only", "   ", true, "empty type string"},
		{"repeated without type", "repeated", false, ""},            // "repeated" itself is recognized as a type
		{"repeated with only whitespace", "repeated   ", false, ""}, // Same, whitespace is trimmed
		{"map missing closing bracket", "map<int32, string", true, "missing closing '>'"},
		{"map missing comma", "map<int32 string>", true, "no comma found"},
		{"map missing key", "map<, string>", true, "empty type string"},
		{"map missing value", "map<int32, >", true, "empty type string"},
		{"xmap invalid format", "xmap<int32>", true, "no comma found"},
		{"unmatched closing bracket", "map<int32, string>>", true, "unmatched '>'"},
		{"unmatched opening bracket", "map<<int32, string>", true, "no comma found"}, // Will fail at comma detection
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTypeString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTypeString() expected error but got nil")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseTypeString() error message = %v, want containing %v", err.Error(), tt.errMsg)
				}
			} else if got == nil {
				t.Errorf("ParseTypeString() returned nil")
			}
		})
	}
}

func TestParseTypeString_StringMethod(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"int32", "int32", "int32"},
		{"string", "string", "string"},
		{"MME Object", "HeroManager", "HeroManager"},
		{"qualified MME Object", "mme.HeroManager", "mme.HeroManager"},
		{"message", "Book", "Book"},
		{"simple map", "map<int32, string>", "map<int32, string>"},
		{"xmap", "xmap<int64, string>", "xmap<int64, string>"},
		{"repeated", "repeated int32", "repeated int32"},
		{"repeated map", "repeated map<string, int32>", "repeated map<string, int32>"},
		{"nested map", "map<int32, map<string, int64>>", "map<int32, map<string, int64>>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if err != nil {
				t.Errorf("ParseTypeString() error = %v", err)
				return
			}
			if got.String() != tt.want {
				t.Errorf("ParseTypeString().String() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestParseTypeString_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKind FieldKind
	}{
		{"leading whitespace", "  int32", FieldKindInt32},
		{"trailing whitespace", "int32  ", FieldKindInt32},
		{"both whitespace", "  int32  ", FieldKindInt32},
		{"whitespace in map", "map< int32 , string >", FieldKindMap},
		{"whitespace in repeated", "repeated  int32", FieldKindRepeated},
		{"whitespace in nested map", "map< int32 , map< string , int64 > >", FieldKindMap},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if err != nil {
				t.Errorf("ParseTypeString() error = %v", err)
				return
			}
			if got.Kind != tt.wantKind {
				t.Errorf("ParseTypeString() Kind = %v, want %v", got.Kind, tt.wantKind)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFieldType_String(t *testing.T) {
	tests := []struct {
		name string
		ft   *FieldType
		want string
	}{
		{"nil", nil, "nil"},
		{"invalid map without key", &FieldType{Kind: FieldKindMap, KeyType: nil, ValueType: &FieldType{Kind: FieldKindString}}, "invalid map"},
		{"invalid map without value", &FieldType{Kind: FieldKindMap, KeyType: &FieldType{Kind: FieldKindInt32}, ValueType: nil}, "invalid map"},
		{"invalid repeated", &FieldType{Kind: FieldKindRepeated, ValueType: nil}, "invalid repeated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ft.String(); got != tt.want {
				t.Errorf("FieldType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTypeString_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(*testing.T, *FieldType)
	}{
		{
			name:  "repeated map with MME Object",
			input: "repeated map<string, mme.HeroManager>",
			validate: func(t *testing.T, ft *FieldType) {
				if ft.Kind != FieldKindRepeated {
					t.Errorf("Kind = %v, want %v", ft.Kind, FieldKindRepeated)
				}
				if ft.ValueType.Kind != FieldKindMap {
					t.Errorf("ValueType.Kind = %v, want %v", ft.ValueType.Kind, FieldKindMap)
				}
				if ft.ValueType.KeyType.Kind != FieldKindString {
					t.Errorf("ValueType.KeyType.Kind = %v, want %v", ft.ValueType.KeyType.Kind, FieldKindString)
				}
				if ft.ValueType.ValueType.Kind != FieldKindMMEObject {
					t.Errorf("ValueType.ValueType.Kind = %v, want %v", ft.ValueType.ValueType.Kind, FieldKindMMEObject)
				}
				if ft.ValueType.ValueType.TypeName != "mme.HeroManager" {
					t.Errorf("ValueType.ValueType.TypeName = %v, want %v", ft.ValueType.ValueType.TypeName, "mme.HeroManager")
				}
			},
		},
		{
			name:  "map with repeated MME Object",
			input: "map<int32, repeated HeroManager>",
			validate: func(t *testing.T, ft *FieldType) {
				if ft.Kind != FieldKindMap {
					t.Errorf("Kind = %v, want %v", ft.Kind, FieldKindMap)
				}
				if ft.ValueType.Kind != FieldKindRepeated {
					t.Errorf("ValueType.Kind = %v, want %v", ft.ValueType.Kind, FieldKindRepeated)
				}
				if ft.ValueType.ValueType.Kind != FieldKindMMEObject {
					t.Errorf("ValueType.ValueType.Kind = %v, want %v", ft.ValueType.ValueType.Kind, FieldKindMMEObject)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeString(tt.input)
			if err != nil {
				t.Errorf("ParseTypeString() error = %v", err)
				return
			}
			if got == nil {
				t.Errorf("ParseTypeString() returned nil")
				return
			}
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseTypeString_BasicType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ParseTypeString("int32")
	}
}

func BenchmarkParseTypeString_MapType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ParseTypeString("map<int32, string>")
	}
}

func BenchmarkParseTypeString_NestedMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ParseTypeString("map<int32, map<string, int64>>")
	}
}

func BenchmarkParseTypeString_RepeatedMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ParseTypeString("repeated map<string, Book>")
	}
}
