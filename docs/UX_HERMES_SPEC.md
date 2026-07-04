# Hermes UX Spec

## Goal

Hermes should let a Relay user connect an AI agent to a messaging app in under 2 minutes, with the fewest possible manual steps, no webhook setup for MVP, and clear terminal guidance at every step.

## Product Principles

1. **CLI owns the complexity**: `relay hermes <platform>` should handle validation, local persistence, connection checks, and retries.
2. **No public URL for MVP**: prefer polling, gateway connections, or Web-style session bridges over webhook flows.
3. **One credential max**: each platform flow should ask for at most one secret, or a QR scan.
4. **Show success fast**: users should see "connected" before learning advanced options.
5. **Recover automatically**: after restart, Hermes should reconnect from saved state without rerunning setup.

---

## Recommended command surface

```text
relay hermes                 # interactive platform picker
relay hermes discord         # guided Discord setup
relay hermes telegram        # guided Telegram setup
relay hermes whatsapp        # QR-based WhatsApp setup
relay hermes status          # show all configured channels
relay hermes disconnect      # picker for disconnect/removal
relay hermes logs            # recent channel activity/errors
```

---

## Universal flow

### `relay hermes`

```text
$ relay hermes

Hermes connects Relay agents to messaging apps.
Fastest setup:
  • Discord   ~1 min
  • Telegram  ~1 min
  • WhatsApp  ~2 min

? Pick a platform:
  > Discord
    Telegram
    WhatsApp
    View connected channels
    Exit
```

After selection, Hermes launches the platform-specific guided flow.

### Returning user flow

```text
$ relay hermes

Connected channels:
  • Discord   relay-bot in 2 servers   healthy
  • Telegram  @relay_example_bot       healthy
  • WhatsApp  +91 ••••••4821           reconnecting

? What do you want to do?
  > Add another channel
    Reconnect a channel
    Remove a channel
    View status
    Exit
```

### `relay hermes status`

```text
$ relay hermes status

Hermes channels

Discord
  Status: connected
  Bot: relay-bot
  Mode: gateway
  Last event: 8s ago

Telegram
  Status: connected
  Bot: @relay_example_bot
  Mode: long polling
  Last event: 22s ago

WhatsApp
  Status: needs re-auth
  Account: +91 ••••••4821
  Reason: session expired on this machine
  Fix: relay hermes whatsapp
```

### Universal decision tree

```text
Start
 └─ run `relay hermes`
    ├─ no channels configured
    │  └─ show platform picker
    └─ existing channels found
       ├─ healthy → show status + add/manage options
       ├─ degraded → show one-line fix CTA
       └─ missing secrets/session → route to reconnect flow
```

### Shared UX behaviors

- Remember last selected platform.
- Pre-fill known config where possible.
- Validate secrets immediately after paste.
- Never echo tokens in plain text after submission.
- Always end with a "Try this now" example.
- Offer copy-paste safe commands when browser steps are needed.

---

## 1) Discord UX spec

## Recommendation

For MVP, support **user-created bot token flow** first. A pre-made shared Relay bot is possible, but is worse for privacy, permissions, rate-limit isolation, and multi-tenant trust. Long term, offer:

1. **Default**: "Create your own bot (recommended)"
2. **Optional later**: "Invite the hosted Relay bot"

Hermes should automate everything except the final Discord developer portal actions.

## Minimum info needed

- **Required**: bot token
- **Helpful but optional**: server/guild choice happens in Discord invite UI
- **No webhook URL needed**
- **No client secret needed**

## Technical approach

- Connect via Discord Gateway, not webhooks.
- Open browser automatically to:
  - Discord Developer Portal applications page
  - OAuth2 URL generator or prebuilt invite URL
- After token paste, Hermes:
  - validates token format
  - calls `GET /users/@me`
  - checks needed gateway intents/config guidance
  - generates invite URL with minimum required scopes/permissions

