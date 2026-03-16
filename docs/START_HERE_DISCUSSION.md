# Start Here: install, auth, troubleshooting, and bug reports

Maintainer note: post this as a pinned GitHub Discussion when the repo is getting extra traffic.

Thanks for checking out `asc`. If you are new here, this post is the fastest way to get from install to a working command.

## 1. Install

```bash
brew install asc
```

If you are not using Homebrew, use the install script from `https://asccli.sh/install`.

## 2. Authenticate

```bash
asc auth login \
  --name "MyApp" \
  --key-id "ABC123" \
  --issuer-id "DEF456" \
  --private-key /path/to/AuthKey.p8 \
  --network
```

If keychain access is blocked, you are in CI, or you want config-backed auth instead:

```bash
asc auth login \
  --bypass-keychain \
  --name "MyCIKey" \
  --key-id "ABC123" \
  --issuer-id "DEF456" \
  --private-key /path/to/AuthKey.p8
```

## 3. Validate auth

```bash
asc auth status --validate
asc auth doctor
```

## 4. Run a first command

```bash
asc apps list --output table
```

If you want structured output instead:

```bash
asc apps list --output json --pretty
```

## Known limitations and common gotchas

- `asc` is unofficial and App Store Connect behavior can change without notice
- Some workflows depend on API coverage that may still be evolving or inconsistent across endpoints
- Auth can resolve from keychain, config files, and environment variables; mixed sources can be confusing
- Output defaults are TTY-aware: `table` in terminals, `json` in pipes and CI
- If auth behaves differently in automation, retry with `ASC_BYPASS_KEYCHAIN=1`

## Where to ask for help

- Use Discussions for install help, auth setup, workflow advice, and "how do I...?" questions
- Use Issues for reproducible bugs and concrete feature requests

## How to file a useful bug

Please include:

- `asc version`
- your OS
- your install method
- the exact command you ran
- redacted stdout/stderr
- whether it reproduces with `ASC_BYPASS_KEYCHAIN=1`
- redacted `ASC_DEBUG=api` or `asc --api-debug ...` output when safe

If you are filing an auth or API issue, running `asc auth doctor` first is especially helpful.
