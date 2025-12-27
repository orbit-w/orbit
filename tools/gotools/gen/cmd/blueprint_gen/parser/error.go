package field_parser

import (
	"fmt"
)

func ErrFieldNumberInvalid(num int32, line string) error {
	return fmt.Errorf("field number %d is invalid, must be greater than %d: %s", num, FieldNumberMin, line)
}

func ErrFieldNumberNotFound(err error, line string) error {
	return fmt.Errorf("field number not found in: %s: %w", line, err)
}

func ErrFieldNameNotFound(line string) error {
	return fmt.Errorf("field name not found in: %s", line)
}

func ErrInvalidFieldFormat(line string) error {
	return fmt.Errorf("invalid field format: %s", line)
}

func ErrFieldTypeParseFailed(err error, typeStr string) error {
	return fmt.Errorf("failed to parse type: %w: %s", err, typeStr)
}
