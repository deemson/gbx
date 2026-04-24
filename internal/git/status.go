package git

import (
	"bytes"
	"errors"
	"fmt"
)

type PathState byte

const (
	UnknownPathState PathState = iota
	NotChangedPathState
	ModifiedPathState
	AddedPathState
	DeletedPathState
	RenamedPathState
	CopiedPathState
	UpdatedPathState
	UntrackedPathState
	IgnoredPathState
)

type Status struct {
	Commit string
	Branch string
	Paths  []any
}

type UntrackedPathStatus struct {
	Path string
}

type RegularPathStatus struct {
	StateIndex PathState
	StateFS    PathState
	Path       string
}

type MovedPathStatus struct {
	StateIndex PathState
	StateFS    PathState
	Path       string
	OrigPath   string
}

type ConflictPathStatus struct {
	StateThem PathState
	StateUs   PathState
	Path      string
}

// parseStatus parses according to https://www.kernel.org/pub/software/scm/git/docs/git-status.html
func parseStatus(data []byte) (Status, error) {
	tokens := bytes.Split(data, []byte{0})
	var tokenErrs []error
	status := Status{}
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if len(token) == 0 {
			continue
		}
		switch token[0] {
		case '#':
			err := parseStatusMetadata(token, &status)
			if err != nil {
				tokenErrs = append(tokenErrs, &TokenParseError{
					TokenIndex: i,
					Token:      token,
					Err:        err,
				})
			}
		case '?':
			pathStatus, err := parseStatusUntrackedPathToken(token)
			if err != nil {
				tokenErrs = append(tokenErrs, &TokenParseError{
					TokenIndex: i,
					Token:      token,
					Err:        err,
				})
			} else {
				status.Paths = append(status.Paths, pathStatus)
			}
		case '1':
			pathStatus, err := parseStatusRegularPathToken(token)
			if err != nil {
				tokenErrs = append(tokenErrs, &TokenParseError{
					TokenIndex: i,
					Token:      token,
					Err:        err,
				})
			} else {
				status.Paths = append(status.Paths, pathStatus)
			}
		case '2':
			var origPathToken []byte
			if i+1 < len(tokens) {
				i++ // consume the next token as the original path
				origPathToken = tokens[i]
			}
			pathStatus, err := parseStatusMovedPathToken(token, origPathToken)
			if err != nil {
				tokenErrs = append(tokenErrs, &TokenParseError{
					TokenIndex: i,
					Token:      token,
					Err:        err,
				})
			} else {
				status.Paths = append(status.Paths, pathStatus)
			}
		case 'u':
			pathStatus, err := parseStatusConflictPathToken(token)
			if err != nil {
				tokenErrs = append(tokenErrs, &TokenParseError{
					TokenIndex: i,
					Token:      token,
					Err:        err,
				})
			} else {
				status.Paths = append(status.Paths, pathStatus)
			}
		default:
			tokenErrs = append(tokenErrs, &TokenParseError{
				TokenIndex: i,
				Token:      token,
				Err:        errors.New("unknown token type"),
			})
		}
	}
	var err error
	if len(tokenErrs) > 0 {
		err = &ParseError{Errs: tokenErrs}
	}
	return status, err
}

func parseStatusMetadata(token []byte, status *Status) error {
	parts := bytes.Split(token, []byte{' '})
	// parts[0] is expected to be '#'
	switch string(parts[1]) {
	case "branch.oid":
		status.Commit = string(parts[2])
	case "branch.head":
		status.Branch = string(parts[2])
	default:
		return errors.New("unknown metadata")
	}
	return nil
}

func parseStatusUntrackedPathToken(token []byte) (UntrackedPathStatus, error) {
	// parts[0:2] is expected to be '? '
	return UntrackedPathStatus{
		Path: string(token[2:]),
	}, nil
}

