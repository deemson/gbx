package tui

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	log.Logger = zerolog.New(io.Discard)
	zerolog.DefaultContextLogger = &log.Logger
	os.Exit(m.Run())
}

type safeBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuf) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuf) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

type testProgram struct {
	t     *testing.T
	p     *tea.Program
	out   *safeBuf
	in    *io.PipeWriter
	errCh chan error
}

func runTestProgram(t *testing.T, dir string) *testProgram {
	t.Helper()
	out := &safeBuf{}
	inR, inW := io.Pipe()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	p := tea.NewProgram(newModel(dir),
		tea.WithContext(ctx),
		tea.WithInput(inR),
		tea.WithOutput(out),
		tea.WithoutSignalHandler(),
		tea.WithoutCatchPanics(),
		tea.WithWindowSize(120, 30),
	)
	errCh := make(chan error, 1)
	go func() {
		_, err := p.Run()
		errCh <- err
	}()
	t.Cleanup(func() {
		p.Quit()
		select {
		case <-errCh:
		case <-time.After(2 * time.Second):
			t.Log("test program did not quit within 2s")
		}
		_ = inW.Close()
	})
	return &testProgram{t: t, p: p, out: out, in: inW, errCh: errCh}
}

// send injects each rune of s as a printable key press, simulating typing into
// the always-focused filter. Uses Program.Send (thread-safe) rather than the
// raw input pipe, which avoids depending on terminal input parsing.
func (tp *testProgram) send(s string) {
	tp.t.Helper()
	for _, r := range s {
		tp.p.Send(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func (tp *testProgram) waitForContent(substrs ...string) {
	tp.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s := tp.out.String()
		all := true
		for _, sub := range substrs {
			if !strings.Contains(s, sub) {
				all = false
				break
			}
		}
		if all {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	require.Failf(tp.t, "waitForContent timeout",
		"expected all of %q in output.\nOutput:\n%s", substrs, tp.out.String())
}
