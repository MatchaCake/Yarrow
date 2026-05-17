# Yarrow Demo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the standalone hackathon demo described in `hackathon/ralph-spec.md` and push it to `git@github.com:MatchaCake/Yarrow.git`.

**Architecture:** The demo is a single Go module, `yarrow/demo`. The pure `algo` package implements scoring and aggregation, `algoref` generates Python reference fixtures, and `cmd/demo` serves an in-memory HTTP API plus one vanilla HTML file.

**Tech Stack:** Go 1.23, Python 3.14 virtualenv with numpy and pytest, Go stdlib HTTP server, vanilla JS, Tailwind CDN.

---

## File Map

- `data/agents.json`: exact five agent personas from the hackathon spec.
- `data/markets.json`: exact ten curated market snapshots from the hackathon spec.
- `algoref/reference.py`: Python formulas for scoring and aggregation.
- `algoref/generate.py`: fixture generator writing `fixtures/input` and `fixtures/expected`.
- `algoref/tests/test_reference.py`: sanity coverage for the Python reference.
- `algo/doc.go`, `algo/score.go`, `algo/aggregate.go`: pure Go scoring and aggregation package.
- `algo/score_test.go`, `algo/aggregate_test.go`, `algo/diff_test.go`: Go unit and fixture diff tests.
- `cmd/demo/main.go`, `cmd/demo/store.go`, `cmd/demo/handlers.go`: embedded HTTP demo.
- `cmd/demo/index.html`: single-file frontend.
- `README.md`: written last, after verification.

## Tasks

### Task 1: Bootstrap

- [ ] Copy `hackathon/agents-personas.json` to `data/agents.json`.
- [ ] Copy `hackathon/markets-snapshot.json` to `data/markets.json`.
- [ ] Run `go mod init yarrow/demo`.
- [ ] Create `algoref/.venv` and install `numpy pytest`.
- [ ] Create `REFERENCES.md` with the three Metaculus references.
- [ ] Verify `algoref/.venv/bin/python -c "import numpy"` exits 0.

### Task 2: Python Reference

- [ ] Write `algoref/tests/test_reference.py` with known-value tests from the spec.
- [ ] Run `algoref/.venv/bin/python -m pytest algoref/ -q` and observe failure because `algoref/reference.py` does not exist yet.
- [ ] Implement `algoref/reference.py`.
- [ ] Re-run `algoref/.venv/bin/python -m pytest algoref/ -q` and require green.

### Task 3: Fixtures

- [ ] Write `algoref/generate.py`.
- [ ] Run `algoref/.venv/bin/python algoref/generate.py`.
- [ ] Verify `ls fixtures/input | wc -l` and `ls fixtures/expected | wc -l` are equal and at least 60.

### Task 4: Go Algo

- [ ] Write Go tests and diff-test harness first.
- [ ] Run `go test ./algo/... -count=1` and observe failure before implementation.
- [ ] Implement `algo` functions with the pinned signatures.
- [ ] Run `go test ./algo/... -count=1 -race` and require green with at least 60 fixture cases.

### Task 5: Demo Server And Frontend

- [ ] Write API/store tests for route behavior.
- [ ] Run `go test ./cmd/demo -count=1` and observe failure before implementation.
- [ ] Implement the in-memory store, HTTP handlers, embedded frontend, and `-addr` flag.
- [ ] Run the spec curl checks for markets sorting, leaderboard, report sorting, and forecast submission.
- [ ] Open the page in the browser and verify market rows, modal, close behavior, and forecast update.

### Task 6: Final Verification And Push

- [ ] Run `go test ./algo/... -count=1 -race`.
- [ ] Run `algoref/.venv/bin/python -m pytest algoref/ -q`.
- [ ] Run `go test ./... -count=1`.
- [ ] Run `go vet ./...`.
- [ ] Run `gofmt -l .` and require no output.
- [ ] Write `README.md`.
- [ ] Commit the finished demo.
- [ ] Push to `git@github.com:MatchaCake/Yarrow.git`.
