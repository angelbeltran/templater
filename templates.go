// The Templater type is the core of the package.
// To use this package, a directory holdings all templates must exist.
// That directory, must hold the following structure.
//   - /layout.html.tmpl
//   - /page_bodies/
//   - /page_heads/
//   - /component_bodies/
//   - /component_heads/
//
// The layout.html.tmpl file holds the general webpage layout, with a root
// <html> element, itself with <head> and <body> elements.
// The <head> must define a "head" block, and the <body> a "body" block.
// A minimal example is the following.
//
// <!DOCTYPE html>
// <html>
//
//	<head>
//	    {{ block "head" . }} {{ end }}
//	</head>
//	<body>
//	    {{ block "body" . }} {{ end }}
//	</body>
//
// </html>
//
// All other such templates must have the file extension .html.tmpl.
//
// The /page_bodies/ directory holds all templates serving the "body"
// of standalone webpages.
// They may be compiled and executed via Templater.ExecutePage.
// These templates may reuse components defined in /component_bodies/.
//
// The /component_bodies/ directory holds all templates intended for use as
// components, usable in any page or other component (even in themselves!).
//
// The /page_heads/ and /component_heads/ directories hold html intended to
// be placed in the <head> of the page alongside the respective page or
// component in /page_bodies/ or /component_bodies/, respectively.
// When compiling a page via Templater.ExecutePage, the need not be a
// corresponding file in /page_heads/ - it is optional.
//
// To use a component in a page body or other component body, use the
// `componentBody` function.
// It's provided by Templater.ExecutePage and Templater.ExecuteComponentBody.
// It requires the name of the component - name of the file in
// /component_bodies/ without the .html.tmpl file extension.
// It accepts a sequence of key-value pairs describing the "props" provided
// to the component, the odd arguments being key strings, and the even
// arguments being the values.
// These props will be passed as a map[string]any to the component template.
// These props are not required.
// Example:
//
// {{ componentBody "header" "title" "My Website" "subtitle" "Another Pet Project" }}
//
// This would compile the component at /component_bodies/header.html.tmpl
// with the single props title = "My Website" and subtitle = "Another Pet Project".
//
// Just as you can build pages using components, you may need to import the
// respective component <head/> elements, e.g. for stylesheets or scripts.
// To accomplish this, within the /page_heads/ file with the same name as the
// page template in /page_bodies/ use `componentHead` in the same manner as
// `componentBody`.
// Example:
//
// {{ componentHead "header" "title" "My Website" "subtitle" "Another Pet Project" }}
//
// This will compile the template at /component_heads/header.html.tmpl
// with the same props as in the componentBody example.
// Templates within /component_heads/ may also use componentHead to include
// <head/> elements potentially needed by the components embedded via
// componentBody.
// The componentHead function will eliminate duplicate <head/> elements
// when possible.
//
// The usage of `componentBody` and `componentHead` together within templates
// allow the composing of component templates into larger components and
// webpages in a manner that is more modular.
//
// Additional template functions provided are
// - props: constructs a props map[string]any in the many used by componentBody.
package templater

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"maps"
	"os"
	"path"
)

type Templater struct {
	templatesDir string
	funcs        func() template.FuncMap
}

func NewTemplater(templatesDir string, funcs func() template.FuncMap) *Templater {
	return &Templater{
		templatesDir: templatesDir,
		funcs:        funcs,
	}
}

// ExecutePage is basically ExecuteComponentBody except returns html wrapped up in the layout page.
func (tm *Templater) ExecutePage(name string, kvs ...any) ([]byte, error) {
	props, err := NewKVSProps(kvs...)
	if err != nil {
		return nil, err
	}

	const layoutFilename = "layout.html.tmpl"

	layout, err := template.New(layoutFilename).
		Funcs(tm.buildPageFuncMap()).
		ParseFiles(path.Join(tm.templatesDir, layoutFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to parse layout html file: %w", err)
	}

	// define "head" template

	if b, err := os.ReadFile(path.Join(tm.templatesDir, "page_heads", name+".html.tmpl")); err != nil {
		// head template isn't required to exist, only body template.
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to read page head html file: %w", err)
		}
	} else {
		if _, err := layout.New("head").Parse(string(b)); err != nil {
			return nil, fmt.Errorf("failed to parse head html template: %w", err)
		}
	}

	// define "body" template

	if b, err := os.ReadFile(path.Join(tm.templatesDir, "page_bodies", name+".html.tmpl")); err != nil {
		return nil, fmt.Errorf("failed to read page body html file: %w", err)
	} else {
		if _, err := layout.New("body").Parse(string(b)); err != nil {
			return nil, fmt.Errorf("failed to parse body html template: %w", err)
		}
	}

	buf := new(bytes.Buffer)
	if err := layout.Execute(buf, props); err != nil {
		return nil, fmt.Errorf("failed to execute html template: %w", err)
	}

	return buf.Bytes(), nil
}

