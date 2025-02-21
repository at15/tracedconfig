package slowjson

import (
	"strings"
	"testing"
)

func TestParser_Parse_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType NodeType
		wantLen  int // For arrays/objects, number of children
	}{
		{
			name:     "empty object",
			input:    "{}",
			wantType: NodeObject,
			wantLen:  0,
		},
		{
			name:     "simple object",
			input:    `{"key": "value"}`,
			wantType: NodeObject,
			wantLen:  1,
		},
		{
			name:     "empty array",
			input:    "[]",
			wantType: NodeArray,
			wantLen:  0,
		},
		{
			name:     "simple array",
			input:    `[1, 2, 3]`,
			wantType: NodeArray,
			wantLen:  3,
		},
		{
			name:     "complex nested structure",
			input:    `{"array": [1, true, "string"], "object": {"nested": null}}`,
			wantType: NodeObject,
			wantLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			got, err := p.Parse()
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("Parse() got type = %v, want %v", got.Type, tt.wantType)
			}
			if (got.Type == NodeObject || got.Type == NodeArray) && len(got.Children) != tt.wantLen {
				t.Errorf("Parse() got len = %v, want %v", len(got.Children), tt.wantLen)
			}
		})
	}
}

func TestParser_Parse_Errors(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError string
	}{
		{
			name:      "unclosed object",
			input:     `{"key": "value"`,
			wantError: "expected ',' or '}' in object",
		},
		{
			name:      "unclosed array",
			input:     `[1, 2, 3`,
			wantError: "expected ',' or ']' in array",
		},
		{
			name:      "unclosed string",
			input:     `{"key": "value`,
			wantError: "unexpected end of input in string",
		},
		{
			name:      "invalid boolean",
			input:     `{"key": tru}`,
			wantError: "invalid boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			_, err := p.Parse()
			if err == nil {
				t.Error("Parse() expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Parse() error = %v, want error containing %v", err, tt.wantError)
			}
		})
	}
}

func TestNode_DebugContext(t *testing.T) {
	input := `{
    "key1": "value1",
    "key2": {
        "nested": true
    }
}`
	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Test debug context with different line ranges
	tests := []struct {
		name        string
		linesBefore int
		linesAfter  int
		wantLines   int // Expected number of lines in output
	}{
		{
			name:        "no context",
			linesBefore: 0,
			linesAfter:  0,
			wantLines:   1,
		},
		{
			name:        "one line context",
			linesBefore: 1,
			linesAfter:  1,
			wantLines:   3,
		},
		{
			name:        "full context",
			linesBefore: 5,
			linesAfter:  5,
			wantLines:   5, // The input has 5 lines total
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := node.DebugContext(tt.linesBefore, tt.linesAfter)
			lines := strings.Count(context, "\n")
			if lines < tt.wantLines {
				t.Errorf("DebugContext() got %v lines, want at least %v", lines, tt.wantLines)
			}
		})
	}
}

func TestParser_Parse_Types(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType NodeType
		wantVal  string
	}{
		{
			name:     "string",
			input:    `"test string"`,
			wantType: NodeString,
			wantVal:  "test string",
		},
		{
			name:     "number",
			input:    `42.5`,
			wantType: NodeNumber,
			wantVal:  "42.5",
		},
		{
			name:     "boolean true",
			input:    `true`,
			wantType: NodeBoolean,
			wantVal:  "true",
		},
		{
			name:     "boolean false",
			input:    `false`,
			wantType: NodeBoolean,
			wantVal:  "false",
		},
		{
			name:     "null",
			input:    `null`,
			wantType: NodeNull,
			wantVal:  "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			got, err := p.Parse()
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("Parse() got type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Value != tt.wantVal {
				t.Errorf("Parse() got value = %v, want %v", got.Value, tt.wantVal)
			}
		})
	}
}
