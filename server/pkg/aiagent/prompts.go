package aiagent

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Prompt library references.
//
// A step or node can carry its instruction inline, or point at a named prompt
// in the library: "@prompt:ticket_summary" (the active version) or
// "@prompt:ticket_summary@3" (pinned to version 3). Pinning is for a
// production flow that must not change when someone edits the prompt;
// following the active version is for one that should pick up improvements.
//
// The library itself lives in the automation component (it is stored in the
// database); this package only knows how to recognise a reference and who to
// ask, so the workflow steps and the rule chain nodes - which cannot import
// each other's services - resolve prompts the same way.

type (
	// ResolvedPrompt is a library prompt chosen by a reference.
	ResolvedPrompt struct {
		Text string
		// Ref names exactly what was used, "handle@version", for traces
		Ref string
	}

	// PromptResolver finds a prompt; version 0 means the active one.
	PromptResolver func(ctx context.Context, handle string, version int) (*ResolvedPrompt, error)
)

var (
	promptResolverMu sync.RWMutex
	promptResolver   PromptResolver

	// PromptHandlePattern is what a library handle looks like.
	PromptHandlePattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,63}$`)

	promptRef = regexp.MustCompile(`^@prompt:([a-z][a-z0-9_.-]{0,63})(?:@([0-9]+))?$`)
)

func SetPromptResolver(fn PromptResolver) {
	promptResolverMu.Lock()
	promptResolver = fn
	promptResolverMu.Unlock()
}

// IsPromptRef reports whether text is a library reference.
func IsPromptRef(text string) bool {
	return promptRef.MatchString(strings.TrimSpace(text))
}

// ResolvePrompt returns the text to send for a step or node's prompt: the
// library prompt a reference points at, or text itself when it is not a
// reference. ref is "handle@version" for a library prompt, "" otherwise.
//
// A reference to a prompt that does not exist is an error - quietly sending
// "@prompt:foo" to a model as if it were an instruction would be far worse.
func ResolvePrompt(ctx context.Context, text string) (resolved, ref string, err error) {
	m := promptRef.FindStringSubmatch(strings.TrimSpace(text))
	if m == nil {
		return text, "", nil
	}

	promptResolverMu.RLock()
	fn := promptResolver
	promptResolverMu.RUnlock()
	if fn == nil {
		return "", "", fmt.Errorf("prompt library is not available (%s)", m[0])
	}

	version := 0
	if m[2] != "" {
		version, _ = strconv.Atoi(m[2])
	}

	p, err := fn(ctx, m[1], version)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", m[0], err)
	}
	return p.Text, p.Ref, nil
}
