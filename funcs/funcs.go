package funcs

import (
	"html/template"
)

func DefaultMap(name string, props map[string]any) template.FuncMap {
	return template.FuncMap{
		// template execution
		"props": NewKVSProps,
	}
}
