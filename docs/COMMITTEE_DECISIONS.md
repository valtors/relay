# Relay Committee Decisions

Date: 2026-07-04

## Committee Discussion

### PM Agent
Relay already has enough surface area to attract early users. Forty tools, working install scripts, a CLI, and a live landing page are not the main problem right now. The problem is activation. A new user needs to understand in under two minutes what Relay does, why it is different from a plain MCP server, and how to get one useful result on the first try.

The single most impactful next step is not more tools and not orchestration yet. It is packaging the current product into a sharp first-run experience. That means a rewritten README, a fast start path, and a few opinionated example workflows that make Relay feel immediately useful. If a user cannot go from GitHub page to successful local run in five minutes, growth will stall no matter how many tools are added.

The MVP for real usage today is simple:
- install Relay
- start the server
- connect it to one agent client
- run 3 concrete workflows that save time

Recommended workflows:
1. Research and summarize a repo or document set
2. Transform local files and extract structured data
3. Fetch web content and produce a working artifact

What is blocking adoption:
- unclear positioning versus generic MCP servers
- no obvious hero workflow
- likely too much breadth, not enough narrative
- missing proof that first run works cleanly on macOS, Linux, and Windows

### Market Research Agent
First users are not general consumers. They are:
- AI power users already using Claude, Cursor, VS Code, or local MCP clients
- developers searching for open source MCP servers
- builders who want tools plus memory plus coordination in one place
- indie hackers who want less glue code

What they are searching for:
- open source MCP server
- self hosted MCP tools
- MCP server with memory
- MCP server for file tools pdf tools web tools
- alternatives to building custom agent tool backends

Their pain points:
- most MCP servers are narrow and require stitching together many repos
- setup is fragmented
- demos are abstract instead of task-driven
- many projects sound smart but do not show a reliable first run

What Relay solves:
- one install, many tools
- local-first utility
- memory and coordination direction already visible
- practical workflow focus instead of only framework talk

Competitive reality right now:
- the market rewards clarity over completeness
- projects win attention through tight demos, screenshots, and copy-paste setup
- users trust repos that show exact client config, exact commands, and exact outputs

Channels that fit without risking account safety:
- GitHub README and releases
- one polished dev.to or Hashnode post
- Reddit only where there is clear value and direct relevance
- X posts from personal account kept human and low volume
- comments on relevant MCP and agent threads only when genuinely helpful

### GTM Agent
The first 100 stars will not come from broad marketing. They will come from a repo page that converts plus a few authentic distribution touches.

Launch asset priority:
1. README that sells outcomes
2. 60 to 90 second terminal demo GIF
3. three ready-to-copy client configs
4. one launch post with a real story, not hype

The first 10 contributors will come after users can reproduce success. Contributors do not join confusion. They join momentum. That means clear issue labels, a short contributor path, and a visible roadmap that says what is wanted now.

Best launch sequence:
- tighten README and examples
- cut a clean release
- post to GitHub, X, Reddit, and one builder community
- follow up with one technical breakdown post
- invite contributors through small scoped issues

Tone rules:
- casual
- specific
- no inflated claims
- show commands and outputs
- say what works today and what is coming next

### QA Agent
Breadth is enough for now. Depth and trust are the issue.

What would embarrass Relay in production:
- install script works but first client config is confusing
- tool names are inconsistent or hard to discover
- cross-platform startup differs from docs
- docs promise memory and coordination more strongly than current product supports
- smoke tests pass but user journey fails

What is likely untested:
- first-run flow on clean machines
- exact copy-paste config snippets for popular clients
- error messages when a dependency or path is missing
- tool discoverability under real agent use

Quality bar before pushing harder on reach:
- install succeeds on Windows, macOS, Linux
- start command works without surprise
- README quickstart is verified end to end
- at least 3 real workflows are tested exactly as documented
- scope claims match current capabilities

### CEO Synthesis
The committee agrees on a hard truth. Relay does not need more product surface this week. It needs conversion. The biggest gap between the current state and real usage is onboarding clarity, not missing features.

The single most impactful thing to do next is to rewrite the README and quickstart around one sharp activation path with three concrete workflows, verified client configs, and proof that first run works cross-platform.

Why this wins:
- it turns existing product value into user activation immediately
- it improves GitHub page conversion without risky promotion
- it creates assets reusable for posts, demos, and contributor onboarding
- it exposes real product gaps faster than building more tools in the dark

