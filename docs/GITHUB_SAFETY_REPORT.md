# GitHub Safety Report

Last updated: 2026-07-04
Account: `tamish560`
Org: `valtors`
Repos: `valtors/relay`, `valtors/relay-landing`

## Executive summary

Status for today: **WARNING**

Reason:

- Hard API rate limits are fine. Core API has **4995/5000** requests left this hour.
- Behavioral risk is the real problem on a new account.
- Today already includes **2 issues**, **about 6 commits**, **1 new repo**, profile edits, and repo metadata changes.
- That is not extreme, but it is enough that more public activity today starts looking noisy.

Bottom line: **stop non-essential GitHub activity for the rest of today**. If something must ship, do **one final commit max**, no new issues unless truly needed, and no discussions.

## 1. Today's activity audit

### Hard limit audit

| Limit | Current state | Risk |
|---|---:|---|
| Core REST API | 4995 / 5000 left | Safe |
| Search API | 30 / 30 left | Safe |
| GraphQL | 5000 / 5000 left | Safe |

### Behavioral audit

| Activity | Today | Risk |
|---|---:|---|
| Issues created | 2 | Moderate |
| PRs opened | 0 mentioned | Safe |
| Commits pushed | about 6 | Moderate |
| Repos created | 1 | Moderate |
| Discussions | not mentioned | Safe if kept at 0 today |
| Profile / metadata edits | yes | Low alone, higher in combination |

### Verdict

This is **not near a hard rate limit**, but it is **close to the safe behavioral ceiling for one day on a new account**. The mix of repo creation, multiple pushes, and issue creation is the part that matters.

## 2. Remaining budget for today

Use the stricter budget, not the theoretical maximum.

| Item | Guideline | Done today | Safe remaining today | Recommendation |
|---|---:|---:|---:|---|
| Issues | 3 / day | 2 | **1 max** | Only if important, wait 20 to 30 min |
| PRs | 2 / day | 0 | **0 to 1** | Prefer 0 today |
| Commits | 5 to 6 / session | about 6 | **0 preferred, 1 max** | Only a critical fix |
| Discussions | keep very low | 0 | **0** | Avoid today |

### Should we stop pushing today?

**Yes, for non-critical work.**

Safe rule:

- Best move: no more pushes today
- If required: 1 final commit, then stop
- Do not open a PR after that unless it is necessary

## 3. Launch plan rules

### Discussions

Risk: high on a new account if posts are frequent, polished, or repetitive.

Rules:

- Max **1 discussion in a day**
- Prefer **0 discussions** on days with issues or PRs
- Wait **30 to 60 minutes** after any other public GitHub post
- Keep it short, plain, and specific
- Do not seed multiple categories in one sitting

### Good first issue creation

Safe, but only if done slowly.

Rules:

- Create **1 issue at a time**
- Max **1 new good first issue per day**
- Make each one concrete and code-specific
- Do not open a batch of similar onboarding issues
- Add labels after creation, not through a fast script loop

### Pull requests

Risk: medium. PRs are public and easy to flag if they look machine-made.

Rules:

- Max **1 PR per day** for the first week
- Keep PR descriptions under **150 words**
- If several small fixes exist, combine them into one PR
- Do not open and close PRs repeatedly
- Avoid self-review spam, force-push loops, and many follow-up comments

### Social posts linking to the repo

Usually safe. The risk comes from synchronized spammy behavior.

Rules:

- Fine to post socially, but do not immediately create several GitHub threads after
- Do not paste the same promo text everywhere
- If social traffic lands on the repo, keep GitHub posting calm that day

## 4. Content review checklist

Copy-paste this before posting anything on GitHub:

```md
- [ ] No em dashes, use hyphens or commas instead
- [ ] No corporate buzzwords like leverage, synergy, cutting-edge
- [ ] Reads like a real developer wrote it casually
- [ ] Issues stay under 200 words
- [ ] PR descriptions stay under 150 words
- [ ] No wall of bullets, keep lists to 5 items max
- [ ] Commit messages are lowercase, short, and casual
- [ ] No excessive formatting, bold spam, or emoji spam
- [ ] Mentions real code, files, or behavior when possible
- [ ] Not copy-pasted from another issue, PR, or discussion
```

### Safe writing examples

