package commandparser

import (
	"errors"
	"strings"
)

// parses the command string and returns the command keyword and parameters
func ParseCommand(command string) (string, []string, error) {
	// Trim leading and trailing whitespaces
	command = strings.TrimSpace(command)

	// Split the command into words
	words := strings.Fields(command)
	if len(words) == 0 {
		return "", nil, errors.New("empty command")
	}

	// Extract the command keyword and parameters
	cmd := strings.ToUpper(words[0])
	params := words[1:]

	// Check the validity of the command and parameters
	if err := validateCommand(cmd, params); err != nil {
		return "", nil, err
	}

	return cmd, params, nil
}

// validateCommand checks the validity of the command and parameters
func validateCommand(cmd string, params []string) error {
	switch cmd {
	case "SET":
		if len(params) < 2 {
			return errors.New("invalid command")
		}
	case "GET":
		if len(params) != 1 {
			return errors.New("invalid command")
		}
	case "QPUSH":
		if len(params) < 2 {
			return errors.New("invalid command")
		}
	case "QPOP":
		if len(params) != 1 {
			return errors.New("invalid command")
		}
	case "BQPOP":
		if len(params) != 2 {
			return errors.New("invalid command")
		}
	default:
		return errors.New("invalid command")
	}

	return nil
}
