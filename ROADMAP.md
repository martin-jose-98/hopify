# 🎮 Go TUI Roadmap — Progress Tracker

> Gamified progression: **Quest → XP → Level → Boss Battle**.
> ~1h/day, 4-5 days/week. Target: **3 months**.

## Status Snapshot

- **Current Phase:** Phase 0
- **Current Level:** 1 (Novice Novice)
- **Total XP:** 0 / 360
- **Completed Quests:** 0
- **Bosses Slain:** 0
- **Achievements:** none yet

---

## Mechanics

| Element | Meaning |
|---|---|
| Quest | Focused task with a checkable deliverable |
| XP | Check a quest → earn its XP |
| Level | Every 100 XP = level up |
| Boss | Capstone quest, 1 per phase |
| Achievement | Side badge for a behavior/habit |

Level map: 1 = 0 XP · 2 = 100 · 4 = 200 · 6 = 300 · 8 = 400 · 10 = 500 (XP shown are cumulative targets; per-quest XP listed below).

Bonus XP (+5 each): commit daily · review last week's code · explain your app to someone.

---

## Phase 0 — Level 1: Foundations (Week 1)

### Quests

- [X] **Q0.1** — Install Go, `go version`, hello world — `5 XP`
- [ ] **Q0.2** — Tour of Go through step 25 (control flow) — `5 XP`
- [ ] **Q0.3** — CLI tool: flags, stdin, exit codes — `10 XP`
  - Deliverable: `greet --name Bob`
- [ ] **BOSS 0 — Static Dashboard: terminal "dashboard" layout** — `10 XP`

**Required to advance:** 3/4 quests incl. boss. **Gib:** working `go.mod` + first binary.

<!-- PHASE-BREAK -->

## Phase 1 — Level 2-3: Text & Color (Week 2-3)

### Quests

- [ ] **Q1.1** — `string`, `fmt`, pointers — `10 XP`
- [ ] **Q1.2** — Install **Lipgloss**; boxed/colored terminal card — `5 XP`
- [ ] **Q1.3** — Static styled menu panel — `15 XP`
- [ ] **Q1.4** — `struct` + `interface` — reusable `Panel` type — `10 XP`
- [ ] **BOSS 1 — Static Dashboard: build it with Lipgloss** — `10 XP`

**Achievement at 3+ style combinators:** "Color Wizard" 🎨
**Required to advance:** 4/5 quests incl. boss.

<!-- PHASE-BREAK -->

## Phase 2 — Level 4-5: Interaction (Week 4-5)

### Quests

- [ ] **Q2.1** — Bubble Tea ELM model (`Init`/`Update`/`View` hello app) — `15 XP`
- [ ] **Q2.2** — Keyboard input + graceful quit — `15 XP`
- [ ] **Q2.3** — Form: `Input` / basic components — `10 XP`
- [ ] **Q2.4** — Concurrency: `goroutine`, `channel` ticker — `15 XP`
- [ ] **Q2.5** — `context` for timeouts/cleanup — `10 XP`
- [ ] **BOSS 2 — Interactive To-Do list: add/list/check/delete + save to file** — `10 XP`

**Achievement:** "Elm-ish" 🧠 (state-machine app).
**Required to advance:** 5/6 quests incl. boss.

<!-- PHASE-BREAK -->

## Phase 3 — Level 6-7: Real Apps (Week 6-8)

### Quests

- [ ] **Q3.1** — Build same app in `tview` AND `bubbletea`; compare — `20 XP`
- [ ] **Q3.2** — Networking: `http` + `encoding/json` → API data in TUI — `15 XP`
- [ ] **Q3.3** — Spinner/progress views (`bubbles`) — `10 XP`
- [ ] **Q3.4** — Error handling patterns for TUIs — `10 XP`
- [ ] **Q3.5** — Table-driven tests for your model code — `10 XP`
- [ ] **BOSS 3 — API Browser TUI: search API, styled results** — `10 XP`

**Achievement:** "Resilient" 🛡️ (app never panics on bad input).
**Required to advance:** 5/6 quests incl. boss.

<!-- PHASE-BREAK -->

## Phase 4 — Level 8-9: Polish (Week 9-10)

### Quests

- [ ] **Q4.1** — Keymaps, mouse, resize handling — `10 XP`
- [ ] **Q4.2** — Themes + lipgloss polish + spinners — `10 XP`
- [ ] **Q4.3** — `log/slog` + logging/debug strategies — `10 XP`
- [ ] **Q4.4** — `go vet`, `staticcheck`, lint — zero warnings — `10 XP`

**Required to advance:** all 4 quests.

<!-- PHASE-BREAK -->

## Phase 5 — Level 10: FINAL BOSS (Week 11-12)

- [ ] **BOSS 5 — Build YOUR own TUI tool — `50 XP`**

Ideas: GitHub issues CLI · finance tracker · pomodoro + tasks · RSS reader.

Clear condition (all required):
- [ ] 2+ of: networking, persistence, concurrency
- [ ] Lipgloss + bubbletea or tview
- [ ] Graceful error handling + `SIGINT`
- [ ] ≥ 5 meaningful tests
- [ ] Installed as a binary

---

## Progression Track

| Level | Title | Unlocks |
|---|---|---|
| 1 | Novice Novice | Phase 0 |
| 2 | Terminal Tinkerer | Phase 1 |
| 4 | Style Weaver | Phase 2 |
| 6 | Input Master | Phase 3 |
| 8 | Systems Guy | Phase 4 |
| 10 | TUI Sage | Final Boss capacity |

---

## Achievements

- [ ] **Color Wizard** 🎨 — Phase 1
- [ ] **Elm-ish** 🧠 — Phase 2
- [ ] **Resilient** 🛡️ — Phase 3
- [ ] **Ship Early** 🚢 — commit daily for 7 straight days
- [ ] **TUI Sage** 🏆 — clear Final Boss

---

## Daily Log

| Date | XP earned | Quest(s) done | Notes |
|---|---|---|---|
| — | — | — | — |

---

## How to update

When quests are completed, check the box, add to Daily Log, and update **Status Snapshot** (phase, level, XP, bosses). Level = XP/100. Next phase unlocks when required quests are done.
