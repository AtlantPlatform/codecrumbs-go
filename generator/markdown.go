package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/xlab/treeprint"

	"github.com/AtlantPlatform/codecrumbs-go/parser"
)

type Markdown struct {
	ProjectName  string
	ProjectEntry string
	SourcePrefix string
}

func NewMarkdownGenerator(projectName, projectEntry, sourcePrefix string) *Markdown {
	return &Markdown{
		ProjectName:  projectName,
		ProjectEntry: projectEntry,
		SourcePrefix: sourcePrefix,
	}
}

func joinPrefixPath(prefix, path string) string {
	if strings.HasPrefix(prefix, "https://github.com") {
		prefix = strings.TrimPrefix(prefix, "https://")
		prefix = strings.TrimSuffix(prefix, "/")
		prefixParts := strings.Split(prefix, "/")
		if len(prefixParts) == 2 {
			pathParts := strings.Split(path, "/")
			repoName := pathParts[0]
			path = strings.Join(pathParts[1:], "/")
			return fmt.Sprintf("https://%s/%s/blob/master/%s", prefix, repoName, path)
		}
		return fmt.Sprintf("https://%s/blob/master/%s", prefix, path)
	}
	return prefix + path
}

func (m *Markdown) RenderDocument(
	mainTrails map[string][]*parser.CodeCrumb,
	sideTrails map[string][]*parser.CodeCrumb,
	remarks []*parser.CodeCrumb,
) ([]byte, error) {
	buf := new(bytes.Buffer)

	var (
		statsMain    int
		statsSide    int
		statsRemarks int
		statsTotal   int
	)
	for _, crumbs := range mainTrails {
		statsMain++
		for range crumbs {
			statsTotal++
		}
	}
	for _, crumbs := range sideTrails {
		statsSide++
		for range crumbs {
			statsTotal++
		}
	}
	for range remarks {
		statsRemarks++
		statsTotal++
	}

	fmt.Fprintf(buf, "# %s\n\n", strings.Title(m.ProjectName))

	fmt.Fprintf(buf, "❓ This document has been generated using [cc-go](https://github.com/AtlantPlatform/codecrumbs-go)"+
		" tool. Running for **%s** project it found **%d** codecrumbs in total. There are **%d** main trails of codecrumbs,"+
		" that are crossing the project's entrypoint, also **%d** side trails and **%d** standalone remarks.\n\n",
		m.ProjectName, statsTotal, statsMain, statsSide, statsRemarks)

	var mainNames []string
	var sideNames []string

	if len(mainTrails) > 0 {
		mainNames = make([]string, 0, len(mainTrails))
		for name := range mainTrails {
			mainNames = append(mainNames, name)
		}
		sort.Strings(mainNames)

		fmt.Fprintf(buf, "- [Main Trails](%s)\n", anchor("Main Trails"))
		for _, name := range mainNames {
			fmt.Fprintf(buf, "  - [%s](%s)\n",
				strings.Title(name), anchor(name))
		}
	}
	if len(sideTrails) > 0 {
		sideNames = make([]string, 0, len(sideTrails))
		for name := range sideTrails {
			sideNames = append(sideNames, name)
		}
		sort.Strings(sideNames)

		fmt.Fprintf(buf, "- [Side Trails](%s)\n", anchor("Side Trails"))
		for _, name := range sideNames {
			fmt.Fprintf(buf, "  - [%s](%s)\n",
				strings.Title(name), anchor(name))
		}
	}
	var anchorsSeen = map[string]int{}
	if len(remarks) > 0 {
		fmt.Fprintf(buf, "- [Remarks](%s)\n", anchor("Remarks"))
		for _, cc := range remarks {
			var title string
			if len(cc.Title) > 0 {
				title = fmt.Sprintf("L%d: %s", cc.SourceLine, strings.Title(cc.Title))
			} else {
				title = fmt.Sprintf("L%d", cc.SourceLine)
			}
			anchorTitle := anchor(title)
			seenTimes := anchorsSeen[anchorTitle]
			anchorsSeen[anchorTitle]++
			if seenTimes > 0 {
				anchorTitle = fmt.Sprintf("%s-%d", anchorTitle, seenTimes)
			}
			fmt.Fprintf(buf, "  - [%s](%s)\n", title, anchorTitle)
		}
	}
	fmt.Fprintf(buf, "\n")

	renderTrail := func(trailName string, trail []*parser.CodeCrumb) {
		fmt.Fprintf(buf, "### %s\n\n", strings.Title(trailName))

		fmt.Fprintf(buf, "~~~\n")
		fmt.Fprintf(buf, "%s", treeForTrail(trail))
		fmt.Fprintf(buf, "~~~\n\n")

		for _, cc := range trail {
			if len(cc.Title) > 0 {
				fmt.Fprintf(buf, "#### %d. %s\n\n",
					cc.TrailStep, strings.Title(cc.Title),
				)
			} else {
				fmt.Fprintf(buf, "#### %d.\n\n",
					cc.TrailStep,
				)
			}
			if len(cc.DescLines) > 0 {
				for _, line := range cc.DescLines {
					fmt.Fprintf(buf, "%s\n", line)
				}
				fmt.Fprintf(buf, "\n")
			}
			fmt.Fprintf(buf, "📖 [%s:%d](%s#L%d)\n\n",
				cc.SourcePath, cc.SourceLine, joinPrefixPath(m.SourcePrefix, cc.SourcePath), cc.SourceLine)
			if len(cc.PeekedLines) > 0 {
				lines, modified := trimTabPrefix(cc.PeekedLines)
				for modified {
					lines, modified = trimTabPrefix(lines)
				}
				lines, modified = trimSpacePrefix(lines)
				for modified {
					lines, modified = trimSpacePrefix(lines)
				}
				fmt.Fprintf(buf, "~~~%s\n", strings.ToLower(cc.LanguageName))
				for _, line := range lines {
					fmt.Fprintf(buf, "%s\n", line)
				}
				fmt.Fprintf(buf, "~~~\n\n")
			}
		}
	}

	if len(mainTrails) > 0 {
		fmt.Fprintf(buf, "## Main Trails\n\n")
		for _, name := range mainNames {
			renderTrail(name, mainTrails[name])
		}
	}
	if len(sideTrails) > 0 {
		fmt.Fprintf(buf, "## Side Trails\n\n")
		for _, name := range sideNames {
			renderTrail(name, sideTrails[name])
		}
	}

	if len(remarks) > 0 {
		fmt.Fprintf(buf, "## Remarks\n\n")
		for _, cc := range remarks {
			if len(cc.Title) > 0 {
				fmt.Fprintf(buf, "### L%d: %s\n\n",
					cc.SourceLine, strings.Title(cc.Title),
				)
			} else {
				fmt.Fprintf(buf, "### L%d\n\n", cc.SourceLine)
			}
			if len(cc.DescLines) > 0 {
				for _, line := range cc.DescLines {
					fmt.Fprintf(buf, "%s\n", line)
				}
				fmt.Fprintf(buf, "\n")
			}
			fmt.Fprintf(buf, "📖 [%s:%d](%s#L%d)\n\n",
				cc.SourcePath, cc.SourceLine, joinPrefixPath(m.SourcePrefix, cc.SourcePath), cc.SourceLine)
			if len(cc.PeekedLines) > 0 {
				lines, modified := trimTabPrefix(cc.PeekedLines)
				for modified {
					lines, modified = trimTabPrefix(lines)
				}
				lines, modified = trimSpacePrefix(lines)
				for modified {
					lines, modified = trimSpacePrefix(lines)
				}
				fmt.Fprintf(buf, "~~~%s\n", strings.ToLower(cc.LanguageName))
				for _, line := range lines {
					fmt.Fprintf(buf, "%s\n", line)
				}
				fmt.Fprintf(buf, "~~~\n\n")
			}
		}
	}

	return buf.Bytes(), nil
}

