# AGENTS.md

Guidelines for coding agents working in the `fnd-follower-counter` repository.

## Project Overview

`fnd-follower-counter` polls follower/subscriber counts from SNS providers (YouTube, Instagram, Twitch) and renders two 4-digit counts side by side on a TM1638 8-digit 7-segment FND module connected to Raspberry Pi GPIO pins (data=17, clk=27, stb=22).

## Package Structure

- `main.go`, `main_test.go`: CLI entrypoint (`parseFlags` for `-c` config path and `-n` console mode), `sideRunner`, signal-aware polling loop (`run`), and display initialization (`openDisplay`). Keep `main.go` at the repository root so `go build .` produces the binary.
- `internal/config`: YAML configuration parsing and validation (`Config`, `SideConfig`, `Target`, `Load`).
- `internal/display`: TM1638 display abstraction (`Device`), 4-digit count formatting (`formatCount4` with `k`/`m` abbreviations), side rendering (`Render`, `SideState`), console ASCII-art renderer (`ConsoleFND`), and startup IP-suffix welcome banner (`ShowWelcome`).
- `internal/sns`: Provider-agnostic `Fetcher` interface and `NewFetcher` factory backed by unexported provider implementations (`youtubeFetcher`, `instagramFetcher`, `twitchFetcher`) and a shared `jsonClient` HTTP helper.

## Build & Verification Commands

Run formatting, static analysis, and race-enabled tests before finishing any Go change:

```sh
gofmt -l .
go vet ./...
go test -race ./...
```

Cross-compile for Raspberry Pi (32-bit ARM):

```sh
GOOS=linux GOARCH=arm go build -o fnd-follower-counter .
```

## Configuration & Environment Variables

- Runtime flags: `-c <path>` selects the config YAML file (defaults to `config.yaml`) and `-n` renders the 8-digit FND output as ASCII art to stdout instead of opening TM1638 GPIO hardware. See `config.example.yaml`.
- Credentials are read from environment variables by `sns.NewFetcher`:
  - YouTube: `YOUTUBE_API_KEY`
  - Instagram: `INSTAGRAM_ACCESS_TOKEN`, `INSTAGRAM_IG_USER_ID`
  - Twitch: `TWITCH_CLIENT_ID`, `TWITCH_CLIENT_SECRET`
- Hardware GPIO access (`periph.io`) requires Raspberry Pi hardware at runtime; keep display and SNS logic behind interfaces/HTTP test servers so `go test ./...` runs without hardware or network access.

## Go Code Style (`golang-code-style`)

Follow `.agents/skills/golang-code-style/SKILL.md` for all Go code:

- **Line breaking**: Break lines beyond ~120 characters at semantic boundaries. Function calls with 4+ arguments must place one argument per line; multi-line signatures put each parameter on its own line.
- **Variables & literals**: Use `:=` for non-zero values and `var` for zero-value initialization. Initialize slices and maps explicitly (never nil). Always use field names in struct composite literals (including test tables).
- **Control flow**: Handle errors and edge cases first with early returns. Drop `else` after `return`/`break`/`continue`, and prefer default-then-override or `switch` over `if-else` chains. Extract 3+ operand conditions into named booleans.
- **Functions & loops**: Keep functions short and focused with `<=4` parameters (group beyond 4 into a struct), ordered `context.Context` first, inputs next, output destinations last. Prefer `range` (including `range n`) over index-based loops.
- **Visibility & organization**: Order files as package doc, imports, constants, types, constructors, methods, helpers. Keep concrete implementations unexported inside `internal/` packages unless external callers require them.
