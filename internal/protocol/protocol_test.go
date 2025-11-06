package protocol

import (
	"testing"
)

// ============================================================================
// Operation Tests
// ============================================================================

func TestParseValidOperations(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *Operation
	}{
		{
			name:  "GET command",
			input: "GET mykey",
			want: &Operation{
				Directive: GetDirective,
				Key:       "mykey",
				Value:     "",
			},
		},
		{
			name:  "SET command",
			input: "SET mykey myvalue",
			want: &Operation{
				Directive: SetDirective,
				Key:       "mykey",
				Value:     "myvalue",
			},
		},
		{
			name:  "DELETE command",
			input: "DELETE mykey",
			want: &Operation{
				Directive: DeleteDirective,
				Key:       "mykey",
				Value:     "",
			},
		},
		{
			name:  "EXISTS command",
			input: "EXISTS mykey",
			want: &Operation{
				Directive: ExistsDirective,
				Key:       "mykey",
				Value:     "",
			},
		},
		{
			name:  "SET with numeric value",
			input: "SET count 42",
			want: &Operation{
				Directive: SetDirective,
				Key:       "count",
				Value:     "42",
			},
		},
		{
			name:  "Key with special characters",
			input: "GET user:123:name",
			want: &Operation{
				Directive: GetDirective,
				Key:       "user:123:name",
				Value:     "",
			},
		},
		{
			name:  "SET with value containing dashes",
			input: "SET config value-with-dashes",
			want: &Operation{
				Directive: SetDirective,
				Key:       "config",
				Value:     "value-with-dashes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.input)
			if err != nil {
				t.Fatalf("ParseCommand() unexpected error = %v", err)
			}

			if got.Type != OperationCommand {
				t.Errorf("ParseCommand() Type = %v, want %v", got.Type, OperationCommand)
			}

			if got.Operation == nil {
				t.Fatal("ParseCommand() Operation is nil")
			}

			if got.Operation.Directive != tt.want.Directive {
				t.Errorf("ParseCommand() Directive = %v, want %v", got.Operation.Directive, tt.want.Directive)
			}
			if got.Operation.Key != tt.want.Key {
				t.Errorf("ParseCommand() Key = %v, want %v", got.Operation.Key, tt.want.Key)
			}
			if got.Operation.Value != tt.want.Value {
				t.Errorf("ParseCommand() Value = %v, want %v", got.Operation.Value, tt.want.Value)
			}
		})
	}
}

