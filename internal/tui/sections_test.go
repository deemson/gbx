package tui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

func TestRelativeScanPathBecomesAbsoluteWithoutChangingLabel(t *testing.T) {
	m := newModelWithDirectories([]Directory{{Label: "./", Path: "."}})
	require.True(t, filepath.IsAbs(m.sections[0].path))
	require.Equal(t, "./", m.sections[0].label)
}

func sectionModel() model {
	m := newModelWithDirectories([]Directory{
		{Label: "./services", Path: "services"},
		{Label: "../TOOLS", Path: "tools"},
	})
	m.sections[0].complete = true
	m.sections[1].complete = true
	return m
}

func TestSectionsKeepArgumentOrderAndHaveNoGutter(t *testing.T) {
	m := sectionModel().addRepoTo(0, "api", git.Repo{}).addRepoTo(1, "cli", git.Repo{})
	lines := strings.Split(ansi.Strip(m.listContent()), "\n")

	require.Equal(t, "./services", strings.TrimRight(lines[0], " "))
	require.Equal(t, "../TOOLS", strings.TrimRight(lines[2], " "))
	require.True(t, strings.HasPrefix(lines[1], "  api")) // repository gutter remains
	require.False(t, strings.HasPrefix(lines[0], "  "))   // heading starts at row start
	require.NotContains(t, m.listContent(), cursorBandSeq+colorDim.Render("./services"))
}

func TestDuplicateBasenamesUpdateBySection(t *testing.T) {
	m := sectionModel().addRepoTo(0, "shared", git.Repo{}).addRepoTo(1, "shared", git.Repo{})
	m = m.setStatusRef(repoRef{section: 1, name: "shared"}, repoStatus{branch: "second"})

	require.Nil(t, m.repos[0].status)
	require.Equal(t, "second", m.repos[1].status.branch)
}

func TestFilterRanksWithinFixedSections(t *testing.T) {
	m := sectionModel().addRepoTo(0, "z-api", git.Repo{}).addRepoTo(0, "alpha", git.Repo{})
	m = m.addRepoTo(1, "api", git.Repo{}).addRepoTo(1, "beta", git.Repo{})
	m.filter = "api"

	matched := m.matched()
	require.Equal(t, []int{0, 1}, []int{matched[0].section, matched[1].section})
	require.Equal(t, []string{"z-api", "api"}, []string{matched[0].name, matched[1].name})
	lines := strings.Split(ansi.Strip(m.listContent()), "\n")
	require.Equal(t, "./services", strings.TrimRight(lines[0], " "))
	require.Equal(t, "../TOOLS", strings.TrimRight(lines[2], " "))
}

func TestEmptySectionStates(t *testing.T) {
	m := sectionModel()
	m.sections[0].complete = false
	m.filter = "api"
	lines := strings.Split(ansi.Strip(m.listContent()), "\n")

	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	require.Equal(t, []string{"./services", "loading...", "../TOOLS", "no repos"}, lines)

	m = m.addRepoTo(1, "cli", git.Repo{})
	lines = strings.Split(ansi.Strip(m.listContent()), "\n")
	require.Equal(t, "no matches", strings.TrimRight(lines[3], " "))
}

func TestCursorSkipsHeadingsAcrossSections(t *testing.T) {
	m := sectionModel().addRepoTo(0, "api", git.Repo{}).addRepoTo(1, "cli", git.Repo{})
	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyDown})

	require.Equal(t, 1, m.cursorIndex())
	require.Equal(t, "cli", m.matched()[m.cursorIndex()].name)
	lines := strings.Split(m.listContent(), "\n")
	require.NotContains(t, lines[2], cursorBandSeq) // second heading
	require.Contains(t, lines[3], cursorBandSeq)    // second repository
}

func TestBulkCommandTargetsEveryOccurrence(t *testing.T) {
	m := sectionModel().addRepoTo(0, "shared", git.Repo{}).addRepoTo(1, "shared", git.Repo{})
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})
	m = updated.(model)

	require.Equal(t, cmdRunning, m.repos[0].cmd)
	require.Equal(t, cmdRunning, m.repos[1].cmd)
}

func TestLongHeadingTruncatesExactSourceOnlyAtRender(t *testing.T) {
	m := newModelWithDirectories([]Directory{{Label: "./a-very-long-directory", Path: "x"}})
	m.sections[0].complete = true
	m = drive(t, m, tea.WindowSizeMsg{Width: 10, Height: 20})

	require.Equal(t, "./a-very-long-directory", m.sections[0].label)
	require.Equal(t, "./a-very-…", strings.Split(ansi.Strip(m.listContent()), "\n")[0])
}

func TestAllEmptySectionsCanScrollWithoutSelectingHeadings(t *testing.T) {
	dirs := make([]Directory, 5)
	for i := range dirs {
		dirs[i] = Directory{Label: "dir", Path: "."}
	}
	m := newModelWithDirectories(dirs)
	for i := range m.sections {
		m.sections[i].complete = true
	}
	m = drive(t, m, tea.WindowSizeMsg{Width: 40, Height: 9}) // four list lines
	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyDown})

	require.Equal(t, -1, m.cursorIndex())
	require.Equal(t, 1, m.top)
}
