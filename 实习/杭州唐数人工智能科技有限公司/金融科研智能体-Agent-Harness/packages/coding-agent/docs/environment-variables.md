# Environment Variables

TFA uses environment variables in three ways:

- Variables such as `TFA_OFFLINE` configure the TFA process.
- TFA sets `TFA_CODING_AGENT` so child processes can detect that they run inside TFA.
- Commands run by the LLM-callable bash tool receive `PI_*` variables describing the current session.

Provider API-key variables are documented separately in [Providers](providers.md#environment-variables-or-auth-file).

## Process Marker

The CLI and RPC entry points set `TFA_CODING_AGENT=true`. Child processes inherit it and can use it to detect that they run inside TFA. It is not session-specific and is not set automatically when TFA is embedded through the SDK.

## Bash Tool Session Environment

Commands run by the bash tool receive the current TFA session state:

| Variable | Description |
|----------|-------------|
| `TFA_SESSION_ID` | Current session ID |
| `TFA_SESSION_FILE` | Absolute path to the current session JSONL file; unset for ephemeral sessions |
| `TFA_PROVIDER` | Currently selected model provider |
| `TFA_MODEL` | Currently selected model ID |
| `TFA_REASONING_LEVEL` | Current effective reasoning level: `off`, `minimal`, `low`, `medium`, `high`, `xhigh`, or `max` |

The values are resolved when each command starts. Switching models or changing the reasoning level therefore affects the next bash command without restarting TFA. `TFA_PROVIDER` and `TFA_MODEL` identify the selected TFA model, not a different upstream model that a router may choose internally.

When asked which model or provider is running, inspect these variables instead of inferring the answer from the system prompt:

```bash
printf '%s/%s\n' "$TFA_PROVIDER" "$TFA_MODEL"
printf 'reasoning=%s session=%s\n' "$TFA_REASONING_LEVEL" "$TFA_SESSION_ID"
```

The session file can be inspected directly when the session is persistent:

```bash
if [ -n "$TFA_SESSION_FILE" ]; then
  tail -n 1 "$TFA_SESSION_FILE"
fi
```

These variables are injected into the LLM-callable bash tool. They are not injected into user-entered `!` or `!!` commands.

### Custom Bash Tools

Bash tools created with `createBashTool()` expose the session environment by default when registered with TFA. Injection happens before `spawnHook`, so a hook receives the variables in `ctx.env`:

```typescript
const bashTool = createBashTool(cwd, {
  spawnHook: (ctx) => ({
    ...ctx,
    env: { ...ctx.env, CI: "1" },
  }),
});
```

Disable session metadata independently of the spawn hook:

```typescript
const bashTool = createBashTool(cwd, {
  exposeSessionEnvironment: false,
  spawnHook: (ctx) => ctx,
});
```

When disabled, TFA removes inherited values for these variables so nested TFA processes do not expose stale parent-session metadata.

## TFA Process Configuration

These variables are read by TFA itself:

| Variable | Description |
|----------|-------------|
| `TFA_CODING_AGENT_DIR` | Override the config directory; default is `~/.tfa/agent` |
| `TFA_CODING_AGENT_SESSION_DIR` | Override session storage; overridden by `--session-dir` |
| `TFA_PACKAGE_DIR` | Override the package directory, useful for Nix/Guix store paths |
| `TFA_OFFLINE` | Disable startup network operations, including update checks, package updates, and install/update telemetry |
| `TFA_SKIP_VERSION_CHECK` | Disable the `pi.dev` latest-version request |
| `TFA_TELEMETRY` | Override install/update telemetry and provider attribution headers: `1`/`true`/`yes` or `0`/`false`/`no` |
| `TFA_CACHE_RETENTION` | Set to `long` for extended provider prompt caching where supported |
| `TFA_SHARE_VIEWER_URL` | Override the base URL used by `/share` |
| `TFA_HARDWARE_CURSOR` | Set to `1` to show the hardware cursor; see [Terminal setup](terminal-setup.md) |
| `VISUAL`, `EDITOR` | External editor fallback when `externalEditor` is unset |
| `HTTP_PROXY`, `HTTPS_PROXY` | Proxy outbound HTTP requests |

Provider credentials such as `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, and cloud-provider configuration are listed in [Providers](providers.md#environment-variables-or-auth-file).