## Exact terminal flow

```text
$ relay hermes discord

Discord setup takes about 1 minute.
You'll create a bot, paste one token, then invite it to your server.

? How do you want to connect?
  > Create my own Discord bot
    Use Relay hosted bot (coming later)

Step 1 of 3 — Create bot
Opening Discord Developer Portal...

We opened:
  https://discord.com/developers/applications

Do this:
  1. New Application
  2. Give it any name
  3. Go to Bot
  4. Click "Reset Token" or "Copy Token"

? Paste your bot token:
> ************************************************

Validating token... done
Bot identity: relay-helper#1842

Step 2 of 3 — Enable required settings
Make sure these are ON in the Bot page:
  [x] MESSAGE CONTENT INTENT
  [x] SERVER MEMBERS INTENT (only if mention routing needs member lookup)

Press Enter once enabled.

Step 3 of 3 — Invite bot
Opening invite URL...

Invite URL scopes:
  bot applications.commands

Required permissions:
  View Channels
  Send Messages
  Read Message History
  Attach Files
  Use Slash Commands

? Press Enter after the bot has joined your server.

Testing gateway connection... done
Testing message send permissions... done

Your agents are now available in Discord.
Try: @relay resize my-image.png to 800px wide
```

## Discord decision tree

```text
User selects Discord
 ├─ choose "create my own bot"
 │  ├─ open developer portal
 │  ├─ user pastes token
 │  │  ├─ invalid format → explain token shape, reprompt
 │  │  ├─ API auth fails → ask for fresh token
 │  │  └─ valid → continue
 │  ├─ show required intents
 │  ├─ open invite URL
 │  ├─ verify bot joined at least one guild
 │  └─ connect gateway + save config
 └─ choose "hosted bot" (future)
    ├─ open OAuth invite
    ├─ user authorizes Relay bot
    └─ map guild/channel access in Relay account
```

## Discord browser automation

Hermes should automatically open:

1. `https://discord.com/developers/applications`
2. After token validation, an invite URL like:

```text
https://discord.com/api/oauth2/authorize?client_id=<app_id>&permissions=<minimal_perm_int>&scope=bot%20applications.commands
```

If browser open fails:

```text
Couldn't open your browser automatically.
Open this URL manually:
https://discord.com/developers/applications
```

## Discord error handling

### Token paste errors

- **Empty input**
  - "Paste the bot token from the Bot tab."
- **Malformed token**
  - "That doesn't look like a Discord bot token."
- **Unauthorized**
  - "Discord rejected that token. Try 'Reset Token' in the Bot page and paste the new one."

### Permission/setup errors

- **Bot not in any server**
  - "The bot isn't in a server yet. Finish the invite step, then press Enter."
- **Missing intent**
  - "Message content intent is off. Turn it on in Bot > Privileged Gateway Intents."
- **Missing send permission**
  - "The bot connected, but can't post in the selected channel. Re-invite with the suggested permissions."

### Runtime errors

- **Gateway disconnect**
  - auto-reconnect with exponential backoff
  - status shows "reconnecting"
- **429 rate limit**
  - obey Discord bucket headers automatically
  - surface only if messages are delayed significantly

## Discord MVP implementation notes

- Ask only for token in CLI.
- Do not require manual app ID entry; derive it from `GET /users/@me` / application lookup.
- Use mentions as first interaction model; slash commands can be added after baseline flow works.

## Recommended Go libraries

- **Gateway/REST**: `github.com/bwmarrin/discordgo`
- Why:
  - mature
  - supports gateway + REST
  - common choice for bots
  - built-in rate limit handling

---

## 2) Telegram UX spec

## Recommendation

Telegram should use the BotFather token flow plus **long polling** for MVP. This is already close to ideal and avoids webhook/public URL complexity completely.

## Minimum info needed

- **Required**: bot token from BotFather
- No webhook URL
- No app registration in a developer portal

