# Security Policy

## Reporting a Vulnerability

Please report suspected security vulnerabilities privately by email:

- **tamish560@users.noreply.github.com**

Do **not** open a public GitHub issue for undisclosed vulnerabilities. Include a clear description, affected version or commit, reproduction steps, impact, and any suggested mitigations if available.

## Supported Versions

| Version | Supported |
| --- | --- |
| Latest release on `main` | Yes |
| Older releases | Best effort only |

## Network Behavior

Relay is primarily a local tool, but it **can** use the network:

- It can listen for inbound HTTP connections when started with `relay start --http` or `relay --http`, using a configurable address such as `--addr :8080`.
- It makes outbound API requests required by configured features, including Anthropic integration.
- `tools/web_tools.go` can fetch arbitrary user-supplied URLs.
- `internal/search/ddg.go` sends search requests to DuckDuckGo.

Relay does not include telemetry or analytics, but network exposure depends on how you run and configure it. If you enable HTTP mode, you are responsible for binding to an appropriate interface and placing it behind any needed network controls.

## What Is In Scope

The following are generally considered security vulnerabilities and should be reported:

- Remote code execution or privilege escalation
- Unauthorized access to local files, secrets, or environment variables
- Insecure exposure or bypass of protections when Relay is run in HTTP mode
- Server-side request forgery, unsafe outbound request handling, or unexpected access to internal services
- Authentication, authorization, licensing, or update-path flaws that allow unauthorized use or access
- Dependency vulnerabilities with a realistic impact on Relay deployments

The following are usually **not** treated as security vulnerabilities by themselves:

- Requests to external services that are explicitly triggered by intended product features
- Issues that require a user to knowingly run Relay in HTTP mode on an untrusted network without appropriate safeguards
- General feature requests, availability problems in third-party services, or non-security bugs

We will review reports in good faith and prioritize fixes based on severity, exploitability, and affected usage.
