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
//
// Additionally, path wildcards of the form {.*} are supported.
// For example, given a component file /component/buttons/{id}/id-button.html.tmpl
//
//	 <button>
//		  {{ .PathParams.id }
//	 </button>
//
// Calling ExecuteComponent with "buttons/123/id-button" will compile to
//
//	 <button>
//		  123
//	 </button>
//
// Similar behavior is provided in ExecutePage.
package templater

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/angelbeltran/templater/funcs"
)

type (
	Templater struct {
		cfg Config
	}

	Config struct {
		Funcs   func(name string, props map[string]any) template.FuncMap
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

func (tm *Templater) WithFuncs(m template.FuncMap) *Templater {
	cpy := *tm
	cpy.cfg.Funcs = func(name string, props map[string]any) template.FuncMap {
		dst := make(template.FuncMap)
		maps.Copy(dst, m)
		maps.Copy(dst, tm.cfg.Funcs(name, props))
		return dst
	}
	return &cpy
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

	return tm.executePage(name, props)
}

func (tm *Templater) executePage(name string, props map[string]any) ([]byte, error) {
	// find a matching file, and parse the path parameters

	filename := name + tm.cfg.FileExt
	pageDir := path.Join(tm.cfg.Dirs.Base, tm.cfg.Dirs.Pages)

	match, err := findBestFilenameMatchInDir(name, tm.cfg.FileExt, pageDir)
	if err != nil {
		return nil, err
	}

	props["PathParams"], _, err = getPathParameters(match, filename)
	if err != nil {
		return nil, err
	}

	// parse the layout template

	layoutFilename := "layout" + tm.cfg.FileExt

	layout, err := template.New(layoutFilename).
		Funcs(tm.buildFuncMap(name, props)).
		ParseFiles(path.Join(tm.cfg.Dirs.Base, layoutFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to parse layout html file: %w", err)
	}

	// define "body" template

	if b, err := os.ReadFile(path.Join(tm.cfg.Dirs.Base, tm.cfg.Dirs.Pages, match)); err != nil {
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

	return tm.executeComponent(name, props)
}

func (tm *Templater) executeComponent(name string, props map[string]any) ([]byte, error) {
	filename := name + tm.cfg.FileExt
	componentDir := path.Join(tm.cfg.Dirs.Base, tm.cfg.Dirs.Components)

	match, err := findBestFilenameMatchInDir(name, tm.cfg.FileExt, componentDir)
	if err != nil {
		return nil, err
	}

	pathParams, _, err := getPathParameters(match, filename)
	if err != nil {
		return nil, err
	}

	props["PathParams"] = pathParams

	t, err := template.New(name).
		Funcs(tm.buildFuncMap(name, props)).
		ParseFiles(path.Join(componentDir, match))
	if err != nil {
		return nil, fmt.Errorf("failed to parse component %s: %w", name, err)
	}

	buf := new(bytes.Buffer)
	if err := t.ExecuteTemplate(buf, path.Base(match), props); err != nil {
		return nil, fmt.Errorf("failed to execute component %s: %w", name, err)
	}

	return buf.Bytes(), nil
}

// findBestFilenameMatchInDir finds the most exact match for a filename, allowing for path segments wildcards for the form {\w+}.
// supports index.html files.
func findBestFilenameMatchInDir(filenameBase, ext, dir string) (string, error) {
	filename := filenameBase + ext
	filenameBaseSegments := getPathSegments(filenameBase)

	var matchesFound [][]string

	err := fs.WalkDir(os.DirFS(dir), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		pWithoutExt := p
		if !d.IsDir() && strings.HasSuffix(pWithoutExt, ext) {
			pWithoutExt = pWithoutExt[:len(pWithoutExt)-len(ext)]
		}

		segments := getPathSegments(pWithoutExt)
		expectMatchingFileOrParentDir := len(segments) == len(filenameBaseSegments)
		expectIndexFile := len(segments) == (len(filenameBaseSegments) + 1)

		switch {
		case expectIndexFile:
			if d.IsDir() {
				return fs.SkipDir
			}
		case expectMatchingFileOrParentDir:
		default:
			if !d.IsDir() {
				return nil
			}
		}

		for i, seg := range segments {
			if i < len(filenameBaseSegments) && filenameBaseSegments[i] == seg {
				continue
			}

			isLastSegment := i == len(segments)-1
			if isLastSegment {
				if expectIndexFile && seg == "index" {
					continue
				}
			}

			base := seg
			if isLastSegment && expectMatchingFileOrParentDir && !d.IsDir() && strings.HasSuffix(seg, ext) {
				base = seg[:len(seg)-len(ext)]
			}
			isWildCard := len(base) > 2 && base[0] == '{' && base[len(base)-1] == '}'

			if isWildCard {
				continue
			}

			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if !d.IsDir() && (expectMatchingFileOrParentDir || expectIndexFile) {
			matchesFound = append(matchesFound, segments)
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk the template directory: %w", err)
	}

	if len(matchesFound) == 0 {
		return "", &ErrNotTemplateFileFound{
			Dir:      dir,
			Filename: filename,
		}
	}

	matchingFilenameSegments := make([]string, len(filenameBaseSegments), len(filenameBaseSegments)+1)
	tree := buildSegmentTree(matchesFound...)
	branch := tree
	for i, seg := range filenameBaseSegments {
		if st, exactMatch := branch[seg]; exactMatch {
			matchingFilenameSegments[i] = seg
			branch = st
		} else if l := len(branch); l > 1 {
			return "", fmt.Errorf("multiple wildcard branches found while looking for matching file for %s at %s: %d", filename, dir, l)
		} else {
			// there should only be a single branch
			for wildcard, st := range branch {
				matchingFilenameSegments[i] = wildcard
				branch = st
			}
		}
	}

	if st, ok := branch["index"]; ok {
		branch = st
		matchingFilenameSegments = append(matchingFilenameSegments, "index")
	}

	return strings.Join(matchingFilenameSegments, "/") + ext, nil
}

type segmentTree map[string]segmentTree

func buildSegmentTree(pathSegmentList ...[]string) segmentTree {
	if len(pathSegmentList) == 0 {
		return make(segmentTree)
	}

	tree := buildSegmentTree(pathSegmentList[1:]...)

	node := tree
	for _, seg := range pathSegmentList[0] {
		subnode, ok := node[seg]
		if !ok {
			subnode = make(segmentTree)
			node[seg] = subnode
		}
		node = subnode
	}

	return tree
}

// getWildcardPathCombinations respected filename extensions
func getWildcardPathCombinations(filename string) []string {
	matchingPathSegments := getWildcardPathSegmentCombinations(getPathSegments(filename))

	var precedingSlash string
	if len(filename) > 0 && filename[0] == '/' {
		precedingSlash = "/"
	}

	var trailingSlash string
	if len(filename) > 0 && filename[len(filename)-1] == '/' {
		trailingSlash = "/"
	}

	paths := make([]string, len(matchingPathSegments))
	for i, segments := range matchingPathSegments {
		paths[i] = precedingSlash + strings.Join(segments, "/") + trailingSlash
	}

	return paths
}

// getWildcardPathSegmentCombinations respected filename extensions
func getWildcardPathSegmentCombinations(segments []string) [][]string {
	const wildcard = "{.*}"

	switch len(segments) {
	case 0:
		return nil
	case 1:
		wildcardSegment := wildcard
		if ext := path.Ext(segments[0]); ext != "" {
			wildcardSegment += ext
		}

		return [][]string{
			[]string{segments[0]},
			[]string{wildcardSegment},
		}
	default:
		head := segments[0]
		tail := segments[1:]

		tailCombinations := getWildcardPathSegmentCombinations(tail)

		combinations := make([][]string, len(tailCombinations)*2)
		for i, comb := range tailCombinations {
			combinations[i*2] = append([]string{head}, comb...)
			combinations[(i*2)+1] = append([]string{wildcard}, comb...)
		}

		return combinations
	}
}

func (tm *Templater) buildFuncMap(name string, props map[string]any) template.FuncMap {
	m := template.FuncMap(map[string]any{
		// template execution
		"component": func(name string, kvs ...any) (template.HTML, error) {
			additionalProps, err := funcs.NewKVSProps(kvs...)
			if err != nil {
				return "", err
			}

			cpy := make(map[string]any, len(props))
			maps.Copy(cpy, props)
			maps.Copy(cpy, additionalProps)

			b, err := tm.executeComponent(name, cpy)
			return template.HTML(b), err
		},
	})

	maps.Copy(m, funcs.DefaultMap(name, props))
	maps.Copy(m, tm.cfg.Funcs(name, props))

	return m
}

func getPathSegments(p string) []string {
	p = path.Clean(p)
	if p == "" || p == "." {
		return nil
	}

	if p[0] == '/' {
		p = p[1:]
	}
	if p == "" {
		return nil
	}

	if p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}

	return strings.Split(p, "/")
}

func getPathParameters(pattern, targetPath string) (params map[string]any, match bool, err error) {
	ext := getExtendedExtension(pattern)
	targetPathExt := getExtendedExtension(targetPath)
	if ext != targetPathExt {
		return nil, false, nil
	}

	patternWithoutExt := pattern[:len(pattern)-len(ext)]
	targetPathWithoutExt := targetPath[:len(targetPath)-len(ext)]

	patternSegments := getPathSegments(patternWithoutExt)
	pathSegments := getPathSegments(targetPathWithoutExt)

	if len(patternSegments) != len(pathSegments) {
		return nil, false, nil
	}

	params = make(map[string]any, len(patternSegments))
	for i, s := range patternSegments {
		isWildcard := len(s) > 2 && s[0] == '{' && s[len(s)-1] == '}'
		if isWildcard {
			wildcard := s[1 : len(s)-1]
			value := pathSegments[i]

			key, parsed, err := parseWildcard(wildcard, value)
			if err != nil {
				return nil, false, fmt.Errorf("failed to parse wildcard: %w", err)
			}

			params[key] = parsed
		} else if exactMatch := pathSegments[i] == s; !exactMatch {
			return nil, false, nil
		}
	}

	return params, true, nil
}

func parseWildcard(wildcardKey, value string) (key string, parsed any, err error) {
	parts := strings.SplitN(wildcardKey, ".", 2)
	if len(parts) == 1 {
		return wildcardKey, value, nil
	}

	parsed, err = parseWildcardValue(parts[1], value)
	return parts[0], parsed, err
}

func parseWildcardValue(typeName, value string) (parsed any, err error) {
	werr := ErrInvalidWildcardValue{
		Value: value,
		Type:  typeName,
	}

	switch typeName {
	// boolean
	case "bool":
		switch strings.ToLower(value) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, werr.errorf(`expected "true" or "false"`)
		}

	// integer
	case "int":
		n, err := strconv.ParseInt(value, 10, strconv.IntSize)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return int(n), nil
	case "int8":
		n, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return int8(n), nil
	case "int16":
		n, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return int16(n), nil
	case "int32":
		n, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return int32(n), nil
	case "int64":
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return int64(n), nil

	// unsigned integer
	case "uint":
		n, err := strconv.ParseUint(value, 10, strconv.IntSize)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return uint(n), nil
	case "uintptr":
		n, err := strconv.ParseUint(value, 10, strconv.IntSize)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return uintptr(n), nil
	case "uint8":
		n, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return uint8(n), nil
	case "uint16":
		n, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return uint16(n), nil
	case "uint32":
		n, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return uint32(n), nil
	case "uint64":
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return uint64(n), nil

	// floating pointer
	case "float32":
		n, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return float32(n), nil
	case "float64":
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return float64(n), nil

	// complex
	case "complex64":
		n, err := strconv.ParseComplex(value, 64)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return complex64(n), nil
	case "complex128":
		n, err := strconv.ParseComplex(value, 128)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return complex128(n), nil

	// bytes
	case "byte":
		n, err := strconv.ParseUint(value, 16, 8)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return byte(uint8(n)), nil
	case "rune":
		n, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, werr.wrap(err)
		}
		return rune(int32(n)), nil
	case "string":
		return value, nil

	default:
		return nil, werr.errorf("unrecognized wildcard type: %q", typeName)
	}
}

func getExtendedExtension(filename string) string {
	base := path.Base(filename)
	startsWithWildcardPrefix := len(base) > 0 && base[0] == '{'

	var res string
	for {
		ext := path.Ext(base)
		if ext == "" || ext == base || (startsWithWildcardPrefix && base[len(base)-1] == '}') {
			return res
		}

		base = base[:len(base)-len(ext)]
		res = ext + res
	}
}
