package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"

	"github.com/AtlantPlatform/codecrumbs-go/generator"
	"github.com/AtlantPlatform/codecrumbs-go/parser"
	"github.com/AtlantPlatform/codecrumbs-go/renderer"
)

var app = cli.App("cc-go", "Learn, design or document codebase by putting breadcrumbs in source code.\n\n"+
	`
 _  _  __  _   ___
(__(__    (_/_(_) 
         .-/      
        (_/       
`)

var (
	projectName  = app.StringOpt("p project", "ProjectName", "Specify project prefix on GitHub (for GFM) or just a name.")
	projectDir   = app.StringOpt("d dir", "", "Project directory path containing augmented source code.")
	excludePaths = app.StringsOpt("exclude", nil, "Exclude specfic path prefixes (e.g. vendor).")
	includePaths = app.StringsOpt("include", nil, "Include path prefixes.")
	projectEntry = app.StringOpt("e entry", "", "Entrypoint file that is likely the source of main codecrumbs trails.")
	sourcePrefix = app.StringOpt("prefix", "", "Source prefix for the file paths referenced in the documentation.")
	outputFormat = app.StringOpt("f format", "markdown", "The format of output to produce. Available: markdown, json.")
	outputFile   = app.StringOpt("o out", "", "Output file path.")
)

const (
	OutputFormatMarkdown = "markdown"
	OutputFormatJSON     = "json"
)

func main() {
	app.Command("render", "Renders generated files into some representation (e.g. Markdown -> HTML)", cmdRender)
	app.Action = func() {
		if len(*projectDir) == 0 {
			log.Fatalln("project directory must be specified with -d or --dir")
		}
		if len(*projectEntry) == 0 {
			log.Warningln("project entrypoint file should be specified with -e or --entry")
		}
		switch *outputFormat {
		case OutputFormatMarkdown, OutputFormatJSON:
		default:
			log.Fatalln("unsupported output format:", *outputFormat)
		}

		crumbsList := make([][]*parser.CodeCrumb, 0)
		excludeRxs, err := compileRxs(*excludePaths)
		if err != nil {
			log.Fatalln(err)
		}
		err = filepath.Walk(*projectDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relativePath := strings.TrimPrefix(path, *projectDir)
			isIncluded := containsPrefix(relativePath, *includePaths)
			if info.IsDir() {
				if isMatching(relativePath, excludeRxs) {
					return filepath.SkipDir
				}
				return nil
			} else if isMatching(relativePath, excludeRxs) {
				return nil
			}
			if len(*includePaths) > 0 && !isIncluded {
				return nil
			}

			commentLineLang, ok := parser.LanguageFor(filepath.Ext(path))
			if !ok {
				return nil
			}
			f, err := os.Open(path)
			if err != nil {
				log.Warningln(err)
				return nil
			}
			defer f.Close()
			crumbs, err := parser.CollectCrumbs(relativePath, commentLineLang, f)
			if err != nil {
				log.WithFields(log.Fields{
					"file": relativePath,
				}).Warningln(err)
				return nil
			} else if len(crumbs) > 0 {
				crumbsList = append(crumbsList, crumbs)
			}
			return nil
		})
		if err != nil {
			log.Fatalln(err)
		}
		groups := regroupCodeCrumbs(*projectEntry, crumbsList)

		var buf []byte
		switch *outputFormat {
		case OutputFormatJSON:
			buf, err = json.MarshalIndent(groups, "", "\t")
			if err != nil {
				log.Fatalln(err)
			}
		case OutputFormatMarkdown:
			g := generator.NewMarkdownGenerator(*projectName, *projectEntry, *sourcePrefix)
			buf, err = g.RenderDocument(
				groups.MainTrails,
				groups.SideTrails,
				groups.Remarks,
			)
			if err != nil {
				log.Fatalln(err)
			}
		}
		if len(*outputFile) > 0 {
			if err := ioutil.WriteFile(*outputFile, buf, 0600); err != nil {
				log.Fatalln(err)
			}
			return
		}
		fmt.Println(string(buf))
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

const (
	SourceTypeMarkdown = "markdown"
)

const (
	RendererTypeGithubFlavoured = "gfm"
	RendererTypeGithubReadme    = "readme"
)

func cmdRender(c *cli.Cmd) {
	formatFrom := c.StringOpt("from", "markdown", "Select format of the source to render. Supported: markdown.")
	formatTo := c.StringOpt("to", "readme", "Select format of the output. Supported: \n\t\t* gfm (GitHub Flavoured Markdown HTML) \n\t\t* readme (GitHub Readme Markdown HTML)\n\t\t*")
	outputFile := c.StringOpt("o output", "", "Output file path.")
	inputFile := c.StringArg("FILE", "", "Input file to read, must be in specified format (e.g. markdown).")
	ghClientID := c.StringOpt("client-id", "", "GitHub Client ID for Authorization of requests.")
	ghClientSecret := c.StringOpt("client-secret", "", "GitHub Client Secret for Authorization of requests.")
	c.Action = func() {
		switch *formatFrom {
		case SourceTypeMarkdown:
		default:
			log.Fatalln("unsupported input format:", *formatFrom)
		}
		switch *formatTo {
		case RendererTypeGithubFlavoured, RendererTypeGithubReadme:
		default:
			log.Fatalln("unsupported output format:", *formatFrom)
		}

		src, err := ioutil.ReadFile(*inputFile)
		if err != nil {
			log.Fatalln(err)
		}
		var buf []byte
		switch *formatTo {
		case RendererTypeGithubFlavoured:
			r := renderer.NewGithubRenderer(*projectName, *ghClientID, *ghClientSecret)
			buf, err = r.RenderGFM(src)
			if err != nil {
				log.Fatalln(err)
			}
			if len(*outputFile) == 0 {
				*outputFile = *inputFile + ".html"
			}
			if err := ioutil.WriteFile(*outputFile, buf, 0600); err != nil {
				log.Fatalln(err)
			}
		case RendererTypeGithubReadme:
			r := renderer.NewGithubRenderer(*projectName, *ghClientID, *ghClientSecret)
			buf, err = r.RenderReadme(src)
			if err != nil {
				log.Fatalln(err)
			}
			if len(*outputFile) == 0 {
				*outputFile = *inputFile + ".html"
			}
			if err := ioutil.WriteFile(*outputFile, buf, 0600); err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func isMatching(path string, rxs []*regexp.Regexp) bool {
	for _, rx := range rxs {
		if rx.MatchString(path) {
			return true
		}
	}
	return false
}

func compileRxs(rxs []string) ([]*regexp.Regexp, error) {
	var compiled []*regexp.Regexp
	for _, rx := range rxs {
		r, err := regexp.Compile(rx)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Regexp: %s error: %v", rx, err)
		}
		compiled = append(compiled, r)
	}
	return compiled, nil
}

func containsPrefix(path string, prefixes []string) bool {
	for _, p := range prefixes {
		if p == "..." {
			return true
		}
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
