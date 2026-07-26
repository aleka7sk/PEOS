# PEOS Progress Tracking Rule

This rule governs how Claude Code must read, trust, and maintain PEOS implementation
progress state. It exists because chat history and memory are not durable records —
`docs/implementation-progress.md` is the single authoritative, persistent source of truth
for what has been implemented, what is in progress, what is blocked, and what is
deliberately deferred. This rule is mandatory for every PEOS implementation, audit, or
architecture task in this repository.

## 1. Read before acting

At the beginning of every PEOS task, read, in this order:

1. `docs/implementation-progress.md` (current state, current packet, current next action);
2. the relevant `spec/0XX-*.md` document(s) for the concept(s) in scope;
3. the relevant `peos/<package>` code, especially its `doc.go` (implemented / deferred
   status is documented there per-concept);
4. relevant `git log` / `git show` history for the packages in scope, when the task
   requires knowing what changed and when.

Do not begin substantive work — architecture, implementation, or audit — without having
read `docs/implementation-progress.md` first.

## 2. Confirm scope before implementing

Before starting implementation work, confirm that the requested task matches the
"Current Next Action" recorded in `docs/implementation-progress.md`. If it does not
match:

- state the conflict explicitly to the user (what the tracker says the next action is,
  versus what is being requested);
- proceed with the user's actual request once stated — this is a disclosure obligation,
  not a blocking one — but do not silently treat the mismatched request as if it were the
  recorded next action.

## 3. Never trust checkboxes blindly

A checkbox in `docs/implementation-progress.md` is a claim, not a fact. Before relying on
a `[x]` (or any other state) to make a decision:

- verify implementation existence by reading the actual file(s), not by trusting the
  tracker's file-path column;
- verify tests exist and pass by running them (`go test ./peos/<package>/...`), not by
  trusting a cached pass/fail result;
- verify quality gates by running them, not by citing a prior run's output as current;
- verify an audit's acceptance by checking whether this file (or another repository
  artifact) actually records an accepted verdict — a verdict that only ever existed in a
  prior chat transcript is not repository evidence until this file is updated to record
  it.

If repository evidence contradicts a recorded state, correct the tracker immediately as
part of the current task and note the correction in the final report's progress update.

## 4. Update after every substantive PEOS action

After every architecture plan, implementation, audit, or corrective packet, update
`docs/implementation-progress.md`:

- update the `Last updated` field to the current date;
- update the affected packet's status row (Packet Progress table) and, if applicable,
  the affected specification's row (Specification Progress table);
- update the Quality Gates table with freshly run results — do not carry forward a gate
  result from a previous task without re-running it if the code changed;
- update the Deferred Architecture section if a deferral was newly introduced, resolved,
  or its blocking dependency changed;
- set exactly one "Current Next Action" reflecting the new state (see rule 12 for the
  format required in the final report).

## 5. Writing code is not completion

Never mark a packet `[x]` merely because code was written. Code existing is one of four
required conditions (rule 6), not the whole of it.

## 6. Definition of packet completion

A packet may be marked `[x]` Complete and accepted only when repository evidence shows
all four of the following:

1. its planned implementation exists (files present, matching the approved architecture);
2. required tests exist and pass (`go test` run, not assumed);
3. required quality gates pass (`gofmt`, `go build`, `go vet`, `go test -race`,
   `staticcheck`, `golangci-lint` where available — do not install a missing tool, just
   note it as unavailable);
4. any required audit has an accepted verdict, and that verdict is recorded in
   `docs/implementation-progress.md` itself (not only in a chat transcript).

An audit performed within the current conversation counts as repository evidence **only
once this rule's update step (rule 4) has written its verdict into the tracker file** —
until that write happens, treat the audit as unrecorded.

## 7. Audit-pending state

When implementation is complete but its required audit has not yet run or has not yet
returned a verdict, mark the packet:

```
[~] Implemented — audit pending
```

Do not use `[x]` for this state, even if the implementation itself looks correct.

## 8. BLOCKER / MAJOR findings

When an audit finds one or more BLOCKER or MAJOR findings:

- mark the packet `[!]`;
- record every such finding in the Packet Progress table's "Remaining findings" column,
  with enough detail to act on it without re-reading the original audit;
- set the corrective packet (name it explicitly, e.g. "Packet F.3.1 — <exact corrective
  scope>") as the new "Current Next Action" — a corrective packet always takes priority
  over the next sequential packet.

## 9. MINOR findings stay visible

A MINOR finding does not block a packet from being marked `[x]` (see precedent: Packet
F.2's one MINOR finding did not block its ACCEPT verdict), but it must remain listed in
the "Remaining findings" column until one of:

- it is fixed (move it to the resolved-history section per rule 11, noting the commit or
  task that fixed it);
- it is explicitly waived by the user or a later audit, with the waiver reason recorded.

Never let a MINOR finding silently disappear from the tracker without one of those two
resolutions being recorded.

## 10. Deferred architecture uses `[-]`, never `[x]`

A concept that PEOS defines but that this SDK deliberately does not implement is `[-]`
Deferred by architecture — never `[x]`, regardless of how well-justified or permanent the
deferral is. Every `[-]` entry in the Deferred Architecture section must state: the
missing concept, its normative source (spec + clause), the reason for deferral, the
ownership/dependency question that must be resolved, and the future packet responsible
(or "not yet named" if none has been scoped).

## 11. Never silently remove history

Do not delete historical packets, findings, or deferrals from the tracker. When a finding
is resolved or a deferral is lifted:

- move it out of the active table/section into a concise "Resolved" note attached to the
  relevant packet or deferred-item entry (a one- or two-line note is sufficient — this is
  not an append-only changelog, it is a compacted history);
- keep enough detail that a future reader can tell what was found and what resolved it,
  without needing to dig through git history or old chat transcripts.

## 12. Every final PEOS report must include a progress update

Every final Claude Code report for a PEOS task — architecture, implementation, audit, or
corrective packet — must end with a progress update block in this exact shape:

```
Progress update:
- Previous status: <packet/spec state before this task>
- New status: <packet/spec state after this task>
- Tracker updated: yes/no
- Exact next action: <one primary next action, matching docs/implementation-progress.md>
- Recommended model: <model tier>
- Recommended mode: <Plan Mode / direct implementation / etc.>
```

If the tracker was not updated as part of the task (for example, a pure read-only
question that changed no state), state `Tracker updated: no` and explain why no update
was needed — do not omit the block.