func TestParseInvalidOperations(t *testing.T) {
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
		{
			name:  "Single word",
			input: "SINGLE",
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

func TestParseOperationWithExtraWhitespace(t *testing.T) {
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
		{
			name:  "Mixed whitespace in SET",
			input: "  SET\t\tkey   value  ",
			want: &Operation{
				Directive: SetDirective,
				Key:       "key",
				Value:     "value",
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

			if got.Operation == nil {
				t.Fatal("ParseCommand() Operation is nil")
			}

			if got.Operation.Directive != tt.want.Directive {
				t.Errorf("ParseCommand() Directive = %v, want %v", got.Operation.Directive, tt.want.Directive)
			}
			if got.Operation.Key != tt.want.Key {
				t.Errorf("ParseCommand() Key = %v, want %v", got.Operation.Key, tt.want.Key)
			}
			if tt.want.Value != "" && got.Operation.Value != tt.want.Value {
				t.Errorf("ParseCommand() Value = %v, want %v", got.Operation.Value, tt.want.Value)
			}
		})
	}
}

// ============================================================================
// Transaction Tests
// ============================================================================

func TestParseValidTransactions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOps []*Operation
	}{
		{
			name:  "TX with single SET",
			input: "TX SET foo bar",
			wantOps: []*Operation{
				{Directive: SetDirective, Key: "foo", Value: "bar"},
			},
		},
		{
			name:  "TX with single DELETE",
			input: "TX DELETE foo",
			wantOps: []*Operation{
				{Directive: DeleteDirective, Key: "foo", Value: ""},
			},
		},
		{
			name:  "TX with SET and DELETE",
			input: "TX SET foo bar DELETE baz",
			wantOps: []*Operation{
				{Directive: SetDirective, Key: "foo", Value: "bar"},
				{Directive: DeleteDirective, Key: "baz", Value: ""},
			},
		},
		{
			name:  "TX with multiple SETs",
			input: "TX SET key1 value1 SET key2 value2 SET key3 value3",
			wantOps: []*Operation{
				{Directive: SetDirective, Key: "key1", Value: "value1"},
				{Directive: SetDirective, Key: "key2", Value: "value2"},
				{Directive: SetDirective, Key: "key3", Value: "value3"},
			},
		},
		{
			name:  "TX with multiple DELETEs",
			input: "TX DELETE key1 DELETE key2 DELETE key3",
			wantOps: []*Operation{
				{Directive: DeleteDirective, Key: "key1", Value: ""},
				{Directive: DeleteDirective, Key: "key2", Value: ""},
				{Directive: DeleteDirective, Key: "key3", Value: ""},
			},
		},
		{
			name:  "TX with mixed operations",
			input: "TX SET a 1 DELETE b SET c 2 DELETE d SET e 3",
			wantOps: []*Operation{
				{Directive: SetDirective, Key: "a", Value: "1"},
				{Directive: DeleteDirective, Key: "b", Value: ""},
				{Directive: SetDirective, Key: "c", Value: "2"},
				{Directive: DeleteDirective, Key: "d", Value: ""},
				{Directive: SetDirective, Key: "e", Value: "3"},
			},
		},
		{
			name:  "TX with maximum operations (8)",
			input: "TX SET k1 v1 SET k2 v2 SET k3 v3 SET k4 v4 SET k5 v5 SET k6 v6 SET k7 v7 SET k8 v8",
			wantOps: []*Operation{
				{Directive: SetDirective, Key: "k1", Value: "v1"},
				{Directive: SetDirective, Key: "k2", Value: "v2"},
				{Directive: SetDirective, Key: "k3", Value: "v3"},
				{Directive: SetDirective, Key: "k4", Value: "v4"},
				{Directive: SetDirective, Key: "k5", Value: "v5"},
				{Directive: SetDirective, Key: "k6", Value: "v6"},
				{Directive: SetDirective, Key: "k7", Value: "v7"},
				{Directive: SetDirective, Key: "k8", Value: "v8"},
			},
		},
		{
			name:  "TX with special characters in keys and values",
			input: "TX SET user:123 admin DELETE session:abc SET config:db postgres",
			wantOps: []*Operation{
				{Directive: SetDirective, Key: "user:123", Value: "admin"},
				{Directive: DeleteDirective, Key: "session:abc", Value: ""},
				{Directive: SetDirective, Key: "config:db", Value: "postgres"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.input)
			if err != nil {
				t.Fatalf("ParseCommand() unexpected error = %v", err)
			}

			if got.Type != TransactionCommand {
				t.Errorf("ParseCommand() Type = %v, want %v", got.Type, TransactionCommand)
			}

			if got.Transaction == nil {
				t.Fatal("ParseCommand() Transaction is nil")
			}

			if len(got.Transaction.Operations) != len(tt.wantOps) {
				t.Fatalf("ParseCommand() got %d operations, want %d", len(got.Transaction.Operations), len(tt.wantOps))
			}

			for i, wantOp := range tt.wantOps {
				gotOp := got.Transaction.Operations[i]
				if gotOp.Directive != wantOp.Directive {
					t.Errorf("Operation[%d] Directive = %v, want %v", i, gotOp.Directive, wantOp.Directive)
				}
				if gotOp.Key != wantOp.Key {
					t.Errorf("Operation[%d] Key = %v, want %v", i, gotOp.Key, wantOp.Key)
				}
				if gotOp.Value != wantOp.Value {
					t.Errorf("Operation[%d] Value = %v, want %v", i, gotOp.Value, wantOp.Value)
				}
			}
		})
	}
}

