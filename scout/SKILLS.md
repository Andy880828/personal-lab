---
name: scout
description: Personal tech-news radar operated by you (the agent). Fetches the user's own sources (HN, tech media, AI labs, people they follow), dedupes and ranks them with explainable rules; you pick 4-7 items, write one-sentence Traditional Chinese summaries, save the daily digest and present it in the conversation. Also answers "what's heating up lately?" via term trends. Use when the user asks for today's/this week's tech news, a briefing, what they missed, what is trending, or to tune their news sources.
---

# scout — Agent Operating Manual

`scout` is one static binary with a local JSON store. It fetches, dedupes,
ranks and stores. It never calls an LLM and never contacts the user: **you**
judge, summarise and report back in the conversation. Telegram or other
channels are deliberately out of scope until the user asks for them.

## 0. Ground rules

1. **Start with `scout status --json`.** It is offline and instant, and its
   `next_steps` is a sensible default plan.
2. **Always pass `--json`** (or export `SCOUT_OUTPUT=json`). Check `ok`
   first; on failure act on `error.code` and `error.hint`.
3. **Relay warnings that matter.** A failing source, truncated text or an
   empty extraction changes what you can honestly claim. Say so.
4. **Summaries come only from text you actually have**: the `read` output, or
   the candidate `excerpt` when reading failed. Never fill gaps from memory or
   guess what an article says from its title. If you could not read it, say
   "（僅依標題與摘要）" or skip it.
5. **Write for the user**: Traditional Chinese (台灣用語), keep product/person
   names in their original form, one sentence per item.
6. **Watchlist changes are the user's decision.** Propose sources/keywords;
   run `sources add|rm|disable` or `keywords add|rm` only after they agree.
   `digest create --replace` and `prune` (without `--dry-run`) also need
   explicit consent.

## 1. Envelope, errors, exit codes

```json
{"ok": true,  "command": "candidates", "schema": "scout/v1", "data": {...}, "warnings": ["..."]}
{"ok": false, "command": "digest create", "schema": "scout/v1", "error": {"code": "not_found", "message": "...", "hint": "...", "details": {"problems": [...], "unknown_ids": [...]}}}
```

| `error.code` | What to do |
|---|---|
| `usage` | Fix flags/args/JSON per `hint`; `scout <cmd> --help --json` gives the schema |
| `no_data` | Run the hinted command (`init`, `fetch`, or create a digest) |
| `not_found` | Wrong id. For digests, `details.unknown_ids` lists them; take ids from `candidates` |
| `conflict` | Already exists. For a same-day digest ask the user before `--replace` |
| `locked` | Another scout run is writing; wait a minute and retry once |
| `network` | Remote unreachable. Offline commands still work; tell the user |
| `unsupported_content` | `read` hit a PDF/video; summarise from the excerpt and say so, or skip |
| `runtime` | Report the message |

Exit codes: `0` ok · `1` runtime error · `2` usage/validation error.
Self-discovery: `scout describe --json` (all commands, flags, error codes,
ranking rules, limits).

## 2. Intent → command

| User intent | Command(s) |
|---|---|
| Anything, first call | `scout status --json` |
| "今天有什麼科技新聞？" / daily briefing | §5 workflow |
| "我錯過了什麼？" (several days) | `scout fetch --json`, then `scout candidates --since 3d --limit 40 --json`, pick up to 7 |
| "最近什麼在升溫？" | §6 workflow |
| "Show me yesterday's / a past digest" | `scout digest list --json`, `scout digest show <date> --json` |
| "Tell me more about #3" | `scout read <id> --json`, answer from the text |
| "Too much X" / "more Y" | propose `scout keywords add --mute X` / `--boost Y`, apply after consent |
| "Follow this blog/person" | §7 adding a source |
| A source keeps failing | `scout sources list --json`, then `scout sources test --url ... --json` |
| Publish the reading page | `scout digest export --out <dir> --json` (served by Caddy, see homelab) |

## 3. Reading results

**fetch** — per source `status`: `ok`, `not_modified` (HTTP 304, healthy,
nothing new) or `error` (counted in `consecutive_failures`). `old_skipped`
counts entries published more than 14 days ago that were not stored (archive
feeds list hundreds). `skipped`/`invalid` count entries without usable links.

**candidates** — the unpicked items in the current window, best first.
- `window_basis`: `last_digest` (continues exactly where the previous digest
  ended — no repeats, no gaps), `default_24h` (no digest yet) or `since_flag`.
- `score` is a default ordering, not a verdict. `reasons` explains every
  point: source weight, followed author, coverage by several sources, HN
  points/comments, boost keywords, recency decay. Quote reasons when the user
  asks why something was picked.
- `sources` / `source_ids`: everyone who listed it; more than one source is a
  real signal. `discussion_url` is usually the HN thread.
- `excerpt` is the feed's own blurb (≤300 chars here), good for triage only.
- `muted_hidden` / `picked_hidden` say how many rows were filtered out; use
  `--include-muted` / `--include-picked` to see them.

