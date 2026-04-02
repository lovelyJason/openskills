package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/huh"
)

type SelectOption struct {
	Label       string
	Value       string
	Disabled    bool
	DisabledMsg string
}

func MultiSelect(title string, options []string) ([]string, error) {
	items := make([]SelectOption, len(options))
	for i, o := range options {
		items[i] = SelectOption{Label: o, Value: o}
	}
	return MultiSelectEx(title, items)
}

func MultiSelectEx(title string, options []SelectOption) ([]string, error) {
	for _, o := range options {
		if o.Disabled {
			hint := o.DisabledMsg
			if hint == "" {
				hint = "not available"
			}
			Dim("  %s — %s", o.Label, hint)
		}
	}

	var enabled []SelectOption
	for _, o := range options {
		if !o.Disabled {
			enabled = append(enabled, o)
		}
	}
	if len(enabled) == 0 {
		return nil, fmt.Errorf("no selectable options")
	}

	if !isInteractive() {
		var vals []string
		for _, o := range enabled {
			vals = append(vals, o.Value)
		}
		return vals, nil
	}

	opts := make([]huh.Option[string], len(enabled))
	for i, o := range enabled {
		opts[i] = huh.NewOption(o.Label, o.Value)
	}

	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(title).
				Options(opts...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}
	return selected, nil
}

func Select(title string, options []string) (string, error) {
	if !isInteractive() && len(options) > 0 {
		return options[0], nil
	}

	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, o)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(opts...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}
	return selected, nil
}

func Confirm(question string) (bool, error) {
	if !isInteractive() {
		return true, nil
	}

	var confirmed bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(question).
				Value(&confirmed),
		),
	)

	if err := form.Run(); err != nil {
		return false, err
	}
	return confirmed, nil
}

func SelectInstallMode() (string, error) {
	if !isInteractive() {
		return "symlink", nil
	}

	opts := []huh.Option[string]{
		huh.NewOption("symlink  \033[2m— link to source repo, auto-updates on git pull\033[0m", "symlink"),
		huh.NewOption("native   \033[2m— copy files to editor directory, manual update\033[0m", "native"),
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Install mode for this marketplace:").
				Options(opts...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}
	return selected, nil
}

func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func Success(msg string, args ...interface{}) {
	fmt.Printf("  \033[32m✓\033[0m "+msg+"\n", args...)
}

func Info(msg string, args ...interface{}) {
	fmt.Printf("  \033[36mℹ\033[0m "+msg+"\n", args...)
}

func Warn(msg string, args ...interface{}) {
	fmt.Printf("  \033[33m⚠\033[0m "+msg+"\n", args...)
}

func Error(msg string, args ...interface{}) {
	fmt.Printf("  \033[31m✗\033[0m "+msg+"\n", args...)
}

func Dim(msg string, args ...interface{}) {
	fmt.Printf("  \033[2m"+msg+"\033[0m\n", args...)
}

func Header(msg string) {
	fmt.Println()
	fmt.Printf("\033[1m%s\033[0m\n", msg)
	fmt.Println(strings.Repeat("─", len(msg)+4))
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	msg    string
	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.Mutex
}

func NewSpinner(msg string) *Spinner {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Spinner{msg: msg, cancel: cancel, done: make(chan struct{})}
	go s.run(ctx)
	return s
}

func (s *Spinner) run(ctx context.Context) {
	defer close(s.done)
	i := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(os.Stderr, "\r\033[K")
			return
		case <-ticker.C:
			s.mu.Lock()
			msg := s.msg
			s.mu.Unlock()
			fmt.Fprintf(os.Stderr, "\r\033[K  \033[36m%s\033[0m %s", spinnerFrames[i%len(spinnerFrames)], msg)
			i++
		}
	}
}

func (s *Spinner) Update(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
}

func (s *Spinner) Stop() {
	s.cancel()
	<-s.done
}
