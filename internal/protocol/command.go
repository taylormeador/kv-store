package protocol

import (
	"bufio"
	"errors"
	"log"
	"strings"
)

var ErrInvalidCommand error = errors.New("invalid command")

type Command struct {
	Type        CommandType
	Operation   *Operation
	Transaction *Transaction
}

type CommandType string

const (
	OperationCommand   CommandType = "OPERATION"
	TransactionCommand CommandType = "TRANSACTION"
)

// Parse the input into a Command, return error if it is not valid.
func ParseCommand(input string) (*Command, error) {
	c := &Command{}

	scanner := bufio.NewScanner(strings.NewReader(input))
	if scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)

		if len(words) > 1 {
			directive := Directive(words[0])
			if !directive.IsValid() {
				return nil, ErrInvalidCommand
			}
			switch directive {
			case TxDirective:
				c.Type = TransactionCommand
				c.Transaction = &Transaction{}
				if err := parseTransaction(words, c.Transaction); err != nil {
					return nil, err
				}
			default:
				c.Type = OperationCommand
				c.Operation = &Operation{}
				if err := parseSingleOperation(words, c.Operation); err != nil {
					return nil, err
				}
			}
		} else {
			return nil, ErrInvalidCommand
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return c, nil
}

// parseNextOperation parses the first operation (if it exists) into the Operation and returns.
// It ignores anything after the first Operation, meaning it can be called recursively for Transactions.
// It returns a slice of strings which are the remaining unparsed words.
// If it does not return an error, then the operation is valid.
func parseNextOperation(words []string, o *Operation) ([]string, error) {
	if len(words) > 0 {
		// A single word is not a valid command
		if len(words) == 1 {
			return nil, ErrInvalidOperation
		}

		for i, word := range words {
			switch i {
			case 0:
				// The first word of an Operation must be a Directive other than `TX`
				o.Directive = Directive(word)
				if !o.Directive.IsValid() || o.Directive == TxDirective {
					return nil, ErrInvalidOperation
				}
			case 1:
				// The second word is the key, and is the last word for all but `SET`
				o.Key = word
				if o.Directive != SetDirective {
					if !o.IsValid() {
						return nil, ErrInvalidOperation
					}
					if i+1 <= len(words)-1 {
						return words[i+1:], nil
					} else {
						return []string{}, nil
					}
				}
			case 2:
				// The third word is the value for `SET` and is always the final word
				o.Value = word
				if !o.IsValid() {
					return nil, ErrInvalidOperation
				}
				if i+1 <= len(words)-1 {
					return words[i+1:], nil
				} else {
					return []string{}, nil
				}
			}
		}
	}
	return words, nil
}

// Parses a single Operation
func parseSingleOperation(words []string, o *Operation) error {
	remainingWords, err := parseNextOperation(words, o)
	if err != nil {
		return err
	}
	if len(remainingWords) > 0 {
		return ErrInvalidOperation
	}
	return nil
}

// parseTransaction populates the tx.Operations slice with valid operations
func parseTransaction(words []string, tx *Transaction) error {
	// Any valid Transaction will have at least "TX <Directive> <Key>" and will be length > 2
	if len(words) > 2 {

		var err error

		// We already know the first word is "TX"
		words = words[1:]
		for {
			nextOperation := &Operation{}
			words, err = parseNextOperation(words, nextOperation)
			if err != nil {
				return err
			}
			tx.Operations = append(tx.Operations, nextOperation)
			if len(words) == 0 {
				if !tx.IsValid() {
					log.Println("DEBUG parsing tx invalid Tx")
					return ErrInvalidCommand
				}
				log.Printf("transaction parsed: %v", tx)
				return nil
			}
		}
	}
	return ErrInvalidCommand
}