// ExecuteComponentBody allows for dynamic template lookup and execution
// It expects an even number of kvs (allows for zero).
// They are treated as key-value pairs and passed in a map[string]any to the template.
func (tm *Templater) ExecuteComponentBody(name string, kvs ...any) ([]byte, error) {
	props, err := NewKVSProps(kvs...)
	if err != nil {
		return nil, err
	}

	filename := name + ".html.tmpl"

	t, err := template.New(name).
		Funcs(tm.buildComponentBodyFuncMap()).
		ParseFiles(path.Join(tm.templatesDir, "component_bodies", filename))
	if err != nil {
		return nil, fmt.Errorf("failed to parse component %s: %w", name, err)
	}

	buf := new(bytes.Buffer)
	if err := t.ExecuteTemplate(buf, path.Base(filename), props); err != nil {
		return nil, fmt.Errorf("failed to execute component %s: %w", name, err)
	}

	return buf.Bytes(), nil
}

func (tm *Templater) executeComponentHead(executeSubComponentHead componentExecutorFunc, name string, kvs ...any) ([]byte, error) {
	props, err := NewKVSProps(kvs...)
	if err != nil {
		return nil, err
	}

	filename := name + ".html.tmpl"

	t, err := template.New(name).
		Funcs(tm.buildComponentHeadFuncMap(executeSubComponentHead)).
		ParseFiles(path.Join(tm.templatesDir, "component_heads", filename))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to parse component head %s: %w", name, err)
	}

	buf := new(bytes.Buffer)
	if err := t.ExecuteTemplate(buf, filename, props); err != nil {
		return nil, fmt.Errorf("failed to execute component head %s: %w", name, err)
	}

	return buf.Bytes(), nil
}

func (tm *Templater) buildPageFuncMap() template.FuncMap {
	componentHeadPropsSeen := make(map[string][][]any)
	componentHeadSeen := make(map[string]bool)

	var uniqueComponentHeadExecutor func(name string, props ...any) (template.HTML, error)
	uniqueComponentHeadExecutor = func(name string, props ...any) (template.HTML, error) {
		if componentHeadSeen[name] {
			// componentHeads should not be duplicated, if possible.
			for _, propsSeen := range componentHeadPropsSeen[name] {
				if len(props) != len(propsSeen) {
					continue
				}

				match := true
				for i := range props {
					if props[i] != propsSeen[i] {
						match = false
						break
					}
				}

				if match {
					return "", nil
				}
			}

			// never seen this combination of componentHead name and props
		}
		componentHeadSeen[name] = true
		componentHeadPropsSeen[name] = append(componentHeadPropsSeen[name], props)

		b, err := tm.executeComponentHead(uniqueComponentHeadExecutor, name, props...)
		return template.HTML(b), err
	}

	funcs := template.FuncMap(map[string]any{
		// template execution
		"componentHead": uniqueComponentHeadExecutor,
		"componentBody": func(name string, props ...any) (template.HTML, error) {
			b, err := tm.ExecuteComponentBody(name, props...)
			return template.HTML(b), err
		},
	})

	maps.Copy(funcs, tm.commonFuncs())

	return funcs
}

func (tm *Templater) buildComponentBodyFuncMap() template.FuncMap {
	funcs := template.FuncMap(map[string]any{
		// template execution
		"componentBody": func(name string, props ...any) (template.HTML, error) {
			b, err := tm.ExecuteComponentBody(name, props...)
			return template.HTML(b), err
		},
	})

	maps.Copy(funcs, tm.commonFuncs())

	return funcs
}

type componentExecutorFunc = func(name string, props ...any) (template.HTML, error)

func (tm *Templater) buildComponentHeadFuncMap(componentHead componentExecutorFunc) template.FuncMap {
	funcs := template.FuncMap(map[string]any{
		// template execution
		"componentHead": componentHead,
	})

	maps.Copy(funcs, tm.commonFuncs())

	return funcs
}

func (tm *Templater) commonFuncs() template.FuncMap {
	funcs := template.FuncMap(map[string]any{
		// template execution
		"props": NewKVSProps,
	})

	maps.Copy(funcs, tm.funcs())

	return funcs
}

// NewKVSProps is the implementation of the `props` template function.
func NewKVSProps(args ...any) (map[string]any, error) {
	if len(args)%2 == 1 {
		return nil, fmt.Errorf("the props function expects an even number of arguments, key-value pairs: received %d arguments", len(args))
	}

	props := make(map[string]any, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		k, ok := args[i].(string)
		if !ok {
			return nil, fmt.Errorf("props expected odd arguments to be key strings: argument %d was a %T", i+1, args[i])
		}

		props[k] = args[i+1]
	}

	return props, nil
}