## Exact terminal flow

```text
$ relay hermes telegram

Telegram setup takes about 1 minute.
You'll create a bot with BotFather, paste the token, and Hermes will start listening.

Step 1 of 2 — Create bot
Opening BotFather...

If Telegram opens:
  1. Send /newbot
  2. Pick a display name
  3. Pick a unique username ending in 'bot'

If it didn't open, message @BotFather manually.

? Paste your Telegram bot token:
> ************************************************

Validating token... done
Bot username: @relay_example_bot

Step 2 of 2 — Start listening
Starting long polling... done

Almost there:
  1. Open your bot in Telegram
  2. Press Start
  3. Send a test message

Waiting for first message... received

Your agents are now available in Telegram.
Try: summarize the PDF I just uploaded
```

## Telegram decision tree

```text
User selects Telegram
 ├─ open BotFather link
 ├─ user creates bot
 ├─ paste token
 │  ├─ invalid format → reprompt
 │  ├─ getMe fails → ask for correct token
 │  └─ success → store token
 ├─ start polling
 ├─ wait for /start or first inbound message
 └─ mark channel healthy
```

## Token validation

Detect likely valid token before network call:

- Pattern: `<digits>:<token>`
- digits section must be numeric
- suffix length must exceed a minimum threshold

Then confirm via `getMe`.

If valid:

- Show bot username and display name
- Never print full token back

## Telegram browser/app automation

Open:

```text
https://t.me/BotFather
```

Optional convenience:

```text
https://t.me/<bot_username>
```

after validation, so user can immediately hit Start.

## Telegram error handling

### Setup errors

- **Malformed token**
  - "Telegram bot tokens look like `123456789:ABC...`."
- **Unauthorized**
  - "Telegram rejected that token. Copy it again from BotFather."
- **Bot never started by user**
  - "Your bot is connected, but Telegram won't deliver chat messages until you press Start in the bot chat."

### Runtime errors

- **Polling conflict**
  - "Another process is already polling this bot. Stop the other instance or disable its webhook."
- **429 flood control**
  - queue outbound messages and retry using Telegram's `retry_after`
- **Large file unsupported**
  - "That file is too large for the current Telegram flow."

## Telegram MVP implementation notes

- Start polling immediately after token validation.
- Wait up to 60 seconds for the first inbound message; if none arrives, still save config and print next steps.
- Support direct messages first; group support can come later.

## Recommended Go libraries

- **Primary**: `github.com/go-telegram-bot-api/telegram-bot-api/v5`
- Alternative: `gopkg.in/telebot.v3`
- Recommendation: use `telegram-bot-api/v5` for straightforward polling, API coverage, and adoption.

---

## 3) WhatsApp UX spec

## Recommendation

For MVP, use a **WhatsApp Web multi-device bridge** with QR login. This is the only path that matches the target UX. Avoid WhatsApp Business Cloud API for MVP because Meta app setup, business verification, and webhook hosting destroy the setup speed goal.

## Minimum info needed

- **Required**: none typed
- **Required action**: scan QR with WhatsApp mobile app
- No tokens
- No webhooks
- No developer portal

## Technical approach

- Use WhatsApp Web multi-device protocol
- Store encrypted session credentials locally after first QR scan
- Reconnect automatically on restart

## Feasible Go approach

Use:

- `go.mau.fi/whatsmeow`

Why:

- actively used Go WhatsApp Web client
- supports QR login flow
- persistent device store support
- suited for local-first architecture

For persistence, pair with:

- `go.mau.fi/whatsmeow/store/sqlstore`
- local SQLite DB

## Exact terminal flow

