package main

import (
	"sort"
	"strings"

	"github.com/AtlantPlatform/codecrumbs-go/parser"
)

type GroupedCodeCrumbs struct {
	MainTrails map[string][]*parser.CodeCrumb `json:"main_trails"`
	SideTrails map[string][]*parser.CodeCrumb `json:"side_trails"`
	Remarks    []*parser.CodeCrumb            `json:"remarks"`
}

func regroupCodeCrumbs(entryPoint string, crumbsList [][]*parser.CodeCrumb) *GroupedCodeCrumbs {
	grouped := &GroupedCodeCrumbs{
		MainTrails: make(map[string][]*parser.CodeCrumb),
		SideTrails: make(map[string][]*parser.CodeCrumb),
		Remarks:    make([]*parser.CodeCrumb, 0, 100),
	}
	for _, crumbs := range crumbsList {
		for _, cc := range crumbs {
			if len(cc.TrailID) == 0 {
				grouped.Remarks = append(grouped.Remarks, cc)
				continue
			}
			if _, ok := grouped.MainTrails[cc.TrailID]; ok {
				grouped.MainTrails[cc.TrailID] = append(
					grouped.MainTrails[cc.TrailID], cc,
				)
				continue
			} else if strings.HasSuffix(cc.SourcePath, entryPoint) {
				if prevTrail, ok := grouped.SideTrails[cc.TrailID]; ok {
					grouped.MainTrails[cc.TrailID] = prevTrail
					delete(grouped.SideTrails, cc.TrailID)
				}
				grouped.MainTrails[cc.TrailID] = append(
					grouped.MainTrails[cc.TrailID], cc,
				)
				continue
			}
			grouped.SideTrails[cc.TrailID] = append(
				grouped.SideTrails[cc.TrailID], cc,
			)
		}
	}
	for _, trail := range grouped.MainTrails {
		sort.Sort(CodeCrumbsByTrail(trail))
	}
	for _, trail := range grouped.SideTrails {
		sort.Sort(CodeCrumbsByTrail(trail))
	}
	sort.Sort(CodeCrumbsByFile(grouped.Remarks))
	return grouped
}

type CodeCrumbsByFile []*parser.CodeCrumb

func (s CodeCrumbsByFile) Len() int      { return len(s) }
func (s CodeCrumbsByFile) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s CodeCrumbsByFile) Less(i, j int) bool {
	return s[i].SourcePath < s[j].SourcePath && s[i].SourceLine < s[j].SourceLine
}

type CodeCrumbsByTrail []*parser.CodeCrumb

func (s CodeCrumbsByTrail) Len() int      { return len(s) }
func (s CodeCrumbsByTrail) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s CodeCrumbsByTrail) Less(i, j int) bool {
	return s[i].TrailStep < s[j].TrailStep
}