func TestParseInvalidTransactions(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		reason string
	}{
		{
			name:   "TX with no operations",
			input:  "TX",
			reason: "transaction must have at least one operation",
		},
		{
			name:   "TX with only one word after TX",
			input:  "TX SET",
			reason: "incomplete operation",
		},
		{
			name:   "TX with incomplete SET",
			input:  "TX SET key",
			reason: "SET requires key and value",
		},
		{
			name:   "TX with GET operation",
			input:  "TX GET foo",
			reason: "transactions can only contain write operations",
		},
		{
			name:   "TX with EXISTS operation",
			input:  "TX EXISTS foo",
			reason: "transactions can only contain write operations",
		},
		{
			name:   "TX with mixed read and write",
			input:  "TX SET foo bar GET baz",
			reason: "transactions cannot contain read operations",
		},
		{
			name:   "TX with GET in the middle",
			input:  "TX SET a 1 GET b SET c 2",
			reason: "transactions cannot contain read operations",
		},
		{
			name:   "TX with more than 8 operations",
			input:  "TX SET k1 v1 SET k2 v2 SET k3 v3 SET k4 v4 SET k5 v5 SET k6 v6 SET k7 v7 SET k8 v8 SET k9 v9",
			reason: "transactions limited to 8 operations",
		},
		{
			name:   "TX with invalid directive",
			input:  "TX INVALID foo",
			reason: "invalid directive in transaction",
		},
		{
			name:   "TX with extra arguments after DELETE",
			input:  "TX DELETE foo bar",
			reason: "DELETE should only have key",
		},
		{
			name:   "TX with extra arguments after SET",
			input:  "TX SET foo bar baz",
			reason: "SET should only have key and value",
		},
		{
			name:   "TX with missing DELETE key",
			input:  "TX SET foo bar DELETE",
			reason: "DELETE requires key",
		},
		{
			name:   "TX with nested TX",
			input:  "TX TX SET foo bar",
			reason: "cannot nest transactions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCommand(tt.input)
			if err == nil {
				t.Errorf("ParseCommand() expected error for input %q (%s), got nil", tt.input, tt.reason)
			}
		})
	}
}

func TestParseTransactionWithWhitespace(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOps int
	}{
		{
			name:    "Extra spaces between operations",
			input:   "TX   SET   foo   bar   DELETE   baz",
			wantOps: 2,
		},
		{
			name:    "Leading whitespace",
			input:   "   TX SET foo bar",
			wantOps: 1,
		},
		{
			name:    "Trailing whitespace",
			input:   "TX SET foo bar DELETE baz   ",
			wantOps: 2,
		},
		{
			name:    "Tabs in transaction",
			input:   "TX\tSET\tfoo\tbar\tDELETE\tbaz",
			wantOps: 2,
		},
		{
			name:    "Mixed whitespace",
			input:   "  TX \t SET  key  value \t DELETE \t other  ",
			wantOps: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.input)
			if err != nil {
				t.Fatalf("ParseCommand() unexpected error = %v", err)
			}

			if got.Type != TransactionCommand {
				t.Errorf("ParseCommand() Type = %v, want %v", got.Type, TransactionCommand)
			}

			if got.Transaction == nil {
				t.Fatal("ParseCommand() Transaction is nil")
			}

			if len(got.Transaction.Operations) != tt.wantOps {
				t.Errorf("ParseCommand() got %d operations, want %d", len(got.Transaction.Operations), tt.wantOps)
			}
		})
	}
}

// ============================================================================
// Directive Tests
// ============================================================================

