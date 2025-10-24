package protocol

import (
	"testing"
)

func TestParseValidCommands(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      *Command
		wantError bool
	}{
		{
			name:  "GET command",
			input: "GET mykey",
			want: &Command{
				Directive: GetDirective,
				Key:       "mykey",
				Value:     "",
			},
			wantError: false,
		},
		{
			name:  "SET command",
			input: "SET mykey myvalue",
			want: &Command{
				Directive: SetDirective,
				Key:       "mykey",
				Value:     "myvalue",
			},
			wantError: false,
		},
		{
			name:  "DELETE command",
			input: "DELETE mykey",
			want: &Command{
				Directive: DeleteDirective,
				Key:       "mykey",
				Value:     "",
			},
			wantError: false,
		},
		{
			name:  "EXISTS command",
			input: "EXISTS mykey",
			want: &Command{
				Directive: ExistsDirective,
				Key:       "mykey",
				Value:     "",
			},
			wantError: false,
		},
		{
			name:  "SET with numeric value",
			input: "SET count 42",
			want: &Command{
				Directive: SetDirective,
				Key:       "count",
				Value:     "42",
			},
			wantError: false,
		},
		{
			name:  "Key with special characters",
			input: "GET user:123:name",
			want: &Command{
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
		want  *Command
	}{
		{
			name:  "Multiple spaces between words",
			input: "GET    mykey",
			want: &Command{
				Directive: GetDirective,
				Key:       "mykey",
			},
		},
		{
			name:  "Leading whitespace",
			input: "   GET mykey",
			want: &Command{
				Directive: GetDirective,
				Key:       "mykey",
			},
		},
		{
			name:  "Trailing whitespace",
			input: "GET mykey   ",
			want: &Command{
				Directive: GetDirective,
				Key:       "mykey",
			},
		},
		{
			name:  "Tabs instead of spaces",
			input: "GET\tmykey",
			want: &Command{
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
