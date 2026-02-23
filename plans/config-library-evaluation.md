<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Config library evaluation (issue #470)

## Problem

`applyEnvOverrides` is a procedural if-chain that grows with every new
config knob. Env var names are disconnected from the fields they map to,
making it easy to add a field but forget the env override (or vice versa).

## Candidates evaluated

### koanf

Composable provider pipeline: confmap (defaults) -> file (TOML) -> env ->
unmarshal to struct.

Pros:
- Clean loading pipeline abstraction.
- Built-in dot-delimited key access (`k.Get("llm.model")`).
- Widely used, well maintained.

Cons:
- **5+ new transitive deps** (koanf core, providers, parsers, mapstructure,
  copystructure).
- **Irregular env var naming** defeats automatic mapping.
  `MICASA_MAX_DOCUMENT_SIZE` -> `documents.max_file_size`, `OLLAMA_HOST` ->
  `llm.base_url` with `/v1` appending -- none of these follow a simple
  prefix-strip convention.
- **Custom types** (ByteSize, Duration) need mapstructure decode hooks,
  duplicating logic already in `UnmarshalTOML` / `Parse*` functions.
- **Get() semantics change**: koanf returns raw values (string "50 MiB" vs
  typed uint64 52428800). Current `Get()` returns the resolved typed value,
  which is more useful for `config --get`.
- Validation stays manual either way.

Verdict: dependency cost outweighs the gains for a CLI app with ~12 env vars.

### viper

Similar to koanf but heavier. Pulls in spf13/pflag, fsnotify, multiple
encoding libs. Same env-mapping impedance mismatch. Rejected.

### caarlos0/env (chosen)

Zero-dependency library that reads `env` struct tags and populates fields.

Pros:
- **Zero transitive dependencies**.
- **Struct tags are the single source of truth** for env var names.
- **FuncMap** supports custom type parsers -- ByteSize and Duration parsers
  plug in directly, reusing existing `Parse*` functions.
- **Handles pointer types natively** -- `*bool`, `*int`, `*Duration` work
  out of the box (with registered parsers for custom types).
- **Parse errors surface** instead of being silently swallowed -- invalid
  env var values produce actionable error messages.

Cons:
- OLLAMA_HOST `/v1` suffix still needs post-processing (one-off special case).

Verdict: right level of abstraction. Handles the boilerplate (tag reading, type
conversion, pointer wrapping) while letting us own the custom parsing and
validation.

## Implementation

- `env` struct tags added to all Config fields that accept env overrides.
- `applyEnvOverrides` replaced with `env.ParseWithOptions` + custom FuncMap
  for ByteSize and Duration.
- OLLAMA_HOST `/v1` normalization moved to post-processing in `LoadFromPath`.
- `EnvVars()` function derives env -> config key mapping from struct tags.
- `TestEnvVars` and `TestEnvVarsCoverAllKeys` guard against drift.
- BurntSushi/toml, `defaults()`, validation, `Get()`/`Keys()`, and
  `ExampleTOML()` are unchanged.