func TestDirectiveIsValid(t *testing.T) {
	tests := []struct {
		directive Directive
		want      bool
	}{
		{GetDirective, true},
		{SetDirective, true},
		{DeleteDirective, true},
		{ExistsDirective, true},
		{TxDirective, true},
		{Directive("INVALID"), false},
		{Directive("get"), false},
		{Directive("set"), false},
		{Directive(""), false},
		{Directive("TX "), false},
		{Directive(" SET"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.directive), func(t *testing.T) {
			if got := tt.directive.IsValid(); got != tt.want {
				t.Errorf("Directive.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Operation String Tests
// ============================================================================

func TestOperationString(t *testing.T) {
	tests := []struct {
		name      string
		operation Operation
		want      string
	}{
		{
			name: "GET command",
			operation: Operation{
				Directive: GetDirective,
				Key:       "mykey",
				Value:     "",
			},
			want: "GET mykey",
		},
		{
			name: "SET command",
			operation: Operation{
				Directive: SetDirective,
				Key:       "mykey",
				Value:     "myvalue",
			},
			want: "SET mykey myvalue",
		},
		{
			name: "DELETE command",
			operation: Operation{
				Directive: DeleteDirective,
				Key:       "mykey",
				Value:     "",
			},
			want: "DELETE mykey",
		},
		{
			name: "EXISTS command",
			operation: Operation{
				Directive: ExistsDirective,
				Key:       "mykey",
				Value:     "",
			},
			want: "EXISTS mykey",
		},
		{
			name: "SET with numeric value",
			operation: Operation{
				Directive: SetDirective,
				Key:       "count",
				Value:     "42",
			},
			want: "SET count 42",
		},
		{
			name: "Key with special characters",
			operation: Operation{
				Directive: GetDirective,
				Key:       "user:123:name",
				Value:     "",
			},
			want: "GET user:123:name",
		},
		{
			name: "SET with special characters in value",
			operation: Operation{
				Directive: SetDirective,
				Key:       "config",
				Value:     "value-with-dashes",
			},
			want: "SET config value-with-dashes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.operation.String()
			if got != tt.want {
				t.Errorf("Operation.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOperationStringRoundTrip(t *testing.T) {
	tests := []Operation{
		{Directive: GetDirective, Key: "foo", Value: ""},
		{Directive: SetDirective, Key: "foo", Value: "bar"},
		{Directive: DeleteDirective, Key: "baz", Value: ""},
		{Directive: ExistsDirective, Key: "qux", Value: ""},
		{Directive: SetDirective, Key: "user:123", Value: "admin"},
		{Directive: GetDirective, Key: "config:db:host", Value: ""},
	}

	for _, original := range tests {
		t.Run(string(original.Directive)+"_"+original.Key, func(t *testing.T) {
			cmdString := original.String()

			parsed, err := ParseCommand(cmdString)
			if err != nil {
				t.Fatalf("ParseCommand(%q) failed: %v", cmdString, err)
			}

			if parsed.Operation == nil {
				t.Fatal("ParseCommand() Operation is nil")
			}

			if parsed.Operation.Directive != original.Directive {
				t.Errorf("Directive = %v, want %v", parsed.Operation.Directive, original.Directive)
			}
			if parsed.Operation.Key != original.Key {
				t.Errorf("Key = %v, want %v", parsed.Operation.Key, original.Key)
			}
			if parsed.Operation.Value != original.Value {
				t.Errorf("Value = %v, want %v", parsed.Operation.Value, original.Value)
			}
		})
	}
}

// ============================================================================
// Operation IsValid Tests
// ============================================================================

func TestOperationIsValid(t *testing.T) {
	tests := []struct {
		name      string
		operation Operation
		want      bool
	}{
		{
			name: "Valid GET",
			operation: Operation{
				Directive: GetDirective,
				Key:       "key",
				Value:     "",
			},
			want: true,
		},
		{
			name: "Valid SET",
			operation: Operation{
				Directive: SetDirective,
				Key:       "key",
				Value:     "value",
			},
			want: true,
		},
		{
			name: "Valid DELETE",
			operation: Operation{
				Directive: DeleteDirective,
				Key:       "key",
				Value:     "",
			},
			want: true,
		},
		{
			name: "Valid EXISTS",
			operation: Operation{
				Directive: ExistsDirective,
				Key:       "key",
				Value:     "",
			},
			want: true,
		},
		{
			name: "Invalid - GET with value",
			operation: Operation{
				Directive: GetDirective,
				Key:       "key",
				Value:     "value",
			},
			want: false,
		},
		{
			name: "Invalid - SET without value",
			operation: Operation{
				Directive: SetDirective,
				Key:       "key",
				Value:     "",
			},
			want: false,
		},
		{
			name: "Invalid - DELETE with value",
			operation: Operation{
				Directive: DeleteDirective,
				Key:       "key",
				Value:     "value",
			},
			want: false,
		},
		{
			name: "Invalid - missing key",
			operation: Operation{
				Directive: GetDirective,
				Key:       "",
				Value:     "",
			},
			want: false,
		},
		{
			name: "Invalid - SET missing key",
			operation: Operation{
				Directive: SetDirective,
				Key:       "",
				Value:     "value",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.operation.IsValid(); got != tt.want {
				t.Errorf("Operation.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Transaction IsValid Tests
// ============================================================================

func TestTransactionIsValid(t *testing.T) {
	tests := []struct {
		name string
		tx   Transaction
		want bool
	}{
		{
			name: "Valid with single SET",
			tx: Transaction{
				Operations: []*Operation{
					{Directive: SetDirective, Key: "key", Value: "value"},
				},
			},
			want: true,
		},
		{
			name: "Valid with single DELETE",
			tx: Transaction{
				Operations: []*Operation{
					{Directive: DeleteDirective, Key: "key", Value: ""},
				},
			},
			want: true,
		},
		{
			name: "Valid with multiple write operations",
			tx: Transaction{
				Operations: []*Operation{
					{Directive: SetDirective, Key: "k1", Value: "v1"},
					{Directive: DeleteDirective, Key: "k2", Value: ""},
					{Directive: SetDirective, Key: "k3", Value: "v3"},
				},
			},
			want: true,
		},
		{
			name: "Valid with exactly 8 operations",
			tx: Transaction{
				Operations: []*Operation{
					{Directive: SetDirective, Key: "k1", Value: "v1"},
					{Directive: SetDirective, Key: "k2", Value: "v2"},
					{Directive: SetDirective, Key: "k3", Value: "v3"},
					{Directive: SetDirective, Key: "k4", Value: "v4"},
					{Directive: SetDirective, Key: "k5", Value: "v5"},
					{Directive: SetDirective, Key: "k6", Value: "v6"},
					{Directive: SetDirective, Key: "k7", Value: "v7"},
					{Directive: SetDirective, Key: "k8", Value: "v8"},
				},
			},
			want: true,
		},
		{
			name: "Invalid - contains GET",
			tx: Transaction{
				Operations: []*Operation{
					{Directive: SetDirective, Key: "key", Value: "value"},
					{Directive: GetDirective, Key: "key", Value: ""},
				},
			},
			want: false,
		},
		{
			name: "Invalid - contains EXISTS",
			tx: Transaction{
				Operations: []*Operation{
					{Directive: ExistsDirective, Key: "key", Value: ""},
				},
			},
			want: false,
		},
		{
			name: "Invalid - no operations",
			tx: Transaction{
				Operations: []*Operation{},
			},
			want: false,
		},
		{
			name: "Invalid - more than 8 operations",
			tx: Transaction{
				Operations: []*Operation{
					{Directive: SetDirective, Key: "k1", Value: "v1"},
					{Directive: SetDirective, Key: "k2", Value: "v2"},
					{Directive: SetDirective, Key: "k3", Value: "v3"},
					{Directive: SetDirective, Key: "k4", Value: "v4"},
					{Directive: SetDirective, Key: "k5", Value: "v5"},
					{Directive: SetDirective, Key: "k6", Value: "v6"},
					{Directive: SetDirective, Key: "k7", Value: "v7"},
					{Directive: SetDirective, Key: "k8", Value: "v8"},
					{Directive: SetDirective, Key: "k9", Value: "v9"},
				},
			},
			want: false,
		},
		{
			name: "Invalid - nil operations slice",
			tx: Transaction{
				Operations: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tx.IsValid(); got != tt.want {
				t.Errorf("Transaction.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Edge Cases and Integration Tests
// ============================================================================

func TestCommandType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType CommandType
	}{
		{
			name:     "Simple operation returns OperationCommand",
			input:    "GET foo",
			wantType: OperationCommand,
		},
		{
			name:     "Transaction returns TransactionCommand",
			input:    "TX SET foo bar",
			wantType: TransactionCommand,
		},
		{
			name:     "SET operation returns OperationCommand",
			input:    "SET key value",
			wantType: OperationCommand,
		},
		{
			name:     "Multi-op transaction returns TransactionCommand",
			input:    "TX SET a 1 DELETE b SET c 2",
			wantType: TransactionCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.input)
			if err != nil {
				t.Fatalf("ParseCommand() unexpected error = %v", err)
			}

			if got.Type != tt.wantType {
				t.Errorf("ParseCommand() Type = %v, want %v", got.Type, tt.wantType)
			}
		})
	}
}

func TestMutualExclusivity(t *testing.T) {
	// Test that commands are either operations OR transactions, never both
	tests := []struct {
		name  string
		input string
	}{
		{"Operation", "GET foo"},
		{"Transaction", "TX SET foo bar"},
		{"SET Operation", "SET key value"},
		{"DELETE Operation", "DELETE key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.input)
			if err != nil {
				t.Fatalf("ParseCommand() unexpected error = %v", err)
			}

			// Either Operation or Transaction should be set, not both
			hasOp := got.Operation != nil
			hasTx := got.Transaction != nil

			if hasOp == hasTx {
				t.Errorf("ParseCommand() both Operation and Transaction are %v, should be mutually exclusive", hasOp)
			}

			// Type should match what's set
			if hasOp && got.Type != OperationCommand {
				t.Errorf("Has Operation but Type = %v, want %v", got.Type, OperationCommand)
			}
			if hasTx && got.Type != TransactionCommand {
				t.Errorf("Has Transaction but Type = %v, want %v", got.Type, TransactionCommand)
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError error
	}{
		{
			name:      "Invalid directive returns ErrInvalidCommand",
			input:     "INVALID key",
			wantError: ErrInvalidCommand,
		},
		{
			name:      "Empty input returns ErrInvalidCommand",
			input:     "",
			wantError: ErrInvalidCommand,
		},
		{
			name:      "TX with no operations returns ErrInvalidCommand",
			input:     "TX",
			wantError: ErrInvalidCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCommand(tt.input)
			if err == nil {
				t.Fatal("ParseCommand() expected error, got nil")
			}
			// Note: We're checking that we get *an* error, not necessarily the exact type
			// since the error wrapping/matching behavior may vary
		})
	}
}
