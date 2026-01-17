package funcs

import (
	"html/template"
	"maps"
)

type MapBuilderFunc = func(name string, props map[string]any) template.FuncMap

func DefaultMap(name string, props map[string]any) template.FuncMap {
	return template.FuncMap{
		// template execution
		"props": NewKVSProps,
	}
}

func Chain(fns ...MapBuilderFunc) MapBuilderFunc {
	return func(name string, props map[string]any) template.FuncMap {
		m := make(template.FuncMap)
		for _, fn := range fns {
			maps.Copy(m, fn(name, props))
		}
		return m
	}
}