func trimTabPrefix(lines []string) ([]string, bool) {
	for _, line := range lines {
		if !strings.HasPrefix(line, "\t") {
			return lines, false
		}
	}
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], "\t")
	}
	return lines, true
}

func trimSpacePrefix(lines []string) ([]string, bool) {
	for _, line := range lines {
		if !strings.HasPrefix(line, " ") {
			return lines, false
		}
	}
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], " ")
	}
	return lines, true
}

var wordRx = regexp.MustCompile(`\w+`)

func anchor(title string) string {
	m := wordRx.FindAllStringSubmatch(title, -1)
	words := make([]string, 0, len(m))
	for _, w := range m {
		words = append(words, w[0])
	}
	return "#" + strings.ToLower(strings.Join(words, "-"))
}

func treeForTrail(trail []*parser.CodeCrumb) string {
	tree := treeprint.New()

	branches := make(map[string]treeprint.Tree, len(trail))
	var currentFile string
	for _, cc := range trail {
		if len(currentFile) == 0 {
			branch := tree.AddBranch(filepath.Base(cc.SourcePath))
			branches[cc.SourcePath] = branch
			currentFile = cc.SourcePath
			branch.AddMetaNode(fmt.Sprintf("#%d", cc.TrailStep), strings.Title(cc.Title))
			continue
		} else if cc.SourcePath != currentFile {
			branch, ok := branches[cc.SourcePath]
			if !ok {
				branch = branches[currentFile].AddBranch(filepath.Base(cc.SourcePath))
				branches[cc.SourcePath] = branch
			}
			branch.AddMetaNode(fmt.Sprintf("#%d", cc.TrailStep), strings.Title(cc.Title))
			currentFile = cc.SourcePath
			continue
		}
		branches[currentFile].AddMetaNode(fmt.Sprintf("#%d", cc.TrailStep), strings.Title(cc.Title))
	}

	return tree.String()[2:]
}