func parseStatusRegularPathToken(token []byte) (RegularPathStatus, error) {
	parts := bytes.Split(token, []byte{' '})
	// parts[0] is expected to be '1'
	xy := parts[1]
	stateIndex, stateFS, err := parseStatusPathXY(xy)
	if err != nil {
		return RegularPathStatus{}, err
	}
	sub := parts[2]
	_ = sub
	octalModeHead := parts[3]
	_ = octalModeHead
	octalModeIndex := parts[4]
	_ = octalModeIndex
	octalModeWorktree := parts[5]
	_ = octalModeWorktree
	hashHead := parts[6]
	_ = hashHead
	hashIndex := parts[7]
	_ = hashIndex
	path := bytes.Join(parts[8:], []byte{' '})
	return RegularPathStatus{
		Path:       string(path),
		StateFS:    stateFS,
		StateIndex: stateIndex,
	}, nil
}

func parseStatusMovedPathToken(token []byte, origPathToken []byte) (MovedPathStatus, error) {
	parts := bytes.Split(token, []byte{' '})
	// parts[0] is expected to be '2'
	xy := parts[1]
	stateIndex, stateFS, err := parseStatusPathXY(xy)
	if err != nil {
		return MovedPathStatus{}, err
	}
	sub := parts[2]
	_ = sub
	octalModeHead := parts[3]
	_ = octalModeHead
	octalModeIndex := parts[4]
	_ = octalModeIndex
	octalModeWorktree := parts[5]
	_ = octalModeWorktree
	hashHead := parts[6]
	_ = hashHead
	hashIndex := parts[7]
	_ = hashIndex
	similarityScore := parts[8]
	_ = similarityScore
	path := bytes.Join(parts[9:], []byte{' '})
	return MovedPathStatus{
		Path:       string(path),
		OrigPath:   string(origPathToken),
		StateFS:    stateFS,
		StateIndex: stateIndex,
	}, nil
}

func parseStatusConflictPathToken(token []byte) (ConflictPathStatus, error) {
	parts := bytes.Split(token, []byte{' '})
	// parts[0] is expected to be 'u'
	xy := parts[1]
	stateUs, stateThem, err := parseStatusPathXY(xy)
	if err != nil {
		return ConflictPathStatus{}, err
	}
	sub := parts[2]
	_ = sub
	octalModeStage1 := parts[3]
	_ = octalModeStage1
	octalModeStage2 := parts[4]
	_ = octalModeStage2
	octalModeStage3 := parts[5]
	_ = octalModeStage3
	octalModeWorktree := parts[6]
	_ = octalModeWorktree
	hashStage1 := parts[7]
	_ = hashStage1
	hashStage2 := parts[8]
	_ = hashStage2
	hashStage3 := parts[9]
	_ = hashStage3
	path := bytes.Join(parts[10:], []byte{' '})
	return ConflictPathStatus{
		Path:      string(path),
		StateUs:   stateUs,
		StateThem: stateThem,
	}, nil
}

func parseStatusPathState(b byte) PathState {
	switch b {
	case '.':
		return NotChangedPathState
	case 'M':
		return ModifiedPathState
	case 'A':
		return AddedPathState
	case 'D':
		return DeletedPathState
	case 'R':
		return RenamedPathState
	case 'C':
		return CopiedPathState
	case 'U':
		return UpdatedPathState
	case '?':
		return UntrackedPathState
	case '!':
		return IgnoredPathState
	default:
		return UnknownPathState
	}
}

func parseStatusPathXY(xy []byte) (PathState, PathState, error) {
	if len(xy) != 2 {
		return UnknownPathState, UnknownPathState, fmt.Errorf("XY len is %d expected 2", len(xy))
	}
	stateX := parseStatusPathState(xy[0])
	if stateX == UnknownPathState {
		return UnknownPathState, UnknownPathState, errors.New("unknown X in XY")
	}
	stateY := parseStatusPathState(xy[1])
	if stateY == UnknownPathState {
		return UnknownPathState, UnknownPathState, errors.New("unknown Y in XY")
	}
	return stateX, stateY, nil
}
