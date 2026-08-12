---
name: change-configuration
description: Adds or changes JioTV Go configuration options consistently across schema tags, defaults, example files, documentation, consumers, and tests. Use when adding, renaming, removing, defaulting, or changing behavior of a JIOTV_* environment variable or JSON, YAML, or TOML config key.
---

# Change Configuration

Treat a configuration option as one contract exposed through four inputs: environment, JSON, YAML, and TOML.

## Trace before editing

1. Find the field in `internal/config.JioTVConfig` and every consumer of that field.
2. Establish precedence and semantics for omitted, zero, empty, and explicitly false values.
3. Check whether the option affects startup order, persistent paths, handlers, generated playlists, Docker examples, or browser behavior.

Do not add a setting for behavior that has no real deployment need. Prefer an existing option or fixed project behavior when either satisfies the request.

## Update the contract

- Add matching `yaml`, `env`, `json`, and `toml` tags on the schema field. Environment names use `JIOTV_*`; file keys use snake case.
- Keep the field comment, tag names, documented default, and runtime default identical.
- Preserve explicit false values. `cleanenv` boolean defaults can overwrite an explicit `false`; follow the existing manual-default pattern only when a true default is required.
- Update the consumer at the narrowest existing boundary. Do not read `os.Getenv` outside configuration loading when `config.Cfg` already owns the value.
- For a rename or removal, migrate every consumer and documented/sample name in the same change. Leave no alias unless compatibility is explicitly required.

## Synchronize user-facing surfaces

Update every applicable artifact:

- `configs/jiotv-config.toml`
- `configs/jiotv-config.yml`
- `configs/jiotv-config.json`
- `docs/config.md`, including its embedded examples
- deployment snippets in `docs/get_started.md`, Docker files, or compose files only when users configure the option there

Keep sample values at the real default unless the sample explicitly demonstrates another value. Do not rewrite unrelated entries while synchronizing formats.

## Prove behavior

Add focused tests for the changed contract:

- Load the option through each supported file format when parsing differs by type, especially slices and structured values.
- Load it through its environment variable.
- Check omitted/default behavior and explicit zero, empty, or false behavior where meaningful.
- Test the consuming behavior separately when parsing success alone does not prove the feature.

Use temporary files and restore global `config.Cfg` or process environment changed by a test. Avoid parallel tests that mutate globals, working directory, or environment.

Run the focused config and consumer tests, then `go test ./...`. If documentation examples changed, compare their keys and defaults against `JioTVConfig` before finishing.
