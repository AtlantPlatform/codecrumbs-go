package parser

import (
	"regexp"
	"strings"
)

type LanguageDefinition struct {
	Name       string
	Extensions []string
	Regexps    []string

	regexpsParsed []*regexp.Regexp
}

func (def *LanguageDefinition) Match(lineBytes []byte) ([]byte, bool) {
	for _, rx := range def.regexpsParsed {
		if rx.Match(lineBytes) {
			clean := rx.ReplaceAll(lineBytes, nil)
			return clean, true
		}
	}
	return nil, false
}

func LanguageFor(ext string) (*LanguageDefinition, bool) {
	if def, ok := langsCache[ext]; ok {
		return &def, ok
	}
	return nil, false
}

func init() {
	langsCache = make(map[string]LanguageDefinition)
	addLang(LanguageDefinition{
		Name: "Go",
		Extensions: []string{
			".go",
		},
		Regexps: []string{
			`^\s*//\s?`,
		},
	})
	addLang(LanguageDefinition{
		Name: "Javascript",
		Extensions: []string{
			".js", ".jsx",
		},
		Regexps: []string{
			`^\s*//\s?`,
		},
	})
	addLang(LanguageDefinition{
		Name: "Typescript",
		Extensions: []string{
			".ts",
		},
		Regexps: []string{
			`^\s*//\s?`,
		},
	})
	addLang(LanguageDefinition{
		Name: "PHP",
		Extensions: []string{
			".php",
		},
		Regexps: []string{
			`^\s*//\s?`,
		},
	})
	addLang(LanguageDefinition{
		Name: "Python",
		Extensions: []string{
			".py",
		},
		Regexps: []string{
			`^\s*#\s?`,
		},
	})
	addLang(LanguageDefinition{
		Name: "Java",
		Extensions: []string{
			".java",
		},
		Regexps: []string{
			`^\s*//\s?`,
		},
	})
	addLang(LanguageDefinition{
		Name: "C/C++",
		Extensions: []string{
			".c", ".h", ".cpp", ".cxx", ".objc", ".m",
		},
		Regexps: []string{
			`^\s*//\s?`,
		},
	})
}

func addLang(def LanguageDefinition) {
	for _, rx := range def.Regexps {
		def.regexpsParsed = append(def.regexpsParsed, regexp.MustCompile(rx))
	}
	for _, ext := range def.Extensions {
		langsCache[strings.ToLower(ext)] = def
	}
	SupportedLangs = append(SupportedLangs, def)
}

var SupportedLangs = []LanguageDefinition{}

var langsCache map[string]LanguageDefinition
