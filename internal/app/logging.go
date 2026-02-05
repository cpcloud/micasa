package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/dlclark/regexp2"
	"github.com/dlclark/regexp2/syntax"
)

type logLevel int

const (
	logOff logLevel = iota
	logError
	logInfo
	logDebug
)

func (l logLevel) String() string {
	switch l {
	case logOff:
		return "OFF"
	case logError:
		return "ERROR"
	case logDebug:
		return "DEBUG"
	default:
		return "INFO"
	}
}

type logEntry struct {
	Time    time.Time
	Level   logLevel
	Message string
}

type logState struct {
	enabled      bool
	focus        bool
	displayLevel logLevel
	maxEntries   int
	input        textinput.Model
	filter       *regexp2.Regexp
	filterErr    error
	entries      []logEntry
	highlights   []logMatch
}

type logMatch struct {
	Start int
	End   int
}

func newLogState() logState {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "type a Perl-compatible regex"
	input.CharLimit = 256
	input.Width = 32
	return logState{
		displayLevel: logInfo,
		maxEntries:   500,
		input:        input,
	}
}

func (l *logState) cycleLevel() {
	switch l.displayLevel {
	case logOff:
		l.displayLevel = logError
	case logError:
		l.displayLevel = logInfo
	case logInfo:
		l.displayLevel = logDebug
	case logDebug:
		l.displayLevel = logOff
	}
}

func (l *logState) levelLabel() string {
	return l.displayLevel.String()
}

func (l *logState) setFilter(pattern string) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		l.filter = nil
		l.filterErr = nil
		return
	}
	re, err := regexp2.Compile(pattern, 0)
	if err != nil {
		l.filterErr = err
		l.filter = nil
		return
	}
	l.filter = re
	l.filterErr = nil
}

func (l *logState) append(level logLevel, message string) {
	if level > l.displayLevel {
		return
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	entry := logEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
	}
	l.entries = append(l.entries, entry)
	if len(l.entries) > l.maxEntries {
		l.entries = l.entries[len(l.entries)-l.maxEntries:]
	}
}

func (l *logState) matchLine(line string) bool {
	if l.filterErr != nil || l.filter == nil {
		l.highlights = nil
		return true
	}
	match, err := l.filter.FindStringMatch(line)
	if err != nil {
		l.highlights = nil
		return false
	}
	if match == nil {
		l.highlights = nil
		return false
	}
	groups := match.Groups()
	if len(groups) == 0 {
		l.highlights = []logMatch{{Start: match.Index, End: match.Index + match.Length}}
		return true
	}
	matches := make([]logMatch, 0, len(groups))
	for _, group := range groups {
		if group.Length <= 0 {
			continue
		}
		start := group.Index
		end := group.Index + group.Length
		if start < 0 || end <= start {
			continue
		}
		matches = append(matches, logMatch{Start: start, End: end})
	}
	if len(matches) == 0 {
		matches = []logMatch{{Start: match.Index, End: match.Index + match.Length}}
	}
	l.highlights = matches
	return true
}

// findHighlights returns all non-overlapping match spans for the active filter.
func (l *logState) findHighlights(line string) []logMatch {
	if l.filterErr != nil || l.filter == nil {
		return nil
	}
	match, err := l.filter.FindStringMatch(line)
	if err != nil || match == nil {
		return nil
	}
	var result []logMatch
	for match != nil {
		if match.Length > 0 {
			result = append(result, logMatch{Start: match.Index, End: match.Index + match.Length})
		}
		match, err = l.filter.FindNextMatch(match)
		if err != nil {
			break
		}
	}
	return result
}

func (l *logState) validityLabel() string {
	if l.filterErr != nil {
		if parseErr, ok := l.filterErr.(*syntax.Error); ok {
			return fmt.Sprintf("invalid: %s", parseErr.Code.String())
		}
		message := l.filterErr.Error()
		message = strings.TrimPrefix(message, "error parsing regexp: ")
		message = strings.TrimPrefix(message, "error parsing regex: ")
		return fmt.Sprintf("invalid: %s", message)
	}
	if strings.TrimSpace(l.input.Value()) == "" {
		return "no filter"
	}
	return "valid"
}