Why not orchestration next:
- it is strategically important but not the fastest path to first users
- it increases scope before the current offer has been packaged well
- early adopters need a reason to try Relay now, not a bigger roadmap

Why not Hermes bridge next:
- messaging adapters are interesting but not core to initial developer adoption
- they add support burden and integration complexity before base activation is solved

Why not more tools next:
- 40 tools is already enough to prove utility
- more tools without stronger positioning likely lowers clarity

## DECISIONS

1. Rewrite the README and quickstart now - onboarding clarity is the main blocker to real usage, not missing features.
2. Sell Relay through workflows, not feature lists - users adopt outcomes they can copy, not tool counts.
3. Define one activation goal - a user should install, connect, and complete one useful task in five minutes.
4. Add three verified starter workflows - they create proof, content, and support a cleaner launch story.
5. Keep orchestration on the roadmap, not in this sprint - it matters long term but does not beat better activation this week.
6. Do not ship broader claims than the product supports today - trust is more valuable than hype for an open source infra tool.
7. Focus distribution on human, low-volume channels - protect account safety and let the repo do most of the conversion work.

## IMMEDIATE PRIORITIES

1. Rewrite README headline, opening section, and quickstart
   - lead with what Relay helps an AI agent get done
   - explain memory, tools, and coordination clearly without overselling coordination
   - add install commands for macOS, Linux, and Windows
   - add exact `relay start` flow and expected output

2. Add a "Start Here" section with 3 concrete workflows
   - repo research workflow
   - local file and PDF extraction workflow
   - web fetch to artifact workflow
   - each should include client config, prompt, and expected result

3. Add verified client setup snippets
   - Claude Desktop
   - Cursor or VS Code MCP config
   - one generic JSON config example

4. Tighten product positioning everywhere
   - replace tool inventory-first language with outcome-first language
   - clarify what works today versus what is planned
   - align landing page copy with README messaging

5. Record one short terminal demo
   - install
   - start
   - run one useful workflow
   - export as GIF for README and launch posts

6. Run first-run smoke validation
   - verify every copied command
   - verify every config snippet
   - verify no dead links or outdated claims

## LAUNCH PLAN

### Day 1
- rewrite README
- finalize quickstart
- add three workflows
- verify all commands locally

### Day 2
- record demo GIF
- update landing page copy to match README
- cut a clean tagged release if needed

### Day 3
- publish a GitHub release note focused on use cases
- post one human launch thread on X with demo GIF
- update GitHub repo description and pinned metadata

### Day 4
- post a detailed write-up on dev.to or Hashnode
- focus on "I built an open source MCP server that already includes memory, file, PDF, data, and web tools"
- include exact setup steps and one real workflow

### Day 5
- share selectively in relevant Reddit and Discord communities
- only post where Relay directly fits the discussion
- respond to questions with concrete help, not promotion

### Day 6
- create 5 to 10 small contributor-friendly issues
- label them clearly
- add expected outcome and file pointers
- invite contributors in the README and release notes

### Day 7
- review activation signals
- count stars, installs, README visits, issues, and setup questions
- note where users got stuck
- decide whether the next move is better docs, better setup, or one missing killer workflow

## QUALITY GATES

1. README quickstart works exactly as written on Windows, macOS, and Linux.
2. At least three workflows are tested end to end and produce the documented outcome.
3. Client config examples are valid JSON and load correctly in the named clients.
4. Install scripts and binary release paths match current release artifacts.
5. Product claims clearly separate current capabilities from backlog items.
6. Demo GIF shows the real product flow with no manual cleanup or hidden steps.
7. A new user can understand what Relay is in under 30 seconds from the README top section.

## RISKS AND MITIGATIONS

1. Risk: Building more features delays real usage
   - Mitigation: freeze net-new feature work for this sprint and ship onboarding improvements first.

2. Risk: Messaging overpromises coordination before it is ready
   - Mitigation: describe coordination as the direction, while centering tools and practical workflows that already work today.

3. Risk: Promotion feels spammy or AI-generated
   - Mitigation: keep posting volume low, write in first person, share real build notes, and engage only where Relay clearly helps.

## Single Most Impactful Next Move

Rewrite the README into a ruthless activation asset with a five-minute quickstart, three real workflows, and verified client configs. That is the highest-leverage move because it converts the product you already built into actual user success this week.
