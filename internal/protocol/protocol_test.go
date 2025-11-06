package protocol

import (
	"testing"
)

func TestParseValidCommands(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      *Operation
		wantError bool
	}{
		{
			name:  "GET command",
			input: "GET mykey",
			want: &Operation{
				Directive: GetDirective,
				Key:       "mykey",
				Value:     "",
			},
			wantError: false,
		},
		{
			name:  "SET command",
			input: "SET mykey myvalue",
			want: &Operation{
				Directive: SetDirective,
				Key:       "mykey",
				Value:     "myvalue",
			},
			wantError: false,
		},
		{
			name:  "DELETE command",
			input: "DELETE mykey",
			want: &Operation{
				Directive: DeleteDirective,
				Key:       "mykey",
				Value:     "",
			},
			wantError: false,
		},
		{
			name:  "EXISTS command",
			input: "EXISTS mykey",
			want: &Operation{
				Directive: ExistsDirective,
				Key:       "mykey",
				Value:     "",
			},
			wantError: false,
		},
		{
			name:  "SET with numeric value",
			input: "SET count 42",
			want: &Operation{
				Directive: SetDirective,
				Key:       "count",
				Value:     "42",
			},
			wantError: false,
		},
		{
			name:  "Key with special characters",
			input: "GET user:123:name",
			want: &Operation{
				Directive: GetDirective,
				Key:       "user:123:name",
				Value:     "",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("ParseCommand() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err == nil {
				if got.Directive != tt.want.Directive {
					t.Errorf("ParseCommand() Directive = %v, want %v", got.Directive, tt.want.Directive)
				}
				if got.Key != tt.want.Key {
					t.Errorf("ParseCommand() Key = %v, want %v", got.Key, tt.want.Key)
				}
				if got.Value != tt.want.Value {
					t.Errorf("ParseCommand() Value = %v, want %v", got.Value, tt.want.Value)
				}
			}
		})
	}
}

func TestParseInvalidCommands(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Unknown directive",
			input: "INVALID mykey",
		},
		{
			name:  "GET with too many args",
			input: "GET mykey extraarg",
		},
		{
			name:  "GET with no args",
			input: "GET",
		},
		{
			name:  "SET with only key",
			input: "SET mykey",
		},
		{
			name:  "SET with too many args",
			input: "SET mykey myvalue extra",
		},
		{
			name:  "DELETE with no args",
			input: "DELETE",
		},
		{
			name:  "DELETE with too many args",
			input: "DELETE mykey extraarg",
		},
		{
			name:  "EXISTS with no args",
			input: "EXISTS",
		},
		{
			name:  "EXISTS with too many args",
			input: "EXISTS mykey extraarg",
		},
		{
			name:  "Empty input",
			input: "",
		},
		{
			name:  "Only whitespace",
			input: "   ",
		},
		{
			name:  "Lowercase directive",
			input: "get mykey",
		},
		{
			name:  "Mixed case directive",
			input: "Get mykey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCommand(tt.input)
			if err == nil {
				t.Errorf("ParseCommand() expected error for input %q, got nil", tt.input)
			}
		})
	}
}

func TestParseCommandWithExtraWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *Operation
	}{
		{
			name:  "Multiple spaces between words",
			input: "GET    mykey",
			want: &Operation{
				Directive: GetDirective,
				Key:       "mykey",
			},
		},
		{
			name:  "Leading whitespace",
			input: "   GET mykey",
			want: &Operation{
				Directive: GetDirective,
				Key:       "mykey",
			},
		},
		{
			name:  "Trailing whitespace",
			input: "GET mykey   ",
			want: &Operation{
				Directive: GetDirective,
				Key:       "mykey",
			},
		},
		{
			name:  "Tabs instead of spaces",
			input: "GET\tmykey",
			want: &Operation{
				Directive: GetDirective,
				Key:       "mykey",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.input)
			if err != nil {
				t.Errorf("ParseCommand() unexpected error = %v", err)
				return
			}
			if got.Directive != tt.want.Directive {
				t.Errorf("ParseCommand() Directive = %v, want %v", got.Directive, tt.want.Directive)
			}
			if got.Key != tt.want.Key {
				t.Errorf("ParseCommand() Key = %v, want %v", got.Key, tt.want.Key)
			}
		})
	}
}

func TestDirectiveIsValid(t *testing.T) {
	tests := []struct {
		directive Directive
		want      bool
	}{
		{GetDirective, true},
		{SetDirective, true},
		{DeleteDirective, true},
		{ExistsDirective, true},
		{Directive("INVALID"), false},
		{Directive("get"), false},
		{Directive(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.directive), func(t *testing.T) {
			if got := tt.directive.IsValid(); got != tt.want {
				t.Errorf("Directive.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommandString(t *testing.T) {
	tests := []struct {
		name    string
		command Operation
		want    string
	}{
		{
			name: "GET command",
			command: Operation{
				Directive: GetDirective,
				Key:       "mykey",
				Value:     "",
			},
			want: "GET mykey",
		},
		{
			name: "SET command",
			command: Operation{
				Directive: SetDirective,
				Key:       "mykey",
				Value:     "myvalue",
			},
			want: "SET mykey myvalue",
		},
		{
			name: "DELETE command",
			command: Operation{
				Directive: DeleteDirective,
				Key:       "mykey",
				Value:     "",
			},
			want: "DELETE mykey",
		},
		{
			name: "EXISTS command",
			command: Operation{
				Directive: ExistsDirective,
				Key:       "mykey",
				Value:     "",
			},
			want: "EXISTS mykey",
		},
		{
			name: "SET with numeric value",
			command: Operation{
				Directive: SetDirective,
				Key:       "count",
				Value:     "42",
			},
			want: "SET count 42",
		},
		{
			name: "Key with special characters",
			command: Operation{
				Directive: GetDirective,
				Key:       "user:123:name",
				Value:     "",
			},
			want: "GET user:123:name",
		},
		{
			name: "SET with special characters in value",
			command: Operation{
				Directive: SetDirective,
				Key:       "config",
				Value:     "value-with-dashes",
			},
			want: "SET config value-with-dashes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.command.String()
			if got != tt.want {
				t.Errorf("Command.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommandStringRoundTrip(t *testing.T) {
	// Test that String() produces output that can be parsed back
	tests := []Operation{
		{Directive: GetDirective, Key: "foo", Value: ""},
		{Directive: SetDirective, Key: "foo", Value: "bar"},
		{Directive: DeleteDirective, Key: "baz", Value: ""},
		{Directive: ExistsDirective, Key: "qux", Value: ""},
	}

	for _, original := range tests {
		t.Run(string(original.Directive), func(t *testing.T) {
			// Convert to string
			cmdString := original.String()

			// Parse it back
			parsed, err := ParseCommand(cmdString)
			if err != nil {
				t.Fatalf("ParseCommand(%q) failed: %v", cmdString, err)
			}

			// Should match original
			if parsed.Directive != original.Directive {
				t.Errorf("Directive = %v, want %v", parsed.Directive, original.Directive)
			}
			if parsed.Key != original.Key {
				t.Errorf("Key = %v, want %v", parsed.Key, original.Key)
			}
			if parsed.Value != original.Value {
				t.Errorf("Value = %v, want %v", parsed.Value, original.Value)
			}
		})
	}
}
