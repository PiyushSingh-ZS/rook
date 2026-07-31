package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
)

// maxPatchBytes caps the unified diff we ship to the review UI so a huge
// generated-file diff can't blow up the response.
const maxPatchBytes = 1_500_000

// diffFile is one changed path in a worktree's review diff.
type diffFile struct {
	Path   string `json:"path"`
	Status string `json:"status"` // M, A, D, R, or "?" for untracked
	Add    int    `json:"add"`
	Del    int    `json:"del"`
}

// diffResult is the agent's total contribution in a worktree: everything that
// differs from the fork point (committed + uncommitted), plus untracked files.
type diffResult struct {
	Base      string     `json:"base"`  // the ref/commit the diff is against
	Files     []diffFile `json:"files"`
	Patch     string     `json:"patch"`
	Add       int        `json:"add"`
	Del       int        `json:"del"`
	Truncated bool       `json:"truncated"`
}

// gitOut runs a git subcommand in dir and returns trimmed stdout+stderr. For
// commands that signal "differences found" via exit code 1 (e.g. diff
// --no-index), the caller should ignore the error and read the output.
func gitOut(dir string, args ...string) (string, error) {
	full := append([]string{"-C", dir}, args...)
	out, err := execWithTimeout("git", 15*time.Second, full...)
	return strings.TrimSpace(string(out)), err
}

// isWorkTree reports whether dir is inside a git work tree.
func isWorkTree(dir string) bool {
	out, err := gitOut(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// defaultBranchRef finds the base branch to diff a worktree against — the repo's
// default branch (origin/HEAD), falling back to common names. Returns "" if none
// resolve, in which case the caller diffs against HEAD (uncommitted only).
func defaultBranchRef(dir string) string {
	if out, err := gitOut(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "origin/HEAD"); err == nil && out != "" && !strings.Contains(out, "fatal") {
		return out
	}
	for _, c := range []string{"origin/main", "origin/master", "main", "master"} {
		if _, err := gitOut(dir, "rev-parse", "--verify", "--quiet", c); err == nil {
			return c
		}
	}
	return ""
}

// diffBase returns the commit to diff against: the merge-base between HEAD and
// the default branch (the fork point), so the review shows what THIS worktree
// added and not unrelated commits the base gained since. Falls back to HEAD.
func diffBase(dir string) string {
	def := defaultBranchRef(dir)
	if def != "" {
		if mb, err := gitOut(dir, "merge-base", "HEAD", def); err == nil && mb != "" {
			return mb
		}
		return def
	}
	return "HEAD"
}

// parseNumstat turns `git diff --numstat` output into per-file add/del counts,
// keyed by path. Binary files ("-\t-\tpath") are recorded with -1/-1.
func parseNumstat(out string) map[string][2]int {
	m := map[string][2]int{}
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		add, del := -1, -1
		if parts[0] != "-" {
			add, _ = strconv.Atoi(parts[0])
		}
		if parts[1] != "-" {
			del, _ = strconv.Atoi(parts[1])
		}
		m[parts[2]] = [2]int{add, del}
	}
	return m
}

// computeDiff builds the review diff for a worktree/repo directory.
func computeDiff(dir string) (diffResult, error) {
	var r diffResult
	if !isWorkTree(dir) {
		return r, errf(http.StatusBadRequest, "not a git work tree")
	}
	base := diffBase(dir)
	r.Base = base

	// tracked changes vs the fork point (committed + uncommitted)
	nameStatus, _ := gitOut(dir, "diff", "--name-status", "-M", base)
	nums := parseNumstat(mustOut(gitOut(dir, "diff", "--numstat", base)))
	for _, line := range strings.Split(nameStatus, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		st := fields[0]
		p := fields[len(fields)-1] // for renames (Rxx\told\tnew) take the new path
		f := diffFile{Path: p, Status: string(st[0])}
		if n, ok := nums[p]; ok {
			f.Add, f.Del = n[0], n[1]
		}
		r.Files = append(r.Files, f)
		if f.Add > 0 {
			r.Add += f.Add
		}
		if f.Del > 0 {
			r.Del += f.Del
		}
	}

	var patch strings.Builder
	if p := mustOut(gitOut(dir, "diff", "-M", base)); p != "" {
		patch.WriteString(p)
		patch.WriteString("\n")
	}

	// untracked files: show each as an addition (read-only; no index mutation)
	untracked := mustOut(gitOut(dir, "ls-files", "--others", "--exclude-standard"))
	count := 0
	for _, up := range strings.Split(untracked, "\n") {
		if up == "" || count >= 200 {
			continue
		}
		count++
		// --no-index exits 1 when files differ; ignore the error, read output
		np, _ := gitOut(dir, "diff", "--no-index", "--numstat", "--", os.DevNull, up)
		add := 0
		for path, ad := range parseNumstat(np) {
			_ = path
			if ad[0] > 0 {
				add = ad[0]
			}
		}
		r.Files = append(r.Files, diffFile{Path: up, Status: "?", Add: add})
		r.Add += add
		if patch.Len() < maxPatchBytes {
			dp, _ := gitOut(dir, "diff", "--no-index", "--", os.DevNull, up)
			patch.WriteString(dp)
			patch.WriteString("\n")
		}
	}

	s := patch.String()
	if len(s) > maxPatchBytes {
		s = s[:maxPatchBytes]
		r.Truncated = true
	}
	r.Patch = s
	return r, nil
}

// mustOut drops the error from a gitOut call (used where a non-zero exit just
// means "no output / differences found" and an empty string is the right value).
func mustOut(out string, _ error) string { return out }

// handleDiff backs the review surface: GET /api/diff?path=<worktree-or-repo>.
func handleDiff(ctx *gofr.Context) (any, error) {
	p := ctx.Param("path")
	if p == "" {
		return nil, errf(http.StatusBadRequest, "path required")
	}
	p = filepath.Clean(p)
	fi, err := os.Stat(p)
	if err != nil || !fi.IsDir() {
		return nil, errf(http.StatusBadRequest, "path is not a directory")
	}
	res, err := computeDiff(p)
	if err != nil {
		return nil, err
	}
	return rawJSON(res)
}
