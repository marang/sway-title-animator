# Sway Title Animator workflow conventions

This document defines the normal Linear, branch, review, release, and cleanup
loop for this repository. `AGENTS.md` owns non-negotiable engineering rules;
plans such as `docs/sound-presets-plan.md` own lasting technical decisions.

## Issue routing

Normal animator work uses the Linear `Lab` team, the Sway Title Animator P001
project, and the mutually exclusive `Codebase` → `Sway Title Animator` label.
One executable, reviewable slice is in progress at a time. Branch names include
their `LAB-*` key.

LAB-119 is a documented exception: it coordinates the repository split under
the new Sway Session project because its acceptance criteria cover both
repositories. Cross-repository cutover notes, package ownership, and upgrade
instructions must be synchronized in the two projects before release.

## Delivery loop

`Linear issue → In Progress → current main → focused branch → implementation
and tests → local review → PR → CI → merge → Done`

Start from current `main`, inspect the worktree, preserve unrelated changes,
and make a focused branch. Run relevant focused checks while iterating and
`make verify` before handoff. Run the `code-review` skill before a PR; run
`review-codebase` at the stated project cadence. PR descriptions include user
behavior, compatibility, fallback, verification evidence, and the Linear issue.

Merge only when the diff still matches the issue, local verification and CI are
green, reviews are complete, public documentation is synchronized, and
follow-ups are explicitly bounded. Then move the issue to Done, fast-forward
local `main`, and clean up the branch.

## Split-release coordination

The v0.10.0 split is a paired release, not an automatic replacement. Releases
at or below v0.9.3 package both binaries. The documented upgrade is animator
first, then session without restarting Sway, or an explicit `pacman -U` of both
new package files; do not promise that a helper's update command is atomic and
never require `--overwrite`. Do not declare a blanket package `replaces`
relationship that could uninstall the animator while installing the session
package. Verify binary and Sway-template ownership in both release artifacts
before publishing either one.
