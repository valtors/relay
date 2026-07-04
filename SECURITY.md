# Security Policy

## Reporting a Vulnerability

If you find a security vulnerability, please report it responsibly.

Email: tamish560@users.noreply.github.com

Do not open a public issue for security vulnerabilities.

## Supported Versions

| Version | Supported |
|---|---|
| Latest release | Yes |
| Older releases | No |

## Scope

Relay runs locally and makes outbound API calls to Anthropic only. It does not:
- Listen on any network port
- Accept inbound connections
- Store or transmit your data to any third party
- Include telemetry or analytics

The primary security surface is the license verification system and local file I/O.