**read** — `text` is the main content (prefers `<article>`, then `<main>`).
`container` = `body`/`document` means navigation may be mixed in;
`truncated` + `total_chars` tell you if you saw everything (raise
`--max-chars` if the cut part matters). A `total_chars` under ~300 usually
means a JavaScript-only page or bot blocking.

## 4. Picking and writing (editorial judgment)

Pick **4–7** items (hard limit 12). Aim for what the user would regret
missing in a week, not what is merely popular today.

- Start from the top ~30 candidates; do not just take the first 5.
- **Diversity**: at most two items on the same story or theme; merge
  duplicates (same news, different URLs) into the one primary source.
- **Prefer primary sources** (the lab's announcement, the author's post) over
  rewrites; mention the HN discussion if the debate itself is the news.
- Followed authors and boost keywords reflect stated interests — weigh them,
  but a genuinely important general story still beats a minor on-topic one.
- Skip pure clickbait, listicles and items whose text you could not obtain
  unless the title alone is clearly important (then mark it).
- `read` each pick whose excerpt does not let you write a precise sentence.

**Summary**: one sentence, ideally ≤ 60 字 (hard limit 200 chars). Say who
did what and why it matters; concrete numbers beat adjectives; no hype words
(「震撼」「顛覆」), no "本文介紹". Optional `why`: a few words on why it made
the cut (shown nowhere by default, kept for the record).

## 5. Daily briefing workflow

```bash
scout status --json
scout fetch --json                       # skip if last_fetch.age_hours < 1
scout candidates --limit 30 --json
scout read <id> --json                   # for each pick that needs it
cat > /tmp/picks.json <<'EOF'
{"items": [
  {"id": "b0ecb253f4a7", "summary": "Karpathy 以 HN 歷年留言推薦的 6,454 支影片做出無演算法的 YouTube 式瀏覽介面。", "why": "followed + 1.5k HN points"},
  {"id": "f0cef6501536", "summary": "研究者揭露可接管 ChatGPT 帳號的漏洞鏈與修補時程。"}
], "note": "今天 AI 安全議題偏多"}
EOF
scout digest create --input /tmp/picks.json --dry-run --json   # validate first
scout digest create --input /tmp/picks.json --json
```

Then show `data.markdown` to the user as-is (it has titles, links, sources).
Add one or two lines of your own only if there is a real cross-item insight.
If a static site is served, refresh it with `scout digest export --out <dir> --json`.

Validation errors list **every** problem at once (`details.problems`); fix
them all and resubmit. A digest already existing for today means a briefing
was done — show it (`digest show --json`) unless the user wants a redo.

## 6. "What's heating up" workflow

```bash
scout trends --days 7 --json
```

- `rising` = terms whose article count grew vs the previous 7 days;
  `new: true` = absent before. `top` = most mentioned overall.
- Terms are only word counts from titles. Interpret them: group related
  terms, open `examples` (item ids) or `scout candidates --since 7d
  --include-picked --limit 60 --json` to see what the articles are, and drop
  noise words.
- Report 3–5 themes, each with one sentence and 1–2 representative links.
- Heed warnings: with no previous-window data everything looks "new"; CJK
  titles are not tokenised (check them via candidates).
- If a theme matters to the user but their sources barely cover it, propose a
  source or boost keyword (§7).

## 7. Maintaining the watchlist

- `scout sources list --json` shows each source with health and how many
  stored items it contributed.
- Adding: find the feed URL (try `/feed`, `/rss`, `/atom.xml`, `/index.xml`,
  or the page's `<link rel="alternate" type="application/rss+xml">`), then
  `scout sources test --url <url> --json`, show the sample titles to the
  user, and after consent
  `scout sources add --id <id> --name "<name>" --url <url> [--weight 1.5] [--followed] --json`.
  `add` refuses sources that fail a live fetch.
- Weights: 1 normal, 1.5 strong, 0.5 weak. `--followed` is for people.
- A source failing 3+ times in a row: test it; if the feed moved, propose
  replacing it; if it is just down, leave it.
- The watchlist lives at `$SCOUT_HOME/watchlist.json` (default `~/.scout`);
  suggest the user keep it in git.

## 8. Scheduling

`fetch` is safe to run unattended (cron/systemd timer, e.g. every 3 hours);
it only needs the binary. Picking and summarising needs you, so the daily
digest runs when the user talks to you or when their own scheduled session
asks you to. Never set up schedules without being asked.

## 9. Cost model

- `fetch`: ~1–2 s for 15 sources (parallel, conditional GET; unchanged feeds
  return 304).
- `read`: one HTTP request per article; read only what you will summarise.
- Everything else is offline and instant.
- Stored items older than 60 days are only useful for trends; `status`
  suggests `prune --dry-run` when there are some. Digests are never pruned.

## 10. Don'ts

- Don't parse human output or edit `state.json` by hand.
- Don't present `score` as importance to the user; explain with `reasons`.
- Don't summarise articles you could not read as if you had.
- Don't change sources/keywords, replace digests or prune without consent.
- Don't add a notification channel on your own; reporting happens in the
  conversation.
