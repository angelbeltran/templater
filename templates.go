// The Templater type is the core of the package.
// To use this package, a directory holdings all templates must exist.
// That directory, will typically hold the following structure, but is configurable.
//   - /layout.html.tmpl
//   - /pages/
//   - /components/
//
// The layout.html.tmpl file holds the general webpage layout, with a root
// <html> element, itself with <head> and <body> elements.
// The only requirement is the <body> must define a "body" block.
// While the <head> will typically define a "head" block, but is not required.
// A minimal example is the following.
//
// <!DOCTYPE html>
// <html>
//
//	<head>
//		<!-- a common entrypoint for <head/> content but not used by this library -->
//	    {{ block "head" . }} {{ end }}
//	</head>
//	<body>
//		<!-- REQUIRED: where the /page/ template content will be placed -->
//	    {{ block "body" . }} {{ end }}
//	</body>
//
// </html>
//
// All templates must have the file extension .html.tmpl, unless configured otherwise.
//
// The /pages/ directory holds all templates serving the "body"
// of standalone webpages.
// They may be compiled and executed via Templater.ExecutePage.
// These templates may reuse components defined in /components/.
//
// The /components/ directory holds all templates intended for use as
// components, usable in any page or other component (even in themselves!).
//
// To use a component in a page or other component, use the
// `component` function.
// It's provided by Templater.ExecutePage and Templater.ExecuteComponent.
// It requires the name of the component - name of the file in
// /components/ minus the .html.tmpl file extension.
// It accepts a sequence of key-value pairs describing the "props" provided
// to the component, the odd arguments being key strings, and the even
// arguments being the values.
// These props will be passed as a map[string]any to the component template.
// These props are not required.
// Example:
//
// {{ component "header" "title" "My Website" "subtitle" "Another Pet Project" }}
//
// This would compile the component at /components/header.html.tmpl
// with the single props title = "My Website" and subtitle = "Another Pet Project".
//
// The usage of `component` within templates allows the composing of component
// templates into larger components and webpages in a manner that is more modular.
//
// Additional template functions provided are
// - props: constructs a props map[string]any in the many used by component.
package templater

import (
	"bytes"
	"fmt"
	"html/template"
	"maps"
	"os"
	"path"

	"github.com/angelbeltran/templater/funcs"
)

type (
	Templater struct {
		cfg Config
	}

	Config struct {
		Funcs   func() template.FuncMap
		Dirs    DirsConfig
		FileExt string
	}

	DirsConfig struct {
		Base       string
		Pages      string
		Components string
	}
)

func (tm *Templater) With(cfg Config) *Templater {
	tm.cfg = cfg
	tm.cfg.setDefaultsToZeroFields()
	return tm
}

func (c *Config) setDefaultsToZeroFields() {
	if c.Funcs == nil {
		c.Funcs = funcs.DefaultMap
	}

	c.Dirs.setDefaultsToZeroFields()

	if c.FileExt == "" {
		c.FileExt = ".html.tmpl"
	}
	if c.FileExt[0] != '.' {
		c.FileExt = "." + c.FileExt
	}
}

func (c *DirsConfig) setDefaultsToZeroFields() {
	if c.Base == "" {
		c.Base = "templates"
	}
	if c.Pages == "" {
		c.Pages = "pages"
	}
	if c.Components == "" {
		c.Components = "components"
	}
}

// ExecutePage is basically ExecuteComponent except returns html wrapped up in the layout page.
func (tm *Templater) ExecutePage(name string, kvs ...any) ([]byte, error) {
	props, err := funcs.NewKVSProps(kvs...)
	if err != nil {
		return nil, err
	}

	// parse the layout template

	layoutFilename := "layout" + tm.cfg.FileExt

	layout, err := template.New(layoutFilename).
		Funcs(tm.buildComponentFuncMap()).
		ParseFiles(path.Join(tm.cfg.Dirs.Base, layoutFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to parse layout html file: %w", err)
	}

	// define "body" template

	if b, err := os.ReadFile(path.Join(tm.cfg.Dirs.Base, tm.cfg.Dirs.Pages, name+tm.cfg.FileExt)); err != nil {
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

// ExecuteComponent allows for dynamic template lookup and execution
// It expects an even number of kvs (allows for zero).
// They are treated as key-value pairs and passed in a map[string]any to the template.
func (tm *Templater) ExecuteComponent(name string, kvs ...any) ([]byte, error) {
	props, err := funcs.NewKVSProps(kvs...)
	if err != nil {
		return nil, err
	}

	filename := name + tm.cfg.FileExt

	t, err := template.New(name).
		Funcs(tm.buildComponentFuncMap()).
		ParseFiles(path.Join(tm.cfg.Dirs.Base, tm.cfg.Dirs.Components, filename))
	if err != nil {
		return nil, fmt.Errorf("failed to parse component %s: %w", name, err)
	}

	buf := new(bytes.Buffer)
	if err := t.ExecuteTemplate(buf, path.Base(filename), props); err != nil {
		return nil, fmt.Errorf("failed to execute component %s: %w", name, err)
	}

	return buf.Bytes(), nil
}

func (tm *Templater) buildComponentFuncMap() template.FuncMap {
	m := template.FuncMap(map[string]any{
		// template execution
		"component": func(name string, props ...any) (template.HTML, error) {
			b, err := tm.ExecuteComponent(name, props...)
			return template.HTML(b), err
		},
	})

	maps.Copy(m, funcs.DefaultMap())
	maps.Copy(m, tm.cfg.Funcs())

	return m
}
