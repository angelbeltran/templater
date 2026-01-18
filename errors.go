package templater

import "fmt"

type (
	// ErrNotTemplateFileFound occurs when the template was not found
	ErrNotTemplateFileFound struct {
		Dir      string
		Filename string
	}

	// ErrInvalidWildcardValue is returned when an issue with parsing wildcard parameters occurs
	ErrInvalidWildcardValue struct {
		Value string
		Type  string
		Err   error
	}
)

func (e *ErrNotTemplateFileFound) Error() string {
	return fmt.Sprintf("no template file found in the directory %s matching the filename %s", e.Dir, e.Filename)
}

func (e *ErrInvalidWildcardValue) Error() string {
	return fmt.Sprintf("invalid wildcard value %q of type %s: %v", e.Value, e.Type, e.Err)
}

func (e *ErrInvalidWildcardValue) Unwrap() error {
	return e.Err
}

func (e *ErrInvalidWildcardValue) errorf(format string, args ...any) *ErrInvalidWildcardValue {
	e.Err = fmt.Errorf(format, args...)
	return e
}

func (e *ErrInvalidWildcardValue) wrap(err error) *ErrInvalidWildcardValue {
	e.Err = err
	return e
}
