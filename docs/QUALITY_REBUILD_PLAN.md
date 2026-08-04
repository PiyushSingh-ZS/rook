# rook Quality Rebuild Plan

> The definitive execution doc. rook is a local, single-user Go/GoFr + vanilla-JS
> dashboard for driving AI coding agents. This plan takes all 26 features from
> "just there" to best-in-the-world — by making the *core* of each feature honest
> and active, not by piling on toggles to be listed.
>
> **Authority rule:** where the adversarial verifier and the original audit
> disagree, this plan follows the **verifier** — its adjusted grades, its
> tightened move set, and its scope-creep cuts. Every move the verifier flagged as
> scope creep is explicitly excluded here (see each feature's "Cut" line).

---

## The five highest-leverage bets

These are where the effort goes first. Each is cheap-to-moderate and moves a
flagship or removes a systemic lie.

1. **Make the Overview narrate and lead with the verdict** *(A · Watching, wow 5)* —
   rewrite the "Now" line to narrate from the tool *sequence* with dwell, promote
   the watchdog Health that already exists from 1-of-15 tiles to the lead block,
   and collapse the three duplicated tabs into deep-links. The headline watching
   screen stops being a tile-dump. Mostly S/M, no new subsystems.

2. **Finish the Diff review — the only true "driving" surface** *(B · Driving, grade 7, wow 5)* —
   it is already the best feature and the one place a human decision becomes agent
   action. Consolidate onto the persistent SQLite-backed path, add intraline
   highlighting + keyboard nav, auto-resolve comments on re-diff, and structure the
   *existing* AI reviewer's output into the tree. Reuse-first, no new endpoints.

3. **Fold chains into graphs and make the DAG actually parallel** *(D · Orchestration)* —
   delete `chains.go` (a strict, buggier subset), give nodes a real pass/fail
   signal so conditional edges mean something, and run independent branches
   concurrently in their own worktrees. This *removes* a primitive while making the
   survivor worth more than the thing it replaced.

4. **Stop the gate, watchdog, and quality score from lying** *(A/F, cheap, high-trust)* —
   default the destructive gate ON and fix its fail-open transport; wire the
   watchdog Health into the notifier it claims to feed; render never-verified
   sessions as "unrated" instead of a flattering 100. All small, all pure honesty.

5. **Make notifications escalate and the driving gate confirm** *(A/B/G, cheap)* —
   give `notifyStuck` a real `Priority: urgent` header (today it's byte-identical
   to a routine ping), re-alert while still stuck, and make the allow/deny gate
   re-capture the pane to confirm the action landed instead of returning
   `{ok:true}` optimistically.

---

## 1. Diagnosis — why the features feel "just there"

Reading all 26 audits and their adversarial verifications, seven systemic
patterns recur. They are the reason good code still reads as box-checked.

**P1 — Passive display where the tool is positioned to act.** The recurring
verifier note is "it displays, it never decides." The roster is a grouped list,
not a triage queue. The Overview is 15 tiles you assemble a conclusion from. The
quality score renders a number and nothing consumes it. Insights is a wall of
tables. Audit is a log. The watchdog paints a chip. rook *launches the agents* —
it holds the position to intervene — and mostly chooses to narrate nothing and
gate nothing.

**P2 — Signals computed, then discarded.** Over and over, the data already exists
and the feature throws it away. `statusCheckRollup` is fetched by the GitHub view
and never rendered (a red PR looks identical to a green one). The transcript
carries `is_error` per tool call, but the trace timeline colors nothing. The
watchdog Health is computed and buried as one of fifteen Overview tiles. The
Overview already renders `ago(updatedAt)`, which for a waiting agent *is* the
blocked-duration — but it's unlabeled. The fix is usually "surface what you
already have," not "go build a new signal."

**P3 — Fabricated or fake affordances presented as real.** The trace bars encode
inter-arrival gaps (think + idle time) dressed as tool latency. The allow/deny
gate has hardcoded "opt 2 / opt 3" buttons and `<kbd>a</kbd>/<kbd>d</kbd>` hints
that do the wrong thing. The board ships ~35 lines of dead drag scaffolding and a
header calling itself "draggable." The quality score defaults to "excellent" for a
task that did nothing. These aren't gaps — they're dishonesty the reader trips on.

**P4 — Duplicate implementations of one feature that drift.** Two GitHub browsers
(`app.js` classic vs `operator-github.js`). Two send-to-agent paths for review
comments with different message formats, one non-persisting. Two summary prompt
builders that have *already* diverged in URL-escaping. `columnFor` copied into
`app.js` as `boardColumn`. The destructive gate's 5 regexes vs the audit view's 9.
Chains as a strict subset of graphs. The honest work is often *consolidation*, not
addition.

**P5 — Silent failure and swallowed errors.** The worktree delete discards every
git/os error and returns `{ok:true}` — a green "Removed" toast over a no-op. The
notifier swallows non-2xx from ntfy/Slack, so a revoked webhook means zero alerts
and zero feedback. The reflexion loop dies on a tmux name collision with all four
return values discarded. Auto-review swallows passes 2–3. The destructive gate is
fail-open: if rook is down, `curl -m 5 … ; exit 0` lets the command run. A safety
feature off precisely when it isn't running.

**P6 — The audit's own prescription is the disease.** The most important
meta-pattern: nearly every audit, having diagnosed correctly, then prescribed a
"10-star" that piles on new surfaces — an LLM-judge eval platform inside a badge,
a merge queue in the GitHub tab, a policy engine for 5 regexes, a ticket-inbox
aggregator, drag-to-dispatch that contradicts a documented decision, a
remote-control-from-Slack notifier. The verifier's verdict is "overreach" on 24 of
26 features. **The failure mode the owner is fighting is reproduced by the audits
themselves.** This plan takes the verifier's shorter, sharper move sets.

**P7 — The good parts are consistently undersold, the missing parts overbuilt.**
The live terminal (`term_ws.go`), the diff backend's PR-base resolution, the DAG
scheduler's skip-propagation, the ⌘K palette's existing j/k nav, the prompt-
delivery state machine — these are genuine craft that the audits skate past on
their way to proposing parallel subsystems. Excellence here is mostly *finishing
and connecting what exists*, not starting new things.

**The throughline:** make each feature (a) tell the truth, (b) surface the signal
it already computes, (c) act where rook is uniquely positioned to, and (d) do it
by connecting existing parts rather than growing new ones.

---

## 2. Grade table

Grades are the **verifier's adjustedGrade**. Sorted worst-first within each
cluster. Effort is the aggregate of the *kept* moves only.

### A · Watching

| Feature | Grade | Disposition | Wow | Effort |
|---|---|---|---|---|
| Trace timeline | 3 | improve | 4 | M–L |
| Quality score | 3 | improve | 4 | S–M |
| Watchdog / health | 3 | improve | 4 | M |
| Overview — what is it doing | 4 | improve | 5 | S–M |
| Operator roster | 5 | improve | 4 | S–M |

### B · Driving

| Feature | Grade | Disposition | Wow | Effort |
|---|---|---|---|---|
| Review comments → agent | 3 | improve | 4 | S–M |
| Drive: allow/deny/live terminal | 5 | improve | 4 | S–M |
| Diff review | 7 | improve | 5 | M |

### C · Launching

| Feature | Grade | Disposition | Wow | Effort |
|---|---|---|---|---|
| Follow repo instructions | 3 | improve | 3 | S–M |
| Launch agent | 4 | improve | 4 | S–M |
| Start from a ticket | 4 | improve | 4 | S–M |
| Resume a closed session | 5 | improve | 4 | S |
| GitHub view | 5 | improve | 4 | S–M |

### D · Orchestration

| Feature | Grade | Disposition | Wow | Effort |
|---|---|---|---|---|
| Board (kanban) | 3 | improve | 4 | S |
| Task chains | 3 | **fold → graphs** | 3 | M |
| Task graphs (DAG) | 4 | improve | 4 | L |
| Workspace — worktrees & hooks | 4 | improve (split) | 4 | S–M |

### E · Intelligence

| Feature | Grade | Disposition | Wow | Effort |
|---|---|---|---|---|
| Insights — cost & tokens | 4 | improve | 4 | M |
| Cheap-model routing | 4 | improve | 4 | S |

### F · Automation

| Feature | Grade | Disposition | Wow | Effort |
|---|---|---|---|---|
| Auto-verify + Reflexion retry | 3 | improve | 4 | S–M |
| Destructive-command gate | 3 | improve | 3 | S–M |
| Auto-review | 3 | improve | 4 | S–M |
| Audit — command log | 4 | improve | 4 | M |

### G · Everyday

| Feature | Grade | Disposition | Wow | Effort |
|---|---|---|---|---|
| Keyboard & command palette | 4 | improve | 4 | S–M |
| Notifications | 5 | improve | 4 | S |
| Daily summaries | 5 | improve | 4 | S–M |

---

## 3. Cut & fold list

Be decisive. Fewer, better features.

### Fold

- **Task chains → Task graphs.** A chain is a linear graph. `chains.go` is a
  strict, *buggier* subset: it marks a running step `done` unconditionally (a
  crashed agent reports "success"), has no persistence, and false-advances on any
  Stop in the CWD. Fold it into the DAG (implicit `node[i] depends on node[i-1]`),
  delete `chains.go` / `chains_test.go` / the chain modal / the "New task chain"
  palette entry, add a "linear" quick-template to the graph modal, and migrate
  existing chain rows on boot. **This removes a primitive** — the anti-creep win.

### Split (don't merge)

- **Workspace = worktrees + hooks** is a catch-all screen bolting two unrelated
  things together. Don't build the mega-board the audit wanted; instead keep them
  as two honest halves: a git-authoritative worktree list with one real action
  (Review) and a hooks/events feed with click-through. If neither half is made
  active, the passive worktree list should fold into the diff/review surface,
  which already computes each worktree's contribution.

### Cut (features/capabilities proposed by audits that should NOT be built)

These are the scope-creep items the verifier rejected across the set. Building any
of them reproduces the owner's failure mode:

- **LLM-judge / eval platform inside the Quality score** — a whole subsystem
  (spawned Claude calls, diff-hash caching, config flag) for a badge.
- **Continuous background verify on every scan** — hammers the machine; most
  worktrees have no detected verify command anyway.
- **Watchdog auto-unstick / send-Esc / nudge-injection + per-provider threshold
  config** — destructive capability + knobs; a watchdog's job is to surface truth.
- **Risk-classification engine on the allow/deny gate** — a brittle regex
  classifier whose own note admits "misclassification erodes trust"; the operator
  already sees the exact command.
- **Provider-aware keymap + remember-decision auto-approval policy store** —
  premature abstraction (one provider exists) and removing the human from a
  human-in-the-loop gate.
- **Second AI-review endpoint / localStorage persistence for Diff review** — both
  duplicate existing infrastructure (`/api/review`, the SQLite review store).
- **Auto-resolution re-diff engine + pane-scrape reply capture for Review
  comments** — the audit's self-labeled "money move" is the fragile, admittedly-
  unreliable one; a re-send/nudge button delivers most of it cheaply.
- **"Attempts: 1/2/3" fan-out in the launch modal** — duplicates chains/graphs.
- **Rule extractor + rules-preview UI + verification tail** on Follow-repo-
  instructions — heuristic slop and a compliance monitor that already exists in
  review mode.
- **Ticket Inbox aggregator + two-way status transitions + idempotency plumbing**
  — a different, bigger product than "paste an id, get a good brief."
- **Resume: cost-card join, resume menu (worktree/model/fork), codex resume,
  preview endpoint with git diffs, dead-cwd repo picker** — decorating a list.
- **GitHub: inline PR-review client, merge-when-green polling queue, four-bucket
  board + filter chips** — Linear/Graphite/Mergify envy on a launch surface.
- **Board: real HTML5 drag-to-dispatch + `onStartStep` backend + WIP limits** —
  contradicts a documented design decision; WIP limits are unenforceable on a
  board you don't dispatch into.
- **Graph: replay/fork + checkpoints table, NL-goal-to-DAG, judge node type,
  retry-with-repair engine, layout crossing-minimization** — LangGraph/Temporal
  feature-parity chasing.
- **Insights: budgets/alerts subsystem, click-to-expand drill-down** — building an
  alerting product on cost numbers that currently show $0 for a whole provider.
- **Cheap-routing: `classifyTask` auto-router + `autoRoute` toggle, realized-
  savings counterfactual, spawn-modal recommendation + `/api/route-preview`,
  mid-flight up/downshift nudges** — a learned-router-plus-dashboard product; the
  CLI can't even hot-swap a model mid-session.
- **Destructive gate: `mvdan.cc/sh` shell-AST dependency, allowlist UI subsystem,
  `/api/danger-rules` endpoint, blanket sudo/kubectl/docker blocking** — a policy
  platform on top of a transport that fails open.
- **Auto-review: synthesizer, materialize-into-comment-threads, verdict pipeline
  gate + reflexion re-injection** — three new subsystems on a reviewer whose
  output isn't even parsed yet.
- **Audit: graded severity, ack/review queue, narration banner, CSV export,
  PreToolUse guardrail** — a "Risk Console" on a single-user command log.
- **Notifications: inbound action endpoints + Slack app interactivity + Block Kit
  + cross-device "dedup" + danger-scoring ranker** — a remote-control subsystem
  requiring public exposure; device fan-out is the point, not duplication.
- **Daily summaries: deltas/sparkline/streak/impact-ranking, cadence toggles +
  Slack/email delivery** — vanity metrics on a feature that isn't yet reliable.
- **Palette: second-level 6-verb menu + operator-aware ranker + frecency + chord
  overlay** — a palette state machine and a product vision; fast matching + one
  real action is the core.

---

## 4. Sequenced phases

Ordered by leverage (wow ÷ effort) and dependency. Phase 1 is the fastest wow and
the biggest honesty payoff; later phases carry the heavier structural work.

### Phase 1 — Stop lying (cheap honesty, highest trust ROI)

Almost all S-effort. These remove fabrications and wire up signals that already
exist. No new subsystems.

| Feature | Essential moves (kept only) | Files | Effort |
|---|---|---|---|
| Operator roster | Flip "needs" sort to `updatedAt` ASC (most-stuck on top); inline one-lined `s.asking` on needs rows; relabel `ago(updatedAt)` as "blocked Xm" for waiting rows; inline Allow/Deny gated on `s.controllable && s.alive`, `stopPropagation` | `operator.js` (renderRoster ~334–349, respond 1145) | S |
| Quality score | Stop defaulting null→100; render never-verified/zero-activity as **unrated** (gray) | `operator.js` (582,650,656,1029), `quality.go` (emit `rated` bool) | S |
| Watchdog / health | Wire `notify.go` to consume Health as the one escalation source and delete the parallel `waitingSince/stuckAfter` derivation — reconcile the two scan paths (annotate in the notifier or share the annotated slice) | `notify.go`, `watchdog.go` | M |
| Drive gate | Post-action verification: re-capture the pane, confirm the prompt cleared, return real status not `{ok:true}`; delete fake `opt 2/3` labels and fake `a/d` kbd hints | `spawn.go`, `main.go` (handleRespond), `operator.js` | S |
| Destructive gate | Fix the **fail-open transport** first (curl-down = command runs); narrate the block via existing `banner()/pushNtfy/pushChat`; one case-insensitive Go ruleset (fixes 5-vs-9 drift + lowercase `drop table` miss); default the gate **ON** | `hooks.go`, `config.go`, `operator-audit.js` | S–M |
| Notifications | `notifyStuck` sends `Priority: urgent` (today identical to routine ping); re-alert on an interval while still stuck; fix `notifyFinished` missing the waiting→idle completion; log non-2xx instead of swallowing; drop `FOREMAN_NTFY` + hardcoded tag; add `Click` deep-link to the local UI | `notify.go`, `integrations.go` | S |
| Board (kanban) | Delete dead drag scaffolding + fix the "draggable" lies; delete `app.js` `boardColumn` (one source for `columnFor`); rank "Needs you" by urgency (alert first, then longest-waiting); make Review honest or relabel the column "Worktrees" | `board.js`, `app.js` | S |
| Palette | `scrollIntoView({block:'nearest'})` on active item; fuzzy subsequence scorer + `<mark>` highlight (no frecency); generate the `?` sheet from `GO_KEYS`; inline answer/allow-deny reusing `respond()` | `operator.js` (renderPaletteList 1358–1385) | S–M |
| Launch agent | Gate claude-only controls (reject/ hide model + resume when `agent!=claude`); server-side safety guards (agent binary on PATH, dirty-tree → default isolate, specific name-collision error); name the worktree branch (`worktree add -b rook/<name>`), fix the ordering leak (check uniqueness before creating the worktree) | `spawn.go`, `operator.js` | S–M |
| Workspace | Fix the lying delete (`removeWorktree` discards errors, handler always returns `{ok:true}`) + dirty guard before `--force`; fix the prefix-without-separator path guard; git-authoritative list + named branch at source | `worktree.go`, `operator-workspace.js` | S–M |

### Phase 2 — The watching intelligence (Overview + Trace + Audit + Quality persistence)

The headline "what is it doing / do I step in" work. Depends on Phase 1's roster
and Health wiring being in place.

| Feature | Essential moves | Files | Effort |
|---|---|---|---|
| Overview | Rewrite `deriveActivity` to narrate from the last-N tool **sequence** with dwell ("Editing scan.go — 3 edits, no test run in 14m") — **drop** the explore/edit/verify/debug/stuck phase taxonomy; promote the existing watchdog Health (extend with an `Action` field) to the **lead block** with its action button inline; collapse Recent-activity + Changed-files + tool-chips into one-line deep-links; render Quality once (inline card) | `scan.go` (deriveActivity), `watchdog.go` (Health.Action), `operator.js` (renderOverview 543–643) | S–M |
| Trace timeline | Carry real `DurMs` + `IsError` per `ToolCall` by matching `tool_use`→`tool_result` on a **newly-added** `tool_use_id` (this is real parsing, not "just wiring"); render real duration, hatch estimated bars for unmatched (still-running) calls; color failed spans **red** + delete the dead `costUsd` tooltip branch; group by assistant turn to fix parallel-call collapse; demote Overview Recent-activity to a last-3 deep-link | `scan.go` (contentBlock + result struct + ToolCall), `operator.js` (renderTraceTab 707–714), `charts.js` (455–507) | M–L |
| Audit — command log | Widen ingestion to Edit/Write/MultiEdit (not just Bash/Shell) — **no** severity/reviewed columns; give `ToolCall` a stable per-call id and dedup on `session|id` (fixes the 80-char-prefix collision and the silent drop of distinct commands); make the click honest (relabel "select agent," or size a real jump correctly — the Trace tab is a 40-cap chart, not a browsable list) | `audit.go`, `scan.go`, `operator.js`, `operator-audit.js` | M |
| Quality score | Dedup: delete factor-3's independent `leadingRepeat` recompute, source a single "stability" factor from the already-attached `s.Health` (factor 5 already reads it); persist only the **latest** verify pass/fail per worktree, keyed to a **diff/commit hash** so a stale green pass is invalidated (drop the history/trend table) | `quality.go`, `db.go`, `watchdog.go` | S–M |

### Phase 3 — The flagship: Diff review + Review comments

The only true "driving" surface. Reuse-first; the AI reviewer and persistent
review store already exist.

| Feature | Essential moves | Files | Effort |
|---|---|---|---|
| Diff review | Consolidate onto the persistent SQLite-backed thread path served at `/` — retire the `/classic` fire-and-forget `onSend`→`/api/send`; **intraline (word-level)** highlighting; keyboard nav `j/k/n/p` + mark-viewed-and-advance; **auto-resolution** (on next diff, re-read the changed region and flip `open`→`addressed` — the schema's unused `addressed` state); structure the **existing** `review.go` reviewer output into inline `file:line` suggestions seeding the existing comment store (do **not** add a new endpoint); cache `computeDiff` per `(dir, HEAD)` with short TTL (today two sequential 20s `gh` calls per open, no cache) | `diffview.js`, `operator.js` (live thread UI 718–800), `diff.go` (195–215), `review.go`, `review_comments.go` | M |
| Review comments → agent | Unify the two send paths onto the persisted endpoint (delete `app.js`'s raw `/api/send` string, re-add clipboard fallback there, resolve the dead `open` staging); give the agent real context — capture the ± hunk at `openEditor` time and include `side` + snippet (drop group-by-file); add a **re-send/nudge** button (send-all filters `state=='open'`, so a `sent` comment can never re-route today) | `app.js` (896–906), `operator.js` (786–798), `review_comments.go` (88, 201), `diffview.js` | S–M |

### Phase 4 — Orchestration fold

The structural bet: one primitive, made truly parallel and outcome-aware.

| Feature | Essential moves | Files | Effort |
|---|---|---|---|
| Task chains → graphs | Reimplement `handleChainCreate` as a thin adapter building a `taskGraph` with implicit linear `DependsOn`; delete the bespoke chain engine + tests + modal + palette entry; migrate existing rows on boot; minimal **one-line** hand-off packet (files-touched by N-1 + verify verdict — not `summaries.go`, not full diffs) | `chains.go` (delete), `graph.go`, `main.go`, `review.go` (229), `operator.js` | M |
| Task graphs (DAG) | **N-way parallel** dispatch: replace the single `agentRunning` flag with per-branch concurrent launch, each independent branch in its own worktree, join = wait for all upstreams (manual merge gate — **not** auto-conflict-resolution); **real per-node pass/fail** via a `SUCCESS/FAIL` marker contract in the node prompt, parsed in `completeGraphNode` (drop the judge node type + retry engine); per-node timeout + heartbeat watchdog (fold in the completion-soundness fixes); node **drill-in only** (click → transcript/diff/status). Fix the **cross-graph/graph+chain false-completion** bug (`advanceGraphsForCWD` completes every graph whose runDir==cwd, keyed on directory not session), the sub-4s debounce loss, and the silent verify-degrades-to-pass | `graph.go` (scheduleGraph, completeGraphNode, advanceGraphsForCWD, runVerifyForGraph), `graph_test.go`, `operator-graph.js` | L |
| Board (kanban) | (Review-column honesty finalized here once graph/worktree state is consolidated) | `board.js` | S |

### Phase 5 — Automation loops made honest

| Feature | Essential moves | Files | Effort |
|---|---|---|---|
| Auto-verify + Reflexion | Fix the **name-collision loop death** (reflect session name is `reflect-<base>-<HEAD-short>`; no commit → same HEAD → duplicate tmux name → loop silently dies after attempt 1); fix the **one-line prompt flattening** (`strings.Fields` collapses the whole markdown buffer to one line); process-group kill on timeout (`Setpgid`, kill the group — today only `sh` is killed, orphaning children); keep `.rook-reflect` out of the reviewed diff; one `VerifyCmd` override for both auto and manual paths; no-progress exit via output-tail hash; **make the reflection real or delete the "Reflexion (Shinn et al.)" claim**; narrate each attempt through the existing `banner/pushChat` | `review.go` (runVerify, reflex branch), `reflect.go`, `spawn.go` (184), `config.go` | S–M |
| Auto-review | Parse the `VERDICT` + findings from the transcript into a typed struct (today decorative — appears only in the prompt); notify on completion reusing the verify notification path; stop swallowing passes 2–3 (aggregate errors, not `firstErr`); default to a **single consolidated pass** (decide vs. reconcile). Fix the **bogus-SHIP** bug (gate uses fork-base `diffBase` but the reviewer sees only uncommitted `git diff` → committed work reviewed as empty → confident SHIP on nothing) and the **stale/no-review** on same-HEAD re-runs | `review.go` (69, 97–99, 219, 246–250), `operator.js` | S–M |

### Phase 6 — Launching polish + everyday reliability

Lower dependency; each is independently shippable.

| Feature | Essential moves | Files | Effort |
|---|---|---|---|
| GitHub view | Port agent-awareness from the classic page (`sessionForItem` + jump-to-session badge + double-launch warn already exist there); render the CI checks pill from the already-fetched `statusCheckRollup`; stop hardcoding squash (wire the method picker through + `--delete-branch`); add the tmux-availability fallback the classic page has (Work/Review silently fail without it) | `operator-github.js` (108, 119), `github.go` (203), `integrations.go` (76) | S–M |
| Start from a ticket | Fix the broken sources (Jira api/2→api/3 + ADF flatten; Linear `issue(id:)` expects a UUID but the modal advertises `LIN-456`); stop silent 1500-char truncation + reuse GitHub fields already fetched; ticket→repo **config map** (2 of 3 sources launch into an empty cwd); **fence the untrusted ticket body** as data before it becomes an autonomous agent prompt (prompt-injection boundary) | `integrations.go`, `operator.js` | S–M |
| Resume a closed session | Capture message-count + last-turn role + snippet in the head-scan that already runs; filter stub transcripts below a small message-count threshold (kills the bare-project-name noise); surface last-activity + a "waiting on you?" flag inline in the existing palette row; guard the double-resume race + fix the UTF-8 byte-slice title cut (`title[:90]`) | `resume.go` (72), `operator.js` | S |
| Follow repo instructions | Unify the file list (launcher + GitHub hand-off share one source — the real drift bug); read content + skip files the agent auto-loads (CLAUDE.md for claude, AGENTS.md for codex) + inject **raw** size-capped content for the ones it would miss (no distillation); tests pinning that behavior | `agentdocs.go`, `operator-github.js` | S–M |
| Insights — cost & tokens | Fix cross-provider cost (Codex renders **$0** — never calls `pricePerToken`; capture the token split in `parseCodexRollout` + add OpenAI rates); close the `~/.codex` ingestion gap in `TokenWindows`/`ScanTrends` (they glob only `~/.claude`); fix the `$/M-tok` denominator (cost includes cache tokens, `TokensTotal` doesn't); add `Cost` to `DayStat` for a spend-over-time line + WoW chip; narration strip reusing the quality×cost pairing; burn-rate + projected-exhaustion on the 5h window (**drop** the plan-cap knob) | `scan.go` (271,728,814,862), `pricing.go`, `usage.go`, `operator.js` | M |
| Cheap-model routing | Route the spawns rook already owns (review lenses, reflexion retries, auto-verify pass an explicit model — cheap tier for mechanical fix loops, strong for review) via the existing `summaryModel` pattern — **not** a big `taskModelPolicy` map; fix the two stale Haiku 4.5 numbers in `pricing.go` (and the default-branch mis-bucketing of expensive families) + update the golden test | `spawn.go`, `review.go` (96, 268), `scheduler.go`, `pricing.go`, `pricing_test.go` | S |
| Daily summaries | Escape the `saveURL` query params **Go-side only** (JS already uses `encodeURIComponent`; and do **not** URL-escape the gh-command filters — those need shell quoting); consolidate the two prompt builders into one server source (the "twin" comment is accurate and already bit — JS escapes, Go doesn't); surface the spawned session + failure (the session name is already returned, the modal throws it away); restore the date range the operator view regressed (presets exist in `app.js`); catch-up backfill on startup (persist `lastRun`, not in-memory); the deterministic factual **spine** justified **only** by reliability + cost | `scheduler.go`, `summary.go`, `operator-summaries.js`, `db.go` | S–M |

---

## 5. Per-feature detail

Each entry: what's verified-today, the honest gap, the 10-star direction, and the
**tightened moves** (verifier's kept set, with files + effort). Scope-creep items
are named on the "Cut" line and are excluded from the plan.

---

### A · Watching

#### Operator roster — grade 5, wow 4

- **Files:** `operator.js` (renderRoster, groupOf, statusOf), `scan.go`
- **Verified today:** Buckets sessions into four groups, hides empties, sorts each
  by `updatedAt` DESC. Auto-selects the highest-priority agent on load (nice). But
  the backend already computes `s.Asking` (the exact pending ask) and the Overview
  renders it with inline Allow/Deny — the roster surfaces none of it. j/k nav, a
  typing-focus guard, ⌘K palette, and g-jumps already exist.
- **Honest gap:** A "needs you" row is a passive label — it never says *what* the
  agent is waiting on. And the sort is inverted: `updatedAt` froze at block-time,
  so the longest-stuck agent sinks to the bottom of "Needs you."
- **10-star:** An active triage queue — each needs row inlines the actual ask and a
  live "blocked Xm," longest-wait on top, with inline Allow/Deny so you clear a
  gate without opening the workspace.
- **Moves (all zero-backend-change):**
  1. On needs rows, append one-lined width-clamped `esc(s.asking)`; relabel the
     existing `ago(updatedAt)` as "blocked Xm" for waiting rows — it already *is*
     the wait duration. *(operator.js; M)*
  2. Flip the needs bucket sort to `updatedAt` **ASC**; keep DESC elsewhere.
     *(operator.js:334; S)*
  3. Inline Allow/Deny on needs rows, gated on `s.controllable && s.alive`
     (matching the Overview gate), `stopPropagation` so buttons don't select the
     row. *(operator.js, reuse respond() 1145; M)*
- **Cut:** keyboard triage (already exists, mis-attributed), unread-marker
  localStorage layer, folding the header cluster. `statusUpdatedAt` is **not** in
  the payload — any move depending on it is dead until a backend change; use
  `updatedAt` ASC instead.

#### Overview — what is it doing — grade 4, wow 5

- **Files:** `operator.js` (renderOverview, qualityHTML), `scan.go` (deriveActivity)
- **Verified today:** A static 15-tile snapshot. The "Now" line comes from
  `deriveActivity`, a bare switch on the **single last tool**. The Quality card is
  the one crafted element (with an honest "not a correctness judge" note) but it's
  rendered three times. A run-timeline **already exists** one tab over
  (`rookCharts.traceTimeline`).
- **Honest gap:** The headline of a watching tool is a single-tool status label,
  and the real triage signal (watchdog Health) is buried as 1 of 15 tiles. Most of
  the screen re-tiles other tabs.
- **10-star:** The single active triage-and-narration screen — narrate intent from
  the tool sequence, lead with a Health verdict + one recommended action, deep-link
  everything that merely re-lists another tab.
- **Moves:**
  1. Rewrite `deriveActivity` to narrate from the last-N tool sequence with dwell
     ("Editing scan.go — 3 edits, no test run in 14m"). **Drop** the phase
     taxonomy. *(scan.go; S–M)*
  2. Promote the existing `computeHealth` `{Level, Reason}` to the lead block with
     its action button inline — extend Health with an `Action` field; **do not**
     create `attention.go` or a new Session field. *(watchdog.go, operator.js; S)*
  3. Collapse Recent-activity + Changed-files + tool-chips into one-line deep-links
     to Trace/Files/Insights. *(operator.js 594–623; S)*
  4. Render Quality once (inline card). *(operator.js; S)*
- **Cut:** the run-timeline sparkline (already exists in Trace — self-contradicting
  with the de-dup move), the invented verdict signals "token-burn-with-no-edits"
  and "drifting off the ask" (the latter needs correctness judgment quality.go
  deliberately refuses), the phase classifier, a new `attention.go`.
- **Note:** error-streak is **not** derivable today (`ToolCall` carries only
  Name/Summary/Timestamp); narration is stateless per scan, so phase *transitions*
  can't be detected from a single snapshot.

#### Trace timeline — grade 3, wow 4

- **Files:** `charts.js` (traceTimeline), `operator.js` (renderTraceTab)
- **Verified today:** An SVG gantt with grow-in animation, tick axis, tooltip. But
  the bar widths are **inter-arrival gaps** (think + idle time), not tool
  durations — the whole visual encoding is fabricated latency. The last span's
  duration is a guessed median. Depth-indent, `costUsd` tooltip, and the legend
  are dead code paths. Parallel calls in one turn share a timestamp and collapse to
  3px slivers.
- **Honest gap:** The one thing a waterfall exists to show — how long each thing
  took — is wrong, and per-tool pass/fail (`is_error`, already parsed into
  `ToolErrors`) is never attached to a span.
- **10-star:** Real durations + failure color, turn-grouped so parallel calls stack
  as siblings and between-turn gaps become honest "thinking Ns."
- **Moves:**
  1. Carry real `DurMs` + `IsError` per `ToolCall` by matching `tool_use`→
     `tool_result` on a **newly added** `tool_use_id`. This is **new parsing** +
     cross-line correlation (~M–L), not "just wiring" — the join key does not exist
     in the code yet. *(scan.go)*
  2. Use real `durMs`; hatch estimated bars for unmatched (still-running) calls
     instead of faking a median. *(operator.js 707–714)*
  3. Color failed spans red from `is_error`; delete the dead `costUsd` branch.
     *(charts.js:497)*
  4. Group by assistant turn to fix the parallel-call collapse. *(operator.js,
     charts.js)*
  5. Demote Overview Recent-activity to a last-3 deep-link. *(operator.js 618–623)*
- **Cut:** click-to-expand + "jump to terminal" + "send fix to agent" (a second
  intervention surface), self-written narration + "critical path" (meaningless on a
  serial log), the p90 "slow-outlier" amber tier (invents cross-session per-tool
  history that doesn't exist).
- **Note:** "real" duration still includes permission-approval/human wait — label
  it, don't sell it as ground truth. The trace also freezes at first render
  (`ws.traceLoaded`) for a *live* feature, and `ToolCalls` is truncated to
  `maxTools`, so any totals are over a tail window.

#### Quality score — grade 3, wow 4

- **Files:** `quality.go`, `scan.go`
- **Verified today:** A deterministic penalty model (100 minus factors). The
  dominant factor (build/tests, 45 pts) is almost never present — `verifyStore` is
  in-memory, filled only on the AutoVerify Stop-hook path, lost on restart, empty
  for hand-started sessions. The frontend defaults null→100 "excellent."
- **Honest gap:** A score that says "excellent" for a task that did nothing is
  worse than no score. It also duplicates the watchdog's loop signal and is never
  persisted.
- **10-star:** An honest, persisted number — unrated when there's no evidence, one
  loop signal shared with the watchdog, a backbone that survives restart and
  invalidates when the tree changes.
- **Moves:**
  1. **First:** stop defaulting null→100; render never-verified/zero-activity as
     **unrated** (gray). *(operator.js, quality.go)*
  2. Delete factor-3's independent `leadingRepeat` recompute; source one
     "stability" factor from `s.Health` (factor 5 already reads it — the dedup is
     narrower than the audit claimed). *(quality.go 94–119)*
  3. Persist only the **latest** verify pass/fail per worktree, keyed to a
     diff/commit hash so a stale green pass is invalidated. **Drop** the
     history/trend table. *(db.go, quality.go)*
- **Cut:** the LLM-judge correctness lane, continuous background verify, the
  risk-ranking board reorder (duplicates the watchdog's surfacing role).
- **Note:** the stale-pass bug (`verifyStore[dir]` never invalidated) means naive
  persistence makes it *worse* — the hash key is mandatory, not optional. `is_error`
  conflation also inflates the reliability penalty with expected/recoverable errors.

#### Watchdog / health — grade 3, wow 4

- **Files:** `watchdog.go`
- **Verified today:** Three static-threshold rules (loop ≥3 byte-identical calls,
  waiting ≥10m, frozen ≥8m). `notify.go` **never reads** `s.Health` — it
  re-derives waiting escalation independently, so loop and frozen fire zero
  notifications. The loop primitive is duplicated verbatim in `quality.go`.
  Stateless, so the chip flaps.
- **Honest gap:** The headline capability ("used to escalate notifications") isn't
  wired, the marquee loop detector is a duplicate, and it compares the human
  *Summary* string, not raw args.
- **10-star:** A truthful triage chip — wired into the one notifier, one real loop
  detector, plus the one high-value unused signal (context fill), remembering when
  a problem started.
- **Moves:**
  1. Make `notify.go` consume Health as the single escalation source; delete the
     parallel `waitingSince/stuckAfter`. **Reconcile the two scans** — the notifier
     runs its own `ScanSessions(1)` that never calls `annotateHealth`. *(notify.go,
     watchdog.go)*
  2. One shared loop detector keyed on `(tool + normalized args / target)` — not
     the Summary string — catching identical runs + A-B-A-B; consolidate watchdog
     and quality.go onto it. *(watchdog.go, quality.go)*
  3. Add the context-exhaustion rule only (`ContextTokens` vs model window — on
     Session, unused). **Fix the first-match-wins early-return** so new rules aren't
     masked. *(watchdog.go, scan.go)*
  4. Add a stateful first-seen-unhealthy timestamp (kills the flap, lets severity
     escalate); prune on death. *(watchdog.go)*
- **Cut:** intervention (Esc/nudge/auto-unstick toggle), configurable thresholds +
  per-provider overrides, error-rate + cost-burn rules, adaptive baselines,
  `computeTriage` ranking + board-lead UI.
- **Note:** "frozen" (age since UpdatedAt) can't distinguish a hung agent from a
  long tool call — pane/process liveness is the diagnostic signal, not more
  thresholds.

---

### B · Driving

#### Drive: allow/deny/live terminal — grade 5, wow 4

- **Files:** `term_ws.go`, `spawn.go`, `operator.js` (respond box)
- **Verified today:** The live terminal (`term_ws.go`) is genuinely excellent — a
  full PTY bridge over WebSocket to xterm.js, clean detach. The gate is the shallow
  half: `applyKeyAction` maps allow→Enter, deny→Escape, and returns `{ok:true}`
  optimistically without re-reading the pane. The `opt 2/3` buttons and the
  `a/d` kbd hints are fake.
- **Honest gap:** A cluster of small dishonesties — fake buttons, fake hints,
  optimistic success — not a blind gate (the actual command *is* shown via
  `pendingFull`).
- **10-star:** The terminal stays the escape hatch; the gate stops lying — confirm
  the outcome, show the real controls, render the command clearly.
- **Moves:**
  1. Post-action verification: re-capture the pane, confirm the prompt cleared,
     return a real status; toast the true outcome. *(spawn.go, main.go,
     operator.js)*
  2. Delete the fake `opt 2/3` labels and the fake `a/d` kbd hints; optionally
     surface the real numbered menu (the terminal already exposes it, so this is
     convenience). *(scan.go, operator.js)*
  3. Render the command in a mono block and branch controls on the **known prompt
     type** (`pendingTool.Name` is already parsed — Bash/Edit permission vs
     AskUserQuestion). *(operator.js, scan.go)*
- **Cut:** the risk-classification engine, the provider-aware keymap (one provider
  exists), the remember-decision auto-approval store, the red/green diff.
- **Note:** add a **pre-action staleness check** — between the UI showing "waiting"
  and the click, the agent may have moved on, so Enter/Escape injects a stray
  keystroke into an active REPL.

#### Diff review — grade 7, wow 5 *(the flagship)*

- **Files:** `diff.go`, `diffview.js` (+ `review.go`, `review_comments.go`,
  `operator.js`, `main.go` routes)
- **Verified today:** The backend PR-base resolution is genuinely strong (prefers
  canonical `gh pr diff`, documented a real bug it fixed). The surface served at
  `/` **already** has persistent, stateful, SQLite-backed review threads and an
  AI reviewer (`review.go`). The audit graded the deprecated `/classic` surface and
  falsely concluded "no persistence, zero AI."
- **Honest gap:** Two surfaces have diverged (persistent vs throwaway); the AI
  reviewer's output is a free-text tmux dump never parsed back into inline
  suggestions; no intraline highlighting; no keyboard nav; two sequential 20s `gh`
  calls per open with no cache.
- **10-star:** The review cockpit that decides what deserves attention and drives
  the agent until it's fixed — reuse-first, not a rebuild.
- **Moves:**
  1. Consolidate onto the persistent path; retire the `/classic` fire-and-forget
     `onSend`. *(app.js, operator.js, main.go)*
  2. Intraline (word-level) highlighting — confirmed absent. *(diffview.js)*
  3. Keyboard nav `j/k/n/p` + mark-viewed-and-advance. *(diffview.js)*
  4. Auto-resolution: re-read the changed region on the next diff, flip
     `open`→`addressed` (the schema's unused state). *(review_comments.go,
     review_resolve logic)*
  5. Structure the **existing** reviewer output into inline `file:line` suggestions
     seeding the existing comment store — **no new endpoint**. *(review.go,
     diffview.js)*
  6. Cache `computeDiff` per `(dir, HEAD)` with short TTL. *(diff.go 195–215)*
- **Cut:** a new `POST /api/diff/review` (duplicates `/api/review`), localStorage
  persistence (duplicates the SQLite store), the "pending checklist" (already
  ships), the context-expansion backend, split-view LCS alignment.

#### Review comments → agent — grade 3, wow 4

- **Files:** `review_comments.go`, `operator.js` (renderReviewThreads)
- **Verified today:** A thin SQLite CRUD with an `open→sent→addressed` lifecycle.
  Not threaded (no replies). The `open` staging is vestigial (create immediately
  send-alls). `formatReviewComments` drops `side` and quotes no hunk. **Two
  divergent send paths** — `app.js` sends raw via `/api/send` with a different
  format and no persistence.
- **Honest gap:** Two implementations of one feature (one non-persisting), and the
  agent gets `file:123 — text` with zero code context.
- **10-star:** Reliable, context-rich, re-nudgeable delivery — not a "review
  conversation" subsystem.
- **Moves:**
  1. Unify the two paths onto the persisted endpoint (delete the raw string,
     re-add the clipboard fallback, resolve the dead `open` staging). *(app.js
     896–906, operator.js 786–798)*
  2. Capture the ± hunk at `openEditor` time; include `side` + snippet (**drop**
     group-by-file). Rewrite the test that pins the impoverished format.
     *(review_comments.go 88, diffview.js)*
  3. Add a **re-send/nudge** button — send-all filters `state=='open'`, so a `sent`
     comment can never re-route today (the real loop-closure gap). *(review_comments.go 201)*
- **Cut:** rank/drag-reorder/ordinal column, the auto-resolution re-diff loop
  (the "money move" is the fragile one), pane-scrape reply capture.
- **Note:** delivery is `applyKeyAction(text)+Enter` with no agent-busy check — a
  mid-turn blast can corrupt the live prompt; `line` is never rebased, so anchors
  drift after the first edit.

---

### C · Launching

#### Launch agent — grade 4, wow 4

- **Files:** `spawn.go`, `operator.js` (newAgent, wireRepoPicker)
- **Verified today:** A searchable repo combobox and a robust `sendInitialPrompt`
  state machine (dismisses the trust dialog, waits for REPL-ready, verifies
  keystrokes) are genuine craft. But the model dropdown is Claude-only yet shown
  for every agent (`codex --model sonnet` dies in a detached pane), worktree
  isolation is opt-in and **off** by default (agent runs in the live checkout),
  and the worktree is detached-HEAD with no branch.
- **Honest gap:** The default path has a live-checkout footgun, a worktree-leak
  ordering bug, and single-line-only prompts.
- **10-star:** A single Claude launch that is bulletproof and safe-by-default
  before the surface widens.
- **Moves:**
  1. Gate claude-only controls: reject/hide model + resume when `agent!=claude`
     (backend already gates resume). *(spawn.go, operator.js)*
  2. Server-side safety guards: agent binary on PATH (not just tmux), dirty-tree →
     default isolate, specific "name already running" error. Default single-launch
     worktree **ON** (chains/graphs already default ON). *(spawn.go, operator.js)*
  3. Name the worktree branch (`worktree add -b rook/<name>`); **fix the ordering
     leak** — check tmux name uniqueness *before* creating the worktree, or roll it
     back on failure. *(spawn.go)*
- **Cut:** "Attempts 1/2/3" fan-out (duplicates chains/graphs), per-agent model
  catalogs for beta agents, the base-branch picker, the streaming lifecycle
  narration, model-from-prompt suggestion.
- **Note:** multi-line prompts are silently flattened (`strings.Fields`) — code
  blocks, lists, and the agentdocs preamble lose structure.

#### Follow repo instructions — grade 3, wow 3

- **Files:** `agentdocs.go`
- **Verified today:** Detects AGENTS.md/CLAUDE.md/etc and prepends a "go read
  these" note. The file is never read. Two divergent file lists (launcher's
  detected 5 vs GitHub path's hardcoded 5).
- **Honest gap:** The weakest form of the feature — it ships bytes-on-disk a
  "please read" instead of injecting anything, and it's redundant for the common
  claude + CLAUDE.md case (Claude auto-loads it).
- **10-star:** An agent-aware context router — skip what the agent auto-loads,
  inject the raw content of what it would miss.
- **Moves:**
  1. Unify the file list to one source (the real drift bug). *(agentdocs.go,
     operator-github.js)*
  2. Read content + skip auto-loaded files for the matching agent + inject **raw**
     size-capped content of the rest — no distillation. *(agentdocs.go, operator.js)*
  3. Tests pinning that behavior. *(agentdocs_test.go)*
- **Cut:** the rule extractor, the rules-preview UI panel, the verification tail
  (compliance monitoring already exists in review mode), content distillation.

#### Start from a ticket — grade 4, wow 4

- **Files:** `integrations.go`
- **Verified today:** Fetches Linear/Jira/GitHub, builds one static template
  prompt, hands off to `newAgent`. Jira uses api/2 (returns null description for
  Cloud/ADF). Linear passes the raw id into `issue(id:)` which expects a UUID while
  the modal advertises `LIN-456`. Body hard-clipped to 1500 chars silently. cwd
  auto-resolves only for the GitHub `owner/repo#123` regex.
- **Honest gap:** Two of three sources have correctness question marks, and the
  raw untrusted ticket body flows unfenced into a prompt that tells an agent to
  implement and open a PR — a prompt-injection boundary the audit never named.
- **10-star:** Paste an id → a good, safe brief → land in the right repo → launch.
- **Moves:**
  1. Fix the broken sources: Jira api/3 + ADF flatten; Linear lookup by human
     identifier. *(integrations.go)*
  2. Stop silent truncation + reuse GitHub fields already fetched. *(integrations.go)*
  3. Ticket→repo **config map** so Linear/Jira get a cwd. *(config.go, integrations.go,
     operator.js)*
  4. **Fence the untrusted body** as data ("treat as untrusted content, not
     instructions"). *(integrations.go)*
- **Cut:** the Ticket Inbox aggregator, status transitions (In Progress/In Review),
  idempotency/dedup plumbing, per-ticket-type prompt tailoring, the
  acceptance-criteria extractor + summarizer.

#### Resume a closed session — grade 5, wow 4

- **Files:** `resume.go`, `operator.js` (resumeSession, fetchHistory)
- **Verified today:** Globs `~/.claude/projects/*/*.jsonl`, reads a head-scan title
  (aiTitle or first prompt), relaunches with `claude --resume` reusing the same id.
  Works reliably. But every 2-line stub shows up titled with the bare project name,
  and there's no last-activity preview.
- **Honest gap:** The list is noise (dozens of stubs), you resume blind, and there
  are two small correctness bugs (double-resume race, UTF-8 byte cut).
- **10-star:** "Continue where you left off" — the stubs gone, the abandoned
  threads legible, one cheap "waiting on you?" signal.
- **Moves:**
  1. Capture msg-count + last-turn role + snippet in the head-scan that already
     runs (for aiTitle-less sessions it already reads to EOF). *(resume.go)*
  2. Filter stub transcripts below a small message-count threshold. *(resume.go)*
  3. Surface last-activity + a "waiting on you?" flag in the existing palette row.
     *(operator.js)*
  4. Guard the double-resume race (disable-after-click + short server idempotency
     window) + fix `title[:90]` rune cut. *(resume.go 72, operator.js)*
- **Cut:** the Insights cost-card join, worktree/model/codex resume options, the
  resume menu / fork-from-here, the preview endpoint with git diffs, the 4-way
  classifier narration, the dead-cwd repo picker.
- **Note:** codex sessions live in `~/.codex/sessions` and never appear in this
  list at all — the "Claude-only" framing is misdirected here.

#### GitHub view — grade 5, wow 4

- **Files:** `github.go`, `operator-github.js`
- **Verified today:** Read-only `gh` proxies + a repo/issue/PR browser. The PR
  branch/review prompt is well-crafted. But `statusCheckRollup` is fetched and
  **discarded** (red PR looks like green), merge is hardcoded squash, and the view
  never cross-references live sessions. The classic `app.js` page already has
  agent-awareness, batch hand-off, and a tmux fallback — the new operator view
  **regressed** and dropped them.
- **Honest gap:** Nothing is missing in the codebase — session state,
  `ghRefFromSession`, CI data on the wire, a method-accepting merge endpoint all
  exist; the new view simply hasn't wired them.
- **10-star:** An agent-aware launch-and-track surface — port the proven pieces.
- **Moves:**
  1. Port agent-awareness from the classic page (`sessionForItem` + jump-badge +
     double-launch warn). *(operator-github.js, operator.js)*
  2. Render the CI checks pill from the already-fetched `statusCheckRollup`.
     *(github.go 203, operator-github.js)*
  3. Wire the merge method picker through + `--delete-branch`. *(operator-github.js
     119, integrations.go 76)*
  4. Add the tmux-availability fallback the classic page has (Work/Review silently
     fail without it). *(operator-github.js 108)*
- **Cut:** the inline PR-review client, the merge-when-green polling queue, the
  four-bucket taxonomy + filter chips, batch "Work all N" as net-new (already
  exists in classic — a free port at most).

---

### D · Orchestration

#### Board (kanban) — grade 3, wow 4

- **Files:** `board.js`
- **Verified today:** Re-buckets sessions into 5 columns by derived state. Ships
  ~35 lines of **dead drag scaffolding** and a header calling itself "draggable" —
  drag is a *documented rejection* (board.js:143-145: "you can't force a state by
  dragging"), not an unfinished feature. `columnFor` is duplicated in `app.js`.
  "Review" is a `/.rook/worktrees/` substring test.
- **Honest gap:** It orchestrates nothing — every card action teleports to the
  sessions view. Its only non-redundant asset is at-a-glance column triage.
- **10-star:** Make the columns earn their keep (rank within them, keep assignment
  honest) — don't grow a second copy of the sessions page.
- **Moves:**
  1. Delete the dead drag scaffolding + fix the "draggable" lies + delete `app.js`
     `boardColumn` (one `columnFor`). Remove the phantom `onMove` cb. *(board.js,
     app.js)*
  2. Rank "Needs you" by urgency (alert first, then longest-waiting) — pure client
     sort. *(board.js)*
  3. Make Review honest (gate on a cheap branch-ahead flag if it already rides the
     payload) or relabel the column "Worktrees." *(board.js)*
- **Cut:** real HTML5 drag + `onStartStep` backend (contradicts the documented
  decision), WIP limits + column aging (unenforceable), the full git ahead/PR-state
  backend, the act-next highlight + keyboard shortcut.

#### Task chains — grade 3, **fold → graphs**

- **Files:** `chains.go`
- **Verified today:** A 211-line in-memory linear runner. No context hand-off (each
  step's prompt is fixed at create time). No failure semantics —
  `advanceChain` marks running→done **unconditionally** (a crashed agent reports
  "success"). No persistence. Advances on **any** Stop in the CWD. It is a strict,
  buggier subset of `graph.go`.
- **Honest gap:** It exists mainly to be a separate line item on the board's Queued
  column, and its unconditional done-marking is a silent success-on-failure
  landmine.
- **10-star:** Delete the separate engine; "chain" becomes a linear preset over the
  DAG, inheriting persistence, verify-gating, and failure propagation for free.
- **Moves:**
  1. Reimplement `handleChainCreate` as a thin adapter → `taskGraph` with implicit
     `node[i] depends on node[i-1]`; delete the chain engine + tests + modal +
     palette entry; migrate rows on boot. *(chains.go delete, graph.go, main.go,
     review.go 229)*
  2. Minimal **one-line** hand-off packet (files-touched by N-1 + verify verdict —
     not `summaries.go`, not full diffs). *(graph.go)*
  3. Fix the graph.go session-correlation bug this inherits (below). *(graph.go)*
- **Cut:** the retry/skip/edit-prompt intervention UI + narration events, the
  dual-frontend modal merge (defer — cosmetic next to the backend fold).

#### Task graphs (DAG) — grade 4, wow 4

- **Files:** `graph.go`, `operator-graph.js`
- **Verified today:** A genuinely correct scheduler — conditional edges, transitive
  skip-propagation, approval nodes, SQLite checkpointing with orphan recovery (all
  tested). But it is **not parallel** (single `agentRunning`, shared worktree), and
  node success is "tmux pane idle = pass" — a plain agent node can never emit
  `fail`, so the headline conditional fail-edges are inert. It duplicates
  `chains.go`.
- **Honest gap:** The defining advantage over the chain (concurrency) is absent,
  and the conditional edges are decorative.
- **10-star:** Actually parallel, with real pass/fail so the edges mean something —
  the survivor of the fold.
- **Moves:**
  1. **N-way parallel** dispatch: per-branch concurrent launch, each independent
     branch in its own worktree, join = wait for all upstreams. Manual merge gate,
     **not** auto-conflict-resolution. *(graph.go, graph_test.go)*
  2. Real per-node pass/fail via a `SUCCESS/FAIL` marker contract in the prompt,
     parsed in `completeGraphNode`. Drop the judge node type + retry engine.
     *(graph.go)*
  3. Per-node timeout + heartbeat (fold in the completion-soundness fixes below).
     *(graph.go)*
  4. Node **drill-in only** (click → transcript/diff/status). *(operator-graph.js)*
- **Cut:** replay/fork + checkpoints table, NL-goal-to-DAG, judge node type,
  retry-with-repair engine, layout crossing-minimization, critical-path highlight,
  live-activity-on-edges.
- **Correctness bugs to fix here:** `advanceGraphsForCWD` completes the running node
  of **every** graph/chain whose runDir==cwd (keyed on directory, not the session
  that stopped) — worse with parallelism; the sub-4s debounce silently loses fast
  nodes; `runVerifyForGraph` returns **pass** when no gate is detected; joins are
  AND-only (no OR-join).

#### Workspace — worktrees & hooks — grade 4, wow 4 *(split, don't merge)*

- **Files:** `worktree.go`, `hooks.go`, `operator-workspace.js`
- **Verified today:** Two unrelated halves on one screen. The worktree half is a
  passive janitor: `os.ReadDir` (not `git worktree list`), a dead "Branch" column
  (worktrees are `--detach`, so it shows "HEAD"), and a force-delete with no dirty
  check. The hooks half is legitimately functional — real-time event ingestion, a
  working destructive-command gate, auto-review on the Stop hook.
- **Honest gap:** The delete **lies** (`removeWorktree` discards every error, the
  handler always returns `{ok:true}` — green "Removed" toast over a no-op), and it
  discards uncommitted work silently.
- **10-star:** Two honest halves — a git-authoritative worktree list with one real
  action (Review), and a hooks/events feed with click-through.
- **Moves:**
  1. Fix the lying delete (return errors, stop hardcoding `{ok:true}`) + dirty
     guard before `--force`; fix the prefix-without-separator path guard
     (`worktrees-evil` passes). *(worktree.go 46, 114, 117)*
  2. Git-authoritative list (`git worktree list --porcelain`) + name the branch at
     the source (drop `--detach` in `createWorktree`); cheap dirty state (leave
     ahead/behind out of the poll — use `diff.go` on demand). *(worktree.go, spawn.go)*
  3. Add one inline action — **Review** — reusing `/api/diff`. Demote Remove.
     *(operator-workspace.js)*
  4. Status pill from dirty + owning-session liveness (`sessionForPath` already
     links the agent). *(worktree.go, operator-workspace.js)*
  5. Events feed: add session id to `hookRecord`, wire click-through via
     `ctx.selectAgent`. *(hooks.go, operator-workspace.js)*
- **Cut:** the merge/PR/push shipping pipeline, setup-script/env-copy/new-worktree
  board control, the full policy engine (rule editor/scope/dry-run/rollup),
  ahead/behind baked into the poll, archive-as-default.
- **Note:** `inUse` is an exact `CWD==path` match, so an agent that `cd`'d into a
  subdir reads as not-in-use and the Remove button is enabled on a live worktree.

---

### E · Intelligence

#### Insights — cost & tokens — grade 4, wow 4 *(verdict: solid)*

- **Files:** `usage.go`, `operator.js` (renderInsights)
- **Verified today:** Richer than the brief — per-model aggregates, two token
  windows, a 30-day trend, and a genuine differentiator no competitor has: the
  **quality × cost pairing** per session/model/project. But two verified
  correctness bugs: Codex sessions render **$0** (never call `pricePerToken`), and
  `$/M-tok` divides a cost that includes cache tokens by a `TokensTotal` that
  doesn't. There's no cost-over-time line.
- **Honest gap:** A whole provider is silently undercounted, the headline
  efficiency metric is mechanically inconsistent, and everything is passive.
- **10-star:** A cost dashboard that acts — reliable cross-provider cost, spend-
  over-time, and a narration strip built on the quality × cost spine.
- **Moves:**
  1. Fix cross-provider cost: capture the token split in `parseCodexRollout`, add
     OpenAI rates, route by provider. *(pricing.go, scan.go)*
  2. Close the `~/.codex` ingestion gap in `TokenWindows`/`ScanTrends` (glob only
     `~/.claude` today) and fix the `$/M-tok` denominator. *(scan.go 271,728,814,862)*
  3. Add `Cost` to `DayStat` for a spend-over-time line + WoW chip. *(scan.go,
     operator.js)*
  4. Narration strip reusing the quality × cost pairing (gated on min sample +
     dollar threshold). *(usage.go, operator.js)*
  5. Burn-rate + projected-exhaustion on the 5h window — **drop** the plan-cap knob.
     *(scan.go, usage.go, operator.js)*
- **Cut:** the budgets/alerts subsystem, click-to-expand drill-down, the plan-cap
  config knob. (The cache-economics tile is a borderline-keep — cheap, additive,
  and answers the "is cache helping" the broken `$/M-tok` currently botches.)

#### Cheap-model routing — grade 4, wow 4

- **Files:** `config.go`, `summary.go`, `pricing.go`
- **Verified today:** There is no router. Three disconnected pieces: a Haiku default
  for the summary agent, a manual per-launch dropdown, and a static pricing table.
  The rook-owned background spawns — review lenses (`review.go:96`), reflexion
  retries (`review.go:268`), auto-verify — pass **no model** and silently inherit
  the account default (Opus). The Haiku rate is stale (3.5 numbers for a dropdown
  that launches 4.5).
- **Honest gap:** "Routing" is marketing, and the one place a cheap model is
  provably safe (mechanical fix loops) is left on the strong model.
- **10-star:** Deterministic task-type policy for the spawns rook already owns — not
  a learned prompt classifier.
- **Moves:**
  1. Route the owned spawns: review→strong, reflect/verify→cheap, via the existing
     `summaryModel` pattern — **not** a big `taskModelPolicy` map. *(spawn.go,
     review.go, scheduler.go)*
  2. Fix the two stale Haiku 4.5 numbers (and the default-branch mis-bucketing of
     expensive families) + update the golden test. *(pricing.go, pricing_test.go)*
- **Cut:** the `classifyTask` auto-router + `autoRoute` toggle, the realized-savings
  counterfactual, the spawn-modal recommendation + `/api/route-preview`, the
  mid-flight up/downshift nudges (the CLI can't hot-swap mid-session), the
  model-tiering optimization.

---

### F · Automation

#### Auto-verify + Reflexion retry — grade 3, wow 4

- **Files:** `review.go` (runVerify), `reflect.go`
- **Verified today:** Detects one build/test command, runs it with a 4-min timeout,
  records pass/fail. The "Reflexion" branch is far shallower than its doc claim: it
  appends the raw failure tail + a **canned** instruction to `reflections.md` and
  spawns a fresh session — the model's actual reflection is never captured. The
  loop is implicit (relies on the spawned agent's Stop hook re-firing).
- **Honest gap:** It's "paste the test log back," not Reflexion — and two defects
  make it largely non-functional past attempt 1: the reflect session name is
  `reflect-<base>-<HEAD-short>`, so with no commit the next attempt collides on the
  tmux name and the loop **silently dies** (all four return values discarded); and
  the whole markdown buffer is flattened to one line by `strings.Fields`.
- **10-star:** An honest, narrated get-to-green loop — or at minimum, stop lying
  about being Reflexion.
- **Moves:**
  1. Process-group kill on timeout (`Setpgid`, kill the group — today only `sh` is
     killed, orphaning `go test`/`jest` children). *(review.go 199)*
  2. Keep `.rook-reflect` out of the reviewed diff (write outside the tree or
     `.git/info/exclude`). *(reflect.go)*
  3. One `VerifyCmd` override applied to **both** the auto and manual paths (auto
     can't override a misdetection today). *(config.go, review.go)*
  4. No-progress exit via output-tail hash equality (stop grinding the cap on
     identical failures). *(review.go, reflect.go)*
  5. Make the reflection real (agent appends its own critique) **or delete the
     "Reflexion (Shinn et al., 2023)" claim**. *(reflect.go, review.go)*
  6. Narrate each attempt through the existing `banner/pushChat`. *(review.go)*
- **Cut:** the distiller sub-call, the full per-ecosystem structured failure parser,
  the dashboard verify/reflex timeline, the multi-step gate ladder.
- **Note:** AutoReview also compounds — `onSessionFinished` re-spawns the review
  subagent on each reflect session's Stop, so a 3-attempt loop spawns up to 3 extra
  review agents.

#### Destructive-command gate — grade 3, wow 3

- **Files:** `hooks.go`
- **Verified today:** Rides a solid, well-tested hooks bridge. The gate itself is 5
  hardcoded regexes against a single input string, deny-only, **defaults off**. It's
  narrower than rook's own advisory list (9 patterns), a block is **silent** (never
  banners/pushes), and the SQL rules are uppercase-only.
- **Honest gap:** The biggest miss: the wrapper is `curl -s -m 5 … 2>/dev/null;
  exit 0` — **fail-open**. When rook is down or curl times out, the command runs.
  A safety feature off precisely when it isn't running.
- **10-star:** A trustworthy gate: on by default, narrating its catches, with one
  case-insensitive ruleset — not a policy platform.
- **Moves:**
  1. Fix the fail-open transport **first**. *(hooks.go writeHookScript)*
  2. Narrate the block via existing `banner()/pushNtfy/pushChat`. *(hooks.go)*
  3. One case-insensitive Go ruleset (fixes the 5-vs-9 drift + the lowercase
     `drop table` miss); regenerate/sync the JS advisory list from it — no live
     endpoint. *(hooks.go, operator-audit.js)*
  4. Default the gate **ON**; catastrophic-only coverage (add `dd of=/dev/*`,
     `mkfs`) + a cheap `&& ; |` segment split + regression tests pinning the known
     holes and the wired `handleHook` decision. *(hooks.go, config.go, hooks_test.go)*
- **Cut:** the `mvdan.cc/sh` shell-AST dependency, the allowlist UI subsystem, the
  `/api/danger-rules` endpoint, blanket sudo/kubectl/docker blocking, the ask-tier
  severity engine.
- **Note:** the matcher runs on **all** tools and reaches into `content/new_string`,
  so an Edit/Write whose data *contains* `rm -rf /` gets denied — a false-positive
  vector.

#### Auto-review — grade 3, wow 4

- **Files:** `review.go`
- **Verified today:** On the Stop hook, spawns N read-only review agents in a
  sequential loop, each with a rotating lens, each told to emit a distilled review
  leading with `SHIP / FIX-FIRST / BLOCK`. Three shallow truths: **no synthesis**
  (N separate tmux transcripts), the **VERDICT is never parsed** (appears only in
  the prompt string), and it's disconnected from `review_comments`. Only the first
  pass's spawn error is checked.
- **Honest gap:** The feature asks the model to decide, then discards the decision
  — and it can review the wrong thing: the gate uses fork-base `diffBase` but the
  reviewer runs only `git diff` (uncommitted), so committed work is reviewed as an
  empty diff → a confident SHIP on nothing.
- **10-star:** One structured, parsed verdict that's discoverable — not three chat
  transcripts.
- **Moves:**
  1. Parse the verdict + findings from the transcript into a typed struct.
     *(review.go)*
  2. Notify on completion reusing the verify notification path (auto-verify fires
     banner/ntfy/chat; auto-review fires none). *(review.go 246–250)*
  3. Stop swallowing passes 2–3 (aggregate errors, not `firstErr`). *(review.go 97–99)*
  4. Default to a **single consolidated pass** (decide vs reconcile). *(review.go)*
  5. Fix the bogus-SHIP scope mismatch and the stale/no-review on same-HEAD re-runs
     (`reviewStamp` is HEAD-only; ignores the uncommitted tree it's reviewing).
     *(review.go 129, 219)*
- **Cut:** the synthesizer, materialize-into-comment-threads, the verdict pipeline
  gate + reflexion re-injection, model tiering.

#### Audit — command log — grade 4, wow 4

- **Files:** `audit.go`, `operator-audit.js`
- **Verified today:** A background ingester INSERT-OR-IGNOREs Bash/Shell calls into
  SQLite; a genuinely polished view (debounced search, selection-preserving poll,
  XSS-escaped). But "click to jump" only selects the agent and lands on Overview —
  it never scrolls to or highlights the command. Bash-only (Edit/Write invisible).
  Dedup key `session|ts|firstN(cmd,80)` drops distinct commands sharing an 80-char
  prefix.
- **Honest gap:** An "audit trail" that is shell-only and silently loses records —
  the ingester polls every 60s over a 40-tool window, so a burst >40 calls loses the
  middle ones permanently.
- **10-star:** A trustworthy, navigable record — capture mutations, stop losing
  records, make the click honest.
- **Moves:**
  1. Widen ingestion to Edit/Write/MultiEdit — **no** severity/reviewed columns.
     *(audit.go)*
  2. Give `ToolCall` a stable per-call id; dedup on `session|id` (fixes the
     collision and the silent drop). *(scan.go, audit.go)*
  3. Make the click honest — relabel "select agent," or size a real jump correctly
     (the Trace tab is a 40-cap chart requiring ≥2 calls, **not** a browsable list,
     so a genuine jump is a Trace rebuild). *(operator.js, operator-audit.js)*
- **Cut:** graded severity 0-3, the ack/review queue, the narration banner, the
  risky-only toggle, JSONL/CSV export, the PreToolUse guardrail (a separate feature).

---

### G · Everyday

#### Keyboard & command palette — grade 4, wow 4

- **Files:** `operator.js` (palette, GO_KEYS)
- **Verified today:** All bindings work. The palette merges four groups but filters
  with a plain substring test — no fuzzy, no ranking, no matched-char highlight —
  and the active item has **no `scrollIntoView`** (the same file does this for its
  repo dropdown). It navigates but never acts.
- **Honest gap:** 2015-era matching and inert — for an operator tool, the palette
  can't answer a blocked agent, stop it, or approve a gate.
- **10-star:** Fast correct matching, a highlight that stays visible, and one real
  action — not a palette state machine.
- **Moves:**
  1. `scrollIntoView({block:'nearest'})` on the active item (a consistency bug —
     the roster and repo dropdown already do it). *(operator.js)*
  2. Fuzzy subsequence scorer + `<mark>` highlight (keep exact-prefix pinned on
     top). **No frecency.** *(operator.js renderPaletteList 1358–1385)*
  3. Generate the `?` sheet from `GO_KEYS` so it can't drift. *(operator.js)*
  4. Inline answer/allow-deny on live-agent items reusing `respond()` — closes the
     "inert" gap without the ranker or verb menu. *(operator.js)*
- **Cut:** the second-level 6-verb menu (`palLevel` state machine), the
  operator-aware ranker + narration, frecency, the live chord overlay.
- **Note:** no ARIA/listbox semantics on a keyboard-first surface; `fetchHistory`
  fires on every open (Resume items flash in late); hover doesn't sync `palIdx`.

#### Notifications — grade 5, wow 4

- **Files:** `notify.go`
- **Verified today:** Three independent triggers (poll loop, hooks, review) wired
  **inconsistently**: `notifyWaiting/notifyStuck` send banner + ntfy but never
  Slack/Discord; `notifyFinished` sends banner only. Escalation is fake —
  `notifyStuck` sends a byte-identical ntfy request to a routine ping, differing
  only in a macOS sound name. `pushNtfy` sets no `Priority`/`Actions`/`Click`. Reads
  a legacy `FOREMAN_NTFY`.
- **Honest gap:** Escalation doesn't escalate, failures are silent, and a genuinely
  abandoned agent pings once at 10m then goes quiet forever.
- **10-star:** Reliable and truly escalating — not a remote-control subsystem.
- **Moves:**
  1. Real escalation: `notifyStuck` sends `Priority: urgent`. *(notify.go 103)*
  2. Consolidate the scattered per-event channel choices into one routing point so
     loop-detected waiting/stuck *can* reach chat — **without** flattening every
     event to every channel (device fan-out is the point; `finished` stays a quiet
     banner). *(notify.go, integrations.go)*
  3. `Click` deep-link to the local UI (delivers "actionable" for free via
     `/api/send`). *(notify.go)*
  4. Rot cleanup: drop `FOREMAN_NTFY` + the hardcoded `construction_worker` tag.
     *(notify.go)*
- **Missed-gap fixes:** `notifyFinished` fires only on busy→idle, so completion from
  a **waiting** state (the common case, right after you answer) is never reported —
  fix the branch; log non-2xx instead of swallowing; re-alert on an interval while
  still stuck (the `escalated` map fires once and never again). *(notify.go 30, 73, 80)*
- **Cut:** inbound action endpoints + Slack app interactivity, Block Kit,
  cross-device "dedup," the danger-scoring ranker.

#### Daily summaries — grade 5, wow 4

- **Files:** `summary.go`, `scheduler.go`, `operator-summaries.js`
- **Verified today:** A fire-and-forget agent wrapper — the backend does zero
  synthesis; it spawns a cheap Claude session with a ~13-line prompt telling the
  agent to run ~10 `gh` calls, dedup, and POST markdown back. The reading view is
  polished. `buildSummaryPrompt` interpolates the `saveURL` params with no escaping
  (Go-side); the operator view regressed to single-day.
- **Honest gap:** Correctness outsourced to the cheapest model orchestrating ~10 API
  calls, with **zero observability** — clicking Generate yields a toast; if the
  agent blocks on a permission prompt, the summary silently never appears. And the
  "twin" prompt builder is alive in `app.js` and has already **drifted** (JS escapes
  the saveURL, Go doesn't).
- **10-star:** rook owns the facts (a deterministic spine) and the model narrates —
  reliable, observable, cheap.
- **Moves:**
  1. Escape the `saveURL` query params **Go-side only** — the JS twin already uses
     `encodeURIComponent`, and the gh-command filters need shell quoting, **not**
     URL-encoding. *(scheduler.go)*
  2. Consolidate the two prompt builders into one server source (the "keep in sync"
     comment is accurate and the hazard already bit). *(scheduler.go, app.js 1050)*
  3. Surface the spawned session + failure — the session name is **already
     returned**; the modal throws it away. Render a "generating…" placeholder that
     flips to "failed" if the session dies with no matching POST. *(scheduler.go,
     summary.go, operator-summaries.js)*
  4. Restore the date range the operator view dropped (presets exist in `app.js`).
     *(operator-summaries.js)*
  5. Catch-up backfill on startup — persist `lastRun` (in-memory today re-fires on a
     restart within the scheduled minute). *(scheduler.go, config.go)*
  6. The deterministic factual **spine** (git log + gh + `/api/claude-activity` →
     structured object the model narrates) — justified **only** by reliability +
     cost. *(scheduler.go, summary.go)*
- **Cut:** day-over-day/week-over-week deltas + sparkline + streak + impact ranking,
  cadence toggles + Slack/push/email delivery, per-task session-linking now (park it
  as the reward once the spine lands).
- **Note:** zero tests, in a repo whose CLAUDE.md mandates TDD + 90% coverage — the
  escaping bug is a one-line table test that would have caught it.

---

*End of plan. Build any phase top-to-bottom; the moves are self-contained and
name their files. When in doubt, prefer connecting what exists over adding a new
surface — that is the entire thesis.*
