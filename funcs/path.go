package funcs

import (
	"path"
	"strings"
)

func GetPathParameters(pattern, targetPath string) (params map[string]string, match bool) {
	ext := GetExtendedExtension(pattern)
	targetPathExt := GetExtendedExtension(targetPath)
	if ext != targetPathExt {
		return nil, false
	}

	patternWithoutExt := pattern[:len(pattern)-len(ext)]
	targetPathWithoutExt := targetPath[:len(targetPath)-len(ext)]

	patternSegments := GetPathSegments(patternWithoutExt)
	pathSegments := GetPathSegments(targetPathWithoutExt)

	if len(patternSegments) != len(pathSegments) {
		return nil, false
	}

	params = make(map[string]string, len(patternSegments))
	for i, s := range patternSegments {
		isWildcard := len(s) > 2 && s[0] == '{' && s[len(s)-1] == '}'
		if isWildcard {
			wildcard := s[1 : len(s)-1]
			value := pathSegments[i]
			params[wildcard] = value
		} else if exactMatch := pathSegments[i] == s; !exactMatch {
			return nil, false
		}
	}

	return params, true
}

func GetPathSegments(p string) []string {
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

func GetExtendedExtension(filename string) string {
	var res string
	for {
		ext := path.Ext(filename)
		if ext == "" || ext == filename {
			return res
		}

		filename = filename[:len(filename)-len(ext)]
		res = ext + res
	}
}
