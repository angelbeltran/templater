package templater

import "fmt"

type ErrNotTemplateFileFound struct {
	Dir      string
	Filename string
}

func (e *ErrNotTemplateFileFound) Error() string {
	return fmt.Sprintf("no template file found in the directory %s matching the filename %s", e.Dir, e.Filename)
}
