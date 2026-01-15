package funcs

import (
	"html/template"
)

func DefaultMap() template.FuncMap {
	return template.FuncMap{
		// template execution
		"props": NewKVSProps,
	}
}
