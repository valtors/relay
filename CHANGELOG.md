# Changelog

## v0.4.9 (2026-07-11)

## Changes

- Changed Go module path from `relay` to `github.com/valtors/relay` for pkg.go.dev indexing
- Added pre-push verification script (`scripts/pre-push.sh`)
- Added comparison guide (`docs/comparison.md`)
- Configurable web timeout via `RELAY_WEB_TIMEOUT` env var


## v0.4.8 (2026-07-11)

## Changelog
* ed118d1885edadb71ab8834391932a0b031c9dac bump: v0.4.8
* 73235466fc5efcac5a958a3a42a0ae7fcbc20756 bump: v0.4.8
* 57b874c3da5ff0dfce6a2a9e337b8fed0edb57de chore: add Dockerfile for Glama verification
* f9501184c700a8009ecf4fec2425622b097d4902 feat: add doctor command to diagnose installs
* d6cbd8b813baaf0bf1ff10023e6b44ad707c6836 feat: wire doctor command into cli
* ba4298569ac540edd00a8c69dc19acc84639e033 style: gofmt doctor.go
* ce77451c4c444fadbd1e7e01fcbd98a5d3ddd360 style: gofmt doctor_test.go
* efc72216fea3410df19cf172f7175ea369270cbf style: gofmt main.go



## v0.4.7 (2026-07-11)

## Changelog
* 79f565f315e99bd818318c20fd20c19e7d9f1fc1 chore: bump npm package to v0.4.7
* 20f391134f26b1d1976007445c819239810d85cb style(tui): remove icon prefix from editor names



## v0.4.6 (2026-07-11)

## Changelog
* 41e462ccc0c9008d83022da70559b3f11731d5fa chore: bump npm package to v0.4.6
* 084e341d9d6db8a094de9431c4364dc2dfa3f5ca fix(cli): launch TUI by default in interactive terminals



## v0.4.5 (2026-07-11)

## Changelog
* ba8cc1c6652a4c3be6c55296470c432d619a62b4 chore: bump npm package to v0.4.5
* b9de3cf7c4e9a7432efd84d0186de539b64abd31 fix(tui): import SelectInput and remove emoji icons
* 73f0112ef4487baf696ad8a6a46362d2fb1c26cd style(tui): replace emoji glyphs with ASCII markers



## v0.4.4 (2026-07-11)

## Changelog
* b7d551f8852d5a56d69976ec8d53b9308f0d3284 chore: bump npm package to v0.4.4
* 7d826dbf5fb80abd7469fefa4e8441cde56dcca2 fix(cli): suggest global install when TUI cannot run via npx
* fc613a0b09792dbf355757f7b649430a01d73e2b fix(tui): pass onExit handler to StatusDashboard
* cf96a2174650d1420c0e5b5e88cfddf560dd70d5 fix(tui): repair broken JSX closing tags



## v0.4.3 (2026-07-11)

## Changelog
* a88ac2152cf4f70bc5206a078969dae44a3af575 chore: bump npm package to v0.4.3
* 2c11b12de86545710692e20134632be16d1be687 fix(cli): always launch TUI for bare npx userelay



## v0.4.2 (2026-07-11)

## Changelog
* bebbe1b8e7af16e4e2de548ec7fc63e183d096d1 chore: bump npm package to 0.4.0
* cce544315e8246e3b67bd30938831b4d1e458d35 chore: bump npm package to 0.4.0
* 966253ec954141ae04c3deb88a8dcde5df8e0d68 chore: bump npm package to v0.4.1
* cf11da82cf4fa01bfb0cc6bfeec428a87434a381 chore: bump npm package to v0.4.2
* 1bedde1fb946ecad4bc69eb62ceadd475315cb30 feat(cli): add --version and -v flags
* 0e77e2b6e3e1e3aa7a0ec244245d76ff08883878 feat(cli): auto-launch setup wizard on first run
* e875c93c8902d42485a02f655b1cc7dfe7fa0ee0 feat(cli): auto-launch setup wizard on first run
* db2009b3cf719a771cc2994079bdb1697770f414 feat(cli): show hint when running in stdio mode in a terminal
* 9b07e30e9813ea33d1c915524464b2c36f8ebe38 feat(tui): add animated first-run setup wizard
* a248e9384e825938b745a47368464b9622b5a818 feat(tui): add setup screen option to TUI router
* aad486285edae3c657eacd7094b22d5472a61d6f feat(tui): add setup screen option to TUI router
* 1459c84d5a160a12985a2614a350c7ccfb9afccf feat(tui): animated first-run setup wizard
* 4fc2a15a3e3d0550858587357f2e692f1726831f feat(ui): add renderHint helper for terminal hints
* b3cbc7232537dd2fa27a74995b2e1b86826773c3 fix(cli): detect TTY via stdout/stderr for npx-launched TUI
* 7d64ff284c48f3b165633a06818a8635fad77b16 fix(tui): correct broken JSX closing tags
* 7ed2b6af96983e02fc257320bb47e7ee36e413e9 fix(tui): correct broken JSX closing tags
* e6ce61c52c791256c1362eab7abe80471d5d9447 fix: correct all-contributors CLI command and permissions



## v0.4.0 (2026-07-10)

## What's new

### Security
- Path traversal protection for file tools
- SSRF mitigation for web tools
- XSS sanitization for output
- Prompt injection guards

### Features
- Interactive TUI mode
- Spectral-edge CLI banner
- Clean CLI output with > markers

### Refactor
- Split tool registrations from registry into per-category functions

### Repo
- Issue and PR templates
- Code of Conduct
- Dependabot and stale bot configs
- Funding.yml
- Good first issues for contributors

**Full Changelog**: https://github.com/valtors/relay/compare/v0.3.0...v0.4.0

## v0.3.0 (2026-07-05)

## Changelog
* afba595b957d884dec12c3b1e0b981c6db15e0cc add image tools and comprehensive test suite
* 42b7a61904f09e100fed3efa8bf6d58e08b53851 add pdf tools, goreleaser, and install scripts
* 4cc19df1773590343b1da570886c8566ead20e23 add relay init wizard and npx package wrapper
* 9311152ea7cda17cab3a1faaca75bb1408fcd51b add tool registry, CLI subcommands, and 27 built-in tools
* a623453c289110fd60a3e596e589055e657ece95 fix critical security issues: add LICENSE, rewrite SECURITY.md, patch Go vulns
* f59678e3f57e1ceb2263e1b5918e614e0ad0cd02 init: relay MCP server - multi-agent launch pipeline in Go
* 1b12f5c3c83e6fbcf3a3e36622d41506f04bd31a refactor: remove all code comments
* f1402323acee37cf29f865251cfa87f5e2879f23 remove license check, relay is open source now
* 13911d80881255f2ddb5c64bf65c60a235e1a560 rename npm package to userelay (relaymcp was blocked)
* 2950f9660181a6fb50b87478d5c0f0c45b572e85 rewrite README: outcome-first, 5-min quickstart, 3 workflows, verified configs
* 19c97f0defa2d129c5b9134759c3e55cf12a376f rewrite readme for expanded product scope
* d097175675fe6dfba7397ce8c1f4749d87b317cb ship npx @valtors/relay: one-command setup, fix release chain
* c7b10a542614753b0bc2a272975d544b56fd1bf3 use 'npx relaymcp' as primary install command



