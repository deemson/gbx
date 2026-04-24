package git

import (
	"bytes"
	"errors"
)

type Status struct {
	Commit string
	Branch string
	Paths  []any
}

type RegularPathStatus struct {
	Path string
}

type MovedPathStatus struct {
	Path     string
	OrigPath string
}

type UnmergedPathStatus struct {
	Path string
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
				tokenErrs = append(tokenErrs, &StatusTokenParseError{
					TokenIndex: i,
					Token:      token,
					Err:        err,
				})
			}
		case '1':
			pathStatus, err := parseStatusRegularPathToken(token)
			if err != nil {
				tokenErrs = append(tokenErrs, &StatusTokenParseError{
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
				tokenErrs = append(tokenErrs, &StatusTokenParseError{
					TokenIndex: i,
					Token:      token,
					Err:        err,
				})
			} else {
				status.Paths = append(status.Paths, pathStatus)
			}
		case 'u':
			pathStatus, err := parseStatusUnmergedPathToken(token)
			if err != nil {
				tokenErrs = append(tokenErrs, &StatusTokenParseError{
					TokenIndex: i,
					Token:      token,
					Err:        err,
				})
			} else {
				status.Paths = append(status.Paths, pathStatus)
			}
		default:
			tokenErrs = append(tokenErrs, &StatusTokenParseError{
				TokenIndex: i,
				Token:      token,
				Err:        errors.New("unknown token type"),
			})
		}
	}
	var err error
	if len(tokenErrs) > 0 {
		err = &StatusParseError{Errs: tokenErrs}
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

func parseStatusRegularPathToken(token []byte) (RegularPathStatus, error) {
	parts := bytes.Split(token, []byte{' '})
	// parts[0] is expected to be '1'
	xy := parts[1]
	_ = xy
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
	path := parts[8]
	return RegularPathStatus{
		Path: string(path),
	}, nil
}

func parseStatusMovedPathToken(token []byte, origPathToken []byte) (MovedPathStatus, error) {
	parts := bytes.Split(token, []byte{' '})
	// parts[0] is expected to be '2'
	xy := parts[1]
	_ = xy
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
	path := parts[9]
	return MovedPathStatus{
		Path:     string(path),
		OrigPath: string(origPathToken),
	}, nil
}

func parseStatusUnmergedPathToken(token []byte) (UnmergedPathStatus, error) {
	parts := bytes.Split(token, []byte{' '})
	// parts[0] is expected to be 'u'
	xy := parts[1]
	_ = xy
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
	path := parts[10]
	_ = path
	return UnmergedPathStatus{
		Path: string(path),
	}, nil
}