```text
$ relay hermes whatsapp

WhatsApp setup takes about 2 minutes.
You'll scan a QR code from your phone. No API keys needed.

Starting secure local session... done

Step 1 of 2 — Scan QR
Open WhatsApp on your phone:
  Settings > Linked Devices > Link a Device

Scan this QR code:

████ █ ▄▄ ▄█ ...
...terminal QR...

Waiting for scan... scanned
Waiting for device confirmation... confirmed

Step 2 of 2 — Finish sync
Syncing account metadata... done
Connected account: +91 ••••••4821

Your agents are now available in WhatsApp.
Send a message to your linked account thread to begin.
```

## WhatsApp decision tree

```text
User selects WhatsApp
 ├─ existing valid session?
 │  ├─ yes → reconnect silently
 │  └─ no → start QR pairing
 ├─ render QR in terminal
 │  ├─ QR expires → refresh QR automatically
 │  └─ scan succeeds → continue
 ├─ wait for device confirmation
 ├─ persist session
 └─ mark channel healthy
```

## QR UX details

- Render QR directly in terminal using block characters.
- Also offer:
  - "Press `c` to copy pairing payload"
  - "Press `r` to refresh QR" if interactive input is supported
- If terminal QR rendering is poor, write a PNG into the project directory only if explicitly requested; default is terminal-only.

## WhatsApp error handling

### Setup errors

- **QR expired**
  - auto-refresh and redraw
  - "QR expired. Showing a new one..."
- **Phone offline**
  - "Your phone needs internet access to finish linking."
- **Multi-device unavailable**
  - "This WhatsApp account can't link right now. Update WhatsApp and try again."

### Session errors

- **Logged out from phone**
  - "WhatsApp disconnected this linked device. Run `relay hermes whatsapp` to relink."
- **Session store corrupt**
  - offer safe reset:
    - "We couldn't restore your WhatsApp session. Remove local session and pair again?"

### Runtime errors

- **Reconnect loop**
  - exponential backoff with clear status output
- **Media upload failure**
  - "Text messaging works, but media upload failed. Check network and retry."

## WhatsApp MVP implementation notes

- Support 1 linked WhatsApp account first.
- Prefer personal chat/thread model initially over group routing.
- Defer advanced business features, templates, or phone-number ownership workflows.

## Recommended Go libraries

- **WhatsApp Web client**: `go.mau.fi/whatsmeow`
- **QR rendering**: `github.com/mdp/qrterminal/v3` or `github.com/skip2/go-qrcode` for PNG fallback
- **SQLite driver**: whichever best matches Relay's dependency policy; if pure Go is preferred, evaluate `modernc.org/sqlite`

---

## 4) Platform-specific setup guides

## Discord quick guide

1. Run `relay hermes discord`
2. Hermes opens Discord Developer Portal
3. User creates app + bot
4. User pastes token
5. Hermes validates token
6. Hermes opens invite URL
7. User adds bot to a server
8. Hermes verifies gateway connection

## Telegram quick guide

1. Run `relay hermes telegram`
2. Hermes opens BotFather
3. User runs `/newbot`
4. User pastes token
5. Hermes validates with `getMe`
6. Hermes starts long polling
7. User presses Start in bot chat
8. Hermes confirms first message

## WhatsApp quick guide

1. Run `relay hermes whatsapp`
2. Hermes starts local WhatsApp Web session
3. Hermes prints QR in terminal
4. User scans with Linked Devices
5. Hermes persists session
6. Hermes reconnects automatically in future

---

## 5) Dream UX details

Target baseline:

```text
$ relay hermes
? Pick a platform: Discord
? Paste your bot token: ****
  Connecting... done!

  Your agents are now available in Discord.
  Try: @relay resize my-image.png to 800px wide
```

To get closest per platform:

- **Discord**: one secret paste + browser-opened invite flow
- **Telegram**: one secret paste + polling
- **WhatsApp**: no secret paste, QR only

Every success screen should include:

- platform name
- connected identity
- what to do next
- one realistic prompt example
- one recovery command

Success template:

