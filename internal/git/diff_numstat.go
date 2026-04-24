package git

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

type DiffNumStat struct {
	Paths []PathDiffNumStat
}

type PathDiffNumStat struct {
	Path         string
	AddedLines   int
	DeletedLines int
}

func parseDiffNumStat(data []byte) (DiffNumStat, error) {
	tokens := bytes.Split(data, []byte{0})
	var tokenErrs []error
	diffNumStat := DiffNumStat{}
	for i, token := range tokens {
		if len(token) == 0 {
			continue
		}
		pathDiffNumStat, err := parseDiffNumStatPathToken(token)
		if err != nil {
			tokenErrs = append(tokenErrs, &TokenParseError{
				TokenIndex: i,
				Token:      token,
				Err:        err,
			})
		} else {
			diffNumStat.Paths = append(diffNumStat.Paths, pathDiffNumStat)
		}
	}
	var err error
	if len(tokenErrs) > 0 {
		err = &ParseError{Errs: tokenErrs}
	}
	return diffNumStat, err
}

func parseDiffNumStatPathToken(token []byte) (PathDiffNumStat, error) {
	parts := bytes.Split(token, []byte{'\t'})
	if len(parts) != 3 {
		return PathDiffNumStat{}, errors.New("bad number of columns")
	}
	addedLines, err := strconv.Atoi(string(parts[0]))
	if err != nil {
		return PathDiffNumStat{}, fmt.Errorf("bad added lines: %w", err)
	}
	deletedLines, err := strconv.Atoi(string(parts[1]))
	if err != nil {
		return PathDiffNumStat{}, fmt.Errorf("bad deleted lines: %w", err)
	}
	return PathDiffNumStat{
		Path:         string(parts[2]),
		AddedLines:   addedLines,
		DeletedLines: deletedLines,
	}, nil
}