| Bad | Better |
|---|---|
| "This pull request implements a comprehensive improvement to the pipeline." | "tighten pipeline error handling" |
| "I would like to start a discussion regarding potential collaboration patterns." | "question about how we want to handle tool retries" |
| "Leverage this issue to improve contributor onboarding." | "add a small starter issue for the CLI help text" |

## 5. Safe 7-day activity schedule

This schedule is intentionally conservative.

| Day | Issues | PRs | Commits | Discussions | Notes |
|---|---:|---:|---:|---:|---|
| Day 1 | 0 to 1 | 0 | 0 to 1 | 0 | Today. Cool down day. |
| Day 2 | 1 | 0 to 1 | 2 to 3 | 0 | Product work only. |
| Day 3 | 1 | 0 | 2 to 3 | 1 | If posting a discussion, no other public thread bursts. |
| Day 4 | 0 to 1 | 1 | 2 to 3 | 0 | Use for one solid PR if needed. |
| Day 5 | 1 | 0 | 2 to 3 | 0 to 1 | Good day for one contributor-facing item. |
| Day 6 | 0 | 0 to 1 | 1 to 2 | 0 | Light day. |
| Day 7 | 0 to 1 | 0 | 1 to 2 | 0 | Wrap up, no bursts. |

### Weekly ceiling

Do not exceed this next week unless the account has been quiet and healthy:

| Type | Weekly ceiling |
|---|---:|
| Issues | 4 to 5 |
| PRs | 2 to 3 |
| Discussions | 1 to 2 |
| Commits pushed | 14 to 18 |
| New repos | 0 |

## 6. Red flags and how to avoid them

| Red flag | Why it gets flagged | Mitigation |
|---|---|---|
| Many issues, PRs, or discussions in a short window | Looks automated or promotional | Space public posts by 20 to 60 minutes |
| Repetitive wording or structure | Looks AI-generated or templated | Rewrite manually, vary openings and sentence shape |
| Em dashes and polished corporate tone | Common AI signal | Use plain language, short sentences, contractions |
| Many commits pushed fast to main | Looks like bot churn | Work locally, then push a small grouped set |
| Open-close-edit-delete loops | Looks unnatural | Post once when ready |
| Bulk label or metadata changes | Looks scripted | Do small manual edits only |
| Cross-posting the same message | Spam pattern | Tailor each post to its context |
| Ignoring 403 or 429 errors | Escalates abuse systems | Stop immediately and wait |
| Heavy API bursts from scripts | Triggers secondary limits | Keep automation under 1 request per second |
| Repo plus profile plus community activity all at once | New-account trust drops | Split account activity across days |

## 7. Emergency procedures

### If rate-limited

1. Stop all scripts, CI triggers, and manual posting.
2. Check limits with `gh api rate_limit`.
3. Wait until reset for primary limits.
4. If the message mentions abuse detection, wait **at least 30 to 60 minutes**.
5. When resuming, cut activity volume by half for 48 hours.

### If account gets flagged or content gets hidden

1. Stop all GitHub activity for **24 hours**.
2. Do not switch accounts, tokens, or IPs.
3. Do not retry the same post with small edits.
4. Review recent issues, PRs, and discussions for AI-like tone or duplication.
5. If needed, contact GitHub Support calmly with a short factual note.

### If a post must go out during a warning period

- Pick only one: issue, PR, or discussion
- Keep it very short
- Reference real code
- No formatting flair
- No follow-up burst after posting

## 8. Specific rules for the 7-day launch

### Allowed

- Slow, code-backed issues
- One careful PR on a day reserved for PRs
- Social posts that link to the repo
- README and docs improvements pushed in grouped commits

### Not allowed

- More than 1 discussion in a day
- More than 1 PR in a day
- More than 3 issues in a day
- Batch creation of good first issues
- Same-day bursts across issues, PRs, discussions, and profile edits

### Best practice sequence

1. Push code quietly.
2. Wait.
3. Open one issue or one PR, not both back-to-back.
4. Wait again before any other public activity.
5. If doing social promotion, keep GitHub quiet the rest of that block.

## Final recommendation

For **today**, treat the account as **at limit for normal work**:

- **Issues left:** 1 max
- **Safe commits left:** 0 preferred, 1 max
- **PRs left:** 0 preferred
- **Discussions left:** 0

If there is no critical need, **stop pushing for today and resume tomorrow on a lighter schedule**.