```text
Connected: <platform identity>
Mode: <gateway/polling/session>
Persistence: saved on this machine

Try this now:
  <example message>

Manage later:
  relay hermes status
```

---

## 6) Security model

## Secrets and session storage

### Preferred storage model

Store credentials in OS-native secure storage whenever possible:

- **Windows**: Credential Manager / DPAPI-backed secret storage
- **macOS**: Keychain
- **Linux**: Secret Service / libsecret, with encrypted-file fallback if unavailable

Recommended Go package direction:

- `github.com/zalando/go-keyring` or similar cross-platform keyring library

### What to store

- **Discord**: bot token
- **Telegram**: bot token
- **WhatsApp**: encrypted local session/device credentials

### Non-secret config on disk

Safe to store in a local config file:

- platform enabled/disabled
- display names/usernames
- guild/chat metadata
- last connected time
- reconnect policy

Suggested path:

```text
~/.relay/hermes/config.json
```

with secrets stored separately in keychain.

If no keychain is available:

- fall back to encrypted local file
- encryption key derived from OS-protected mechanism where possible
- warn clearly:
  - "Secure keychain unavailable; using encrypted local storage on this machine."

## Session persistence

On process start:

1. load saved channel configs
2. restore secrets/sessions
3. reconnect each platform in background
4. mark each channel `connected`, `reconnecting`, or `needs_auth`

If machine restarts:

- **Discord**: reconnect using saved token
- **Telegram**: restart polling using saved token
- **WhatsApp**: restore stored device session; relink only if server invalidates it

## Secret handling rules

- Never print full token after paste.
- Support masked input for tokens.
- Redact tokens from logs and panic traces.
- Do not store secrets in shell history, env files, or project config by default.
- Copy/paste prompts should mention "input is hidden."

## Revocation/removal UX

```text
$ relay hermes disconnect

? Remove which channel?
  > Telegram @relay_example_bot

This will:
  • delete the saved token from secure storage
  • stop background reconnects
  • keep message history/log metadata only if you choose

? Also delete local metadata? (y/N)
```

## Logging model

Logs must exclude:

- full tokens
- WhatsApp session blobs
- raw user message bodies if privacy mode is enabled

Logs may include:

- connection state transitions
- rate-limit waits
- error categories
- non-sensitive platform IDs

---

## Rate limits and resilience

## Discord

- rely on library rate limit handling
- batch outgoing sends if agent produces bursts
- show human-facing warning only for sustained delay

## Telegram

- respect `retry_after`
- queue sends per chat
- cap retries to avoid infinite loops

## WhatsApp

- pace sends conservatively
- avoid sudden high-volume outbound automation
- surface "temporary send delay" rather than raw transport errors

## Cross-platform resilience patterns

- exponential backoff with jitter
- idempotent reconnect
- health state machine:
  - `connected`
  - `connecting`
  - `reconnecting`
  - `needs_auth`
  - `error`
- `relay hermes status` always shows current state + next action

---

## Suggested implementation order

### Phase 1

1. Telegram
2. Discord
3. `relay hermes` picker/status shell

Reason: fastest MVP with lowest technical risk.

### Phase 2

1. WhatsApp QR flow via `whatsmeow`
2. secure storage hardening
3. reconnect + status polish

### Phase 3

1. hosted/shared Discord bot option
2. richer routing controls
3. multi-channel per platform

---

## Final recommendations

### Discord

- MVP should be **bring-your-own bot token**
- Open browser automatically
- Require only token paste
- Use gateway, not webhooks

### Telegram

- MVP should be **BotFather + token paste**
- Validate token instantly
- Use long polling

### WhatsApp

- MVP should be **QR scan via WhatsApp Web bridge**
- Use `whatsmeow`
- No Meta Business API for initial release

### Universal UX

- `relay hermes` should act like a setup wizard
- `relay hermes status` should act like an always-readable health dashboard
- The CLI should remember everything it safely can and ask users for as little as possible
