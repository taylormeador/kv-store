package protocol

import (
	"bufio"
	"errors"
	"strings"
)

var ErrInvalidCommand error = errors.New("invalid command")

type Command struct {
	Directive Directive
	Key       string
	Value     string
}

// Define Directive type using Enum pattern
type Directive string

const (
	GetDirective    Directive = "GET"
	SetDirective    Directive = "SET"
	DeleteDirective Directive = "DELETE"
	ExistsDirective Directive = "EXISTS"
)

func (d Directive) IsValid() bool {
	switch d {
	case GetDirective, SetDirective, DeleteDirective, ExistsDirective:
		return true
	default:
		return false
	}
}

// Parse the input into a Command
func ParseCommand(input string) (*Command, error) {
	c := &Command{}

	// Split input into words
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)

	// Loop through words and parse into Command
	count := 0
	for scanner.Scan() {
		word := scanner.Text()
		switch count {
		case 0:
			c.Directive = Directive(word)
			if !c.Directive.IsValid() {
				return nil, ErrInvalidCommand
			}
			count++
		case 1:
			c.Key = word
			count++
		case 2:
			c.Value = word
			count++
		default:
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Validate argument counts
	switch c.Directive {
	case GetDirective, DeleteDirective, ExistsDirective:
		if count != 2 {
			return nil, ErrInvalidCommand
		}
	case SetDirective:
		if count != 3 {
			return nil, ErrInvalidCommand
		}
	default:
		// Empty input or no valid directive parsed
		return nil, ErrInvalidCommand
	}

	return c, nil
}

// Returns a string of a valid Command
func (c *Command) String() string {
	result := string(c.Directive) + " " + c.Key
	if c.Directive == SetDirective {
		result += " " + c.Value
	}
	return result
}
