# AgentRemote

AgentRemote brings all your agents into one app.

Beeper becomes the universal remote for agents.

Connect agent runtimes to Beeper with full history, live streaming, tool approvals, and encrypted delivery.

Run the bridge next to your agent, then talk to it from Beeper on your phone or desktop.

## Why Use It

- Keep agents on your own machine, server, or private network
- Use Beeper instead of building a separate web UI
- Stream responses and approve tool calls in the same chat
- Reach your agents from anywhere Beeper runs

## Open Source Focus

This repository is centered on the self-hosted path.

That means:

- local developer machines
- homelabs
- office servers
- runtimes behind a firewall
- private deployments that still want a polished remote interface

There is a broader product direction around richer AI chats and more opinionated agent experiences. Open source here is focused on making the bridge layer for private deployments easy to run and hard to break.

## Included Bridges

Each bridge has its own README with setup details and scope:

| Bridge | Purpose |
| --- | --- |
| `ai` | General Matrix-to-AI bridge surface used by the project |
| [`codex`](./bridges/codex/README.md) | Connect the Codex CLI app-server to Beeper |
| [`openclaw`](./bridges/openclaw/README.md) | Connect a self-hosted OpenClaw gateway to Beeper |
| [`opencode`](./bridges/opencode/README.md) | Connect a self-hosted OpenCode server to Beeper |

## Quick Start

Log into Beeper and start a bridge:

```bash
./tools/bridges login --env prod
./tools/bridges run codex
```

Then open Beeper and use the connected bridge from chat.

For a local Beeper environment:

```bash
./tools/bridges login --env local
./tools/bridges whoami
./tools/bridges run codex
```

Configured instances in `bridges.manifest.yml`:

- `ai`
- `codex`
- `openclaw`
- `opencode`

Run any of them directly:

```bash
./tools/bridges run ai
./tools/bridges run codex
./tools/bridges run openclaw
./tools/bridges run opencode
```

Or use the wrapper:

```bash
./run.sh ai
./run.sh codex
./run.sh openclaw
./run.sh opencode
```

## Bridge Manager

Common commands:

```bash
./tools/bridges list
./tools/bridges status
./tools/bridges logs codex --follow
./tools/bridges restart codex
./tools/bridges down codex
./tools/bridges whoami
```

Reset all local bridge state and registrations:

```bash
./tools/bridges delete --remote ai
./tools/bridges delete --remote codex
./tools/bridges delete --remote openclaw
./tools/bridges delete --remote opencode
./tools/bridges logout
```

## Docs

- [`docs/bridge-orchestrator.md`](./docs/bridge-orchestrator.md): local bridge management workflow
- [`docs/matrix-ai-matrix-spec-v1.md`](./docs/matrix-ai-matrix-spec-v1.md): Matrix transport profile for streaming, approvals, state, and AI payloads
- [`bridges/codex/README.md`](./bridges/codex/README.md): Codex bridge details
- [`bridges/openclaw/README.md`](./bridges/openclaw/README.md): OpenClaw bridge details
- [`bridges/opencode/README.md`](./bridges/opencode/README.md): OpenCode bridge details

## Status

Experimental and evolving quickly. The transport and bridge surfaces are real, but the project is still early.

## Build

Requires `libolm` for encryption support.

```bash
./build.sh
```

Or with Docker:

```bash
docker build -t agentremote .
```
