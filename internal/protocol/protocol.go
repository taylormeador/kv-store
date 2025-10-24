package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"os"
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
		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
	}

	return c, nil
}
