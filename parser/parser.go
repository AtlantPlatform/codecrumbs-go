package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
	"strconv"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/xlab/catcher"
)

var (
	separatorParts = []byte(";")
	separatorTrail = []byte("#")

	prefixCC = regexp.MustCompile(`\s?(cc:|CC:)\s?`)
)

func CollectCrumbs(sourcePath string, commentLineLang *LanguageDefinition, r io.Reader) ([]*CodeCrumb, error) {
	// inComment keeps mark if we are in a comments block.
	inComment := false
	// inCommentCC keeps mark if we are in CC'd section of a comment. The CC'd section must
	// be the last section of the commentary block, all the rest will be captured as CC's descripton.
	inCommentCC := false

	var list []*CodeCrumb
	var current *CodeCrumb

	var line int
	s := bufio.NewScanner(r)
	for s.Scan() {
		line++
		lineBytes := s.Bytes()

		if cleanLine, ok := commentLineLang.Match(lineBytes); ok {
			inComment = true
			idx := prefixCC.FindIndex(cleanLine)
			if idx != nil {
				// found a CC comment
				if inCommentCC {
					err := errors.New("cannot place a CC marker in the same comment block multiple times")
					return nil, err
				}
				inCommentCC = true
				current = &CodeCrumb{
					ID:           uuid.Must(uuid.NewRandom()).String(),
					LanguageName: commentLineLang.Name,
					SourcePath:   sourcePath,
					PeekedLines:  []string{},
				}
				current.ParseCC(cleanLine[idx[1]:])
				current.SourceLine = line
			} else if inCommentCC {
				// there is not marker on this line, but we already seen it
				current.DescLines = append(current.DescLines, string(cleanLine))
			}
		} else if inComment {
			// the comment section ended
			inComment = false
			if inCommentCC {
				// had a CC part, so can sumbit it to the list
				inCommentCC = false
				list = append(list, current)
				current = nil
			}
		}
		for _, prevCC := range list {
			if prevCC.PeekNum > 0 {
				prevCC.PeekNum--
				// add a line to the crumb that demanded it
				prevCC.PeekedLines = append(prevCC.PeekedLines, string(lineBytes))
			}
		}
	}
	if current != nil {
		list = append(list, current)
		current = nil
	}
	if err := s.Err(); err != nil {
		log.Errorln("error during scan", err)
	}
	return list, nil
}

type CodeCrumb struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	TrailID      string   `json:"trail_id,omitempty"`
	TrailStep    int      `json:"trail_step,omitempty"`
	DescLines    []string `json:"desc_lines"`
	SourcePath   string   `json:"source_path"`
	SourceLine   int      `json:"source_line"`
	PeekNum      int      `json:"-"`
	PeekedLines  []string `json:"peeked_lines"`
	LanguageName string   `json:"lang_name"`
}

func (cc *CodeCrumb) ParseCC(line []byte) {
	defer catcher.Catch()

	ccParts := bytes.Split(line, separatorParts)
	currentIdx := 0
	if bytes.Contains(ccParts[currentIdx], separatorTrail) {
		trailParts := bytes.Split(ccParts[currentIdx], separatorTrail)
		cc.TrailID = string(bytes.TrimSpace(trailParts[0]))
		cc.TrailStep, _ = strconv.Atoi(string(trailParts[1]))
		currentIdx++
	}
	cc.Title = string(bytes.TrimSpace(ccParts[currentIdx]))
	currentIdx++

	peekNum, _ := strconv.Atoi(string(bytes.TrimSpace(ccParts[currentIdx])))
	if peekNum > 0 {
		cc.PeekNum = peekNum
		currentIdx++
	}

	cc.DescLines = append(cc.DescLines, string(bytes.TrimSpace(ccParts[currentIdx])))
}
