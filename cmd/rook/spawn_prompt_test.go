package main

import "testing"

func TestPaneReady(t *testing.T) {
	// the splash alone is NOT ready
	splash := "Welcome back Piyush!\n Tips for getting started\n Run /init ..."
	if paneReady(splash) {
		t.Fatalf("splash should not be considered ready")
	}
	// once the REPL status line / input is drawn, it's ready
	for _, ready := range []string{
		"⏸ manual mode on · ? for shortcuts · 1 agent",
		"❯ ",
		"esc to interrupt",
	} {
		if !paneReady(ready) {
			t.Fatalf("expected ready for %q", ready)
		}
	}
}

func TestPaneHasPrompt(t *testing.T) {
	prompt := "Review GitHub pull request #134 in zopdev/notification: read the diff…"
	// input still empty (placeholder) → not entered
	if paneHasPrompt("❯ Try \"fix lint errors\"", prompt) {
		t.Fatalf("empty input should not match the prompt")
	}
	// prompt text visible in the input → entered
	if !paneHasPrompt("❯ Review GitHub pull request #134 in zopdev/notif", prompt) {
		t.Fatalf("prompt text in pane should match")
	}
	// empty prompt is trivially 'entered'
	if !paneHasPrompt("anything", "") {
		t.Fatalf("empty prompt should be treated as entered")
	}
}
