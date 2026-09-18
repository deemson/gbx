package tui

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/deemson/gbx/internal/git"
)

const (
	errorPaneMaxHeight = 12
	errorPanelTabWidth = 4
)

type errorPanelSpec struct {
	key          string
	title        string
	body         string
	bodyHeight   int
	paneHeight   int
	overflow     bool
	showFooter   bool
	contentWidth int
}

// selectedError returns the effective error for the cursored row. It deliberately
// follows repoEntry.summary's command-before-load priority.
func (m model) selectedError() (repoEntry, error, bool) {
	ci := m.cursorIndex()
	if ci < 0 {
		return repoEntry{}, nil, false
	}
	r := m.matched()[ci]
	err := r.currentError()
	return r, err, err != nil
}

// syncErrorPanel materializes the derived selected error into the viewport. The
// pane itself is still rendered only in list mode. Keeping this synchronization
// at the Update boundary makes cursor/error transitions testable as model state.
func (m model) syncErrorPanel() model {
	spec, ok := m.makeErrorPanelSpec()
	if !ok {
		m.errorViewKey = ""
		m.errorView.SetContent("")
		m.errorView.GotoTop()
		return m
	}

	m.errorView.SetWidth(spec.contentWidth)
	m.errorView.SetHeight(spec.bodyHeight)
	m.errorView.SoftWrap = false // body has already been word-wrapped safely
	m.errorView.SetContent(spec.body)
	if spec.key != m.errorViewKey {
		m.errorView.GotoTop()
	}
	m.errorViewKey = spec.key
	return m
}

func (m model) makeErrorPanelSpec() (errorPanelSpec, bool) {
	r, err, ok := m.selectedError()
	if !ok || m.width <= 0 || m.baseListHeight() < 3 {
		return errorPanelSpec{}, false
	}

	repoName := strings.ReplaceAll(sanitizeDiagnostic(r.name), "\n", " ")
	title := "Error — " + repoName
	details := errorDetails(err)
	body := ansi.Wrap(details, m.width, " /_")
	bodyLines := lineCount(body)

	// The pane takes at most half of the content area (and never more than 12
	// rows), leaving at least one list row visible above it.
	baseHeight := m.baseListHeight()
	maxPaneHeight := min(errorPaneMaxHeight, max(baseHeight/2, 2), baseHeight-1)
	overflow := 1+bodyLines > maxPaneHeight // one row for the titled divider
	showFooter := overflow && maxPaneHeight >= 3
	bodyHeight := bodyLines
	if overflow {
		bodyHeight = maxPaneHeight - 1
		if showFooter {
			bodyHeight--
		}
	}
	paneHeight := 1 + bodyHeight
	if showFooter {
		paneHeight++
	}

	return errorPanelSpec{
		key:          selectedErrorKey(r.ref(), err, details),
		title:        title,
		body:         body,
		bodyHeight:   bodyHeight,
		paneHeight:   paneHeight,
		overflow:     overflow,
		showFooter:   showFooter,
		contentWidth: m.width,
	}, true
}

func selectedErrorKey(ref repoRef, err error, details string) string {
	identity := ""
	v := reflect.ValueOf(err)
	if v.IsValid() && v.Kind() == reflect.Pointer && !v.IsNil() {
		identity = fmt.Sprintf("%x", v.Pointer())
	}
	return fmt.Sprintf("%d\x00%s\x00%T\x00%s\x00%s", ref.section, ref.name, err, identity, details)
}

func (m model) scrollErrorPanel(msg tea.Msg) model {
	m = m.syncErrorPanel()
	if m.errorViewKey == "" {
		return m
	}
	m.errorView, _ = m.errorView.Update(msg)
	return m
}

// errorPane renders the details as a bottom split between the repository list
// and its footer. It consumes normal layout height rather than covering rows.
func (m model) errorPane() string {
	if m.mode != modeList {
		return ""
	}
	m = m.syncErrorPanel()
	spec, ok := m.makeErrorPanelSpec()
	if !ok {
		return ""
	}

	rows := []string{ruleWithMarker(m.width, colorRed.Bold(true).Render(spec.title)), m.errorView.View()}
	if spec.showFooter {
		rows = append(rows, errorScrollFooter(m.errorView, spec.contentWidth))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func errorScrollFooter(v viewport.Model, width int) string {
	hint := "PgUp/PgDn/wheel scroll"
	pct := fmt.Sprintf("%3.f%%", v.ScrollPercent()*100)
	gap := width - lipgloss.Width(hint) - lipgloss.Width(pct)
	line := hint
	if gap >= 1 {
		line += strings.Repeat(" ", gap) + pct
	} else {
		line = ansi.Truncate(hint, width, "…")
	}
	return colorDim.Render(line)
}

func errorDetails(err error) string {
	var runErr *git.RunError
	if !errors.As(err, &runErr) {
		return sanitizeDiagnostic(err.Error())
	}

	type field struct {
		label string
		value string
	}
	fields := []field{
		{"Summary", runErr.Summary()},
		{"Command", quoteGitCommand(runErr.Res.Args)},
		{"Exit status", strconv.Itoa(runErr.Res.ExitCode)},
		{"Stderr", string(runErr.Res.Stderr)},
		{"Stdout", string(runErr.Res.Stdout)},
	}
	parts := make([]string, 0, len(fields))
	labelStyle := colorYellow.Bold(true)
	for _, f := range fields {
		value := strings.TrimRight(sanitizeDiagnostic(f.value), "\n")
		if strings.TrimSpace(value) == "" {
			continue
		}
		parts = append(parts, labelStyle.Render(f.label)+"\n"+value)
	}
	return strings.Join(parts, "\n\n")
}

func quoteGitCommand(args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, "git")
	for _, arg := range args {
		parts = append(parts, strconv.Quote(arg))
	}
	return strings.Join(parts, " ")
}

// sanitizeDiagnostic turns subprocess bytes into inert terminal text while
// preserving their readable content. Tabs advance to four-column stops.
func sanitizeDiagnostic(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = ansi.Strip(s)

	var b strings.Builder
	col := 0
	for _, r := range s {
		switch {
		case r == '\n':
			b.WriteRune(r)
			col = 0
		case r == '\t':
			spaces := errorPanelTabWidth - col%errorPanelTabWidth
			b.WriteString(strings.Repeat(" ", spaces))
			col += spaces
		case unicode.IsControl(r):
			continue
		default:
			b.WriteRune(r)
			col += ansi.StringWidth(string(r))
		}
	}
	return b.String()
}

func lineCount(s string) int {
	if s == "" {
		return 1
	}
	return strings.Count(s, "\n") + 1
}
