# Hades

> Provisioning creates machines. Hades gives them a soul.

Hades is a **change-execution tool** for servers you fully own.

It's designed for explicit operations, built with clarity, simplicity, and predictability as core values.

Hades is inspired by the principle: **"Execute intent, don't infer it."**

## Quick Start

```bash
# Build Hades or download binary
make build

# init sample
hades init

# run plans
hades run boostrap
hades run deploy
...
```

## What is Hades?

Hades executes **explicit change** on your infrastructure:
- **Not a provisioning tool** (use Terraform for that)
- **Not a desired-state reconciler** (no hidden reconciliation loops)
- **Not a long-running controller** (ephemeral runs only)

Hades **is**:
- **An execution engine** - runs exactly what you tell it
- **A bootstrap/config tool** - setup, configuration, lifecycle
- **Human-first** - predictable, reviewable, copy-pasteable

## Core Principles

1. **Explicit over Implicit** - No magic, no hidden behavior
2. **Predictable** - Same input = same output, always
3. **Reviewable** - Dry-run/Audit logs shows operations per host
4. **Fail Fast** - Errors abort immediately
5. **Zero State** - Runs are ephemeral, no state stored

## Why Hades?

**vs Ansible/Salt** - Hades doesn't reconcile state. It executes changes explicitly.

**vs Terraform** - Hades doesn't manage cloud resources. It operates on existing servers.

## Contributing

Ask before.

## License

MIT License - See [LICENSE](LICENSE) for details.

