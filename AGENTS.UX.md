# Hades Terminal UX Guidelines

This document defines the terminal output style and guidelines for Hades deployment automation tool.

## Philosophy

Hades terminal output follows these principles:

1. **Separation of Concerns**: Terminal shows lifecycle events and status; logs contain detailed command output
2. **Consistent Formatting**: All entities (actions, jobs, targets) follow the same pattern for lifecycle states
3. **Visual Hierarchy**: Symbols and colors provide quick visual scanning of deployment status
4. **Contextual Information**: Each message includes enough context (host, action index, name) to understand what's happening

## Symbol System

### Actions

Actions represent individual operations within a job (run, copy, mkdir, etc.).

| State | Symbol | Color | Format |
|-------|--------|-------|--------|
| In Progress | `◌` | Yellow | `[host] ◌ Action [index] type (name): in progress` |
| Completed | `●` | Green | `[host] ● Action [index] type (name): completed` |
| Skipped | `○` | Blue | `[host] ○ Action [index] type (name): skipped (reason)` |
| Failed | `●` | Red | `[host] ● Action [index] type (name): failed - error` |

**Examples:**
```
[web-01] ◌ Action [0] run: in progress
[web-01] ● Action [0] run: completed

[web-01] ◌ Action [1] copy (config-file): in progress
[web-01] ○ Action [1] copy (config-file): skipped (/etc/app/config.yml already up to date)

[web-01] ◌ Action [2] mkdir: in progress
[web-01] ● Action [2] mkdir: failed - permission denied
```

### Jobs

Jobs are collections of actions executed on a host.

| State | Symbol | Color | Format |
|-------|--------|-------|--------|
| Starting | `◇` | Yellow | `[host] ◇ Job "name": starting` |
| Completed | `◆` | Green | `[host] ◆ Job "name": completed` |
| Skipped | `◇` | Blue | `[host] ◇ Job "name": skipped (guard failed)` |
| Failed | `◆` | Red | `[host] ◆ Job "name": failed - error` |

**Examples:**
```
[web-01] ◇ Job "install-caddy": starting
[web-01] ◌ Action [0] run: in progress
[web-01] ● Action [0] run: completed
[web-01] ◆ Job "install-caddy": completed

[web-02] ◇ Job "install-postgres": skipped (guard failed)
```

### Steps

Steps are the top-level execution units in a plan. Each step shows its status, targets, and timestamp.

| State | Symbol | Color | Format |
|-------|--------|-------|--------|
| Started | `□` | Yellow | `Status: □ Started` |
| Completed | `■` | Green | `Status: ■ Completed` |
| Failed | `■` | Red | `Status: ■ Failed` |

**Examples:**
```
Step 1/2: Deploy to production
  Job: deploy-app
  Targets: web-servers, api-servers
  Status: □ Started
  Started: 2026-02-07 10:45:32

[web-01] ◇ Job "deploy-app": starting
[web-01] ◆ Job "deploy-app": completed

[api-01] ◇ Job "deploy-app": starting
[api-01] ◆ Job "deploy-app": completed

  Status: ■ Completed
```

## Color Coding

All colors use the `github.com/wzshiming/ctc` package constants:

| Color | Usage | Constant |
|-------|-------|----------|
| Yellow | In Progress / Starting | `ctc.ForegroundYellow` |
| Green | Completed / Success | `ctc.ForegroundGreen` |
| Blue | Skipped | `ctc.ForegroundBlue` |
| Red | Failed / Error | `ctc.ForegroundRed` |
| Cyan | (Reserved for future use) | `ctc.ForegroundCyan` |

### Color Application Pattern

```go
fmt.Fprintf(stdout, "[%s] %s●%s Action [0] run: completed\n",
    hostName, ctc.ForegroundGreen, ctc.Reset, ...)
```

Always use `ctc.Reset` after the symbol to prevent color bleeding.

## Message Format Patterns

### Action Messages

**Format:**
```
[hostname] SYMBOL Action [index] type (optional-name): state (optional-details)
```

**Components:**
- `[hostname]`: Host where action executes (always present)
- `SYMBOL`: Lifecycle symbol with color
- `Action`: Literal word "Action"
- `[index]`: Zero-based action index (always present)
- `type`: Action type (run, copy, mkdir, template, etc.)
- `(optional-name)`: User-provided name if defined in YAML
- `state`: Lifecycle state (in progress, completed, skipped, failed)
- `(optional-details)`: Additional context (skip reason, error message, etc.)

**Examples:**
```
[web-01] ◌ Action [0] run: in progress
[web-01] ● Action [1] copy (backup): completed
[web-01] ○ Action [2] copy (config): skipped (/etc/app.conf already up to date)
[web-01] ● Action [3] run: failed - command execution failed: exit status 1
```

### Job Messages

**Format:**
```
[hostname] SYMBOL Job "job-name": state (optional-details)
```

**Components:**
- `[hostname]`: Host where job executes
- `SYMBOL`: Lifecycle symbol with color
- `Job`: Literal word "Job"
- `"job-name"`: Quoted job name from plan
- `state`: Lifecycle state (starting, completed, skipped, failed)
- `(optional-details)`: Additional context (guard condition, error, etc.)

**Examples:**
```
[web-01] ◇ Job "install-caddy": starting
[web-01] ◆ Job "install-caddy": completed
[web-02] ◇ Job "install-postgres": skipped (guard failed)
[web-03] ◆ Job "deploy-app": failed - action 2 failed: permission denied
```

### Target Messages

**Note:** Targets are displayed in the step header only. No separate target lifecycle messages are shown. Target groups are listed in the `Targets:` field of each step, and execution progresses directly to job/action level output.

**Example:**
```
Step 1/1: Deploy
  Job: install-app
  Targets: web-servers, db-servers, api-servers
  Status: □ Started
  Started: 2026-02-07 10:45:32

[web-01] ◇ Job "install-app": starting
[web-01] ◆ Job "install-app": completed

[db-01] ◇ Job "install-app": starting
[db-01] ◆ Job "install-app": completed

  Status: ■ Completed
```

### Step Messages

**Format:**
```
Step N/M: step-name
  Job: job-name
  Targets: target1, target2
  Status: SYMBOL state
  Started: YYYY-MM-DD HH:MM:SS
```

**Components:**
- `Step N/M`: Step number and total count
- `Job`: Job name to execute
- `Targets`: Comma-separated list of target groups
- `Status`: Step status with symbol (□ Started, ■ Completed, ■ Failed)
- `Started`: Timestamp when step began (format: 2006-01-02 15:04:05)

**Examples:**
```
Step 1/2: Deploy to production
  Job: install-app
  Targets: web-servers, api-servers
  Status: □ Started
  Started: 2026-02-07 10:45:32

(execution happens here)

  Status: ■ Completed
```

**Failed Step:**
```
Step 2/2: Database migration
  Job: migrate-db
  Targets: db-servers
  Status: □ Started
  Started: 2026-02-07 10:46:15

(execution with errors)

  Status: ■ Failed
```

### Plan Messages

**Format:**
```
==========
Plan: plan-name
==========

Run ID: hades-YYYYMMDD-HHMMSS
Started: ISO-8601-timestamp

(steps execute here)

✓ Plan completed successfully
Duration: X.Xs
```

## Terminal vs Logs

### Terminal Output

Terminal receives:
- Plan start/completion
- Step header with status and timestamp
- Job lifecycle (per host)
- Action lifecycle (per host)
- Skip messages with action format
- Step completion status

Terminal shows **status and lifecycle**, not command output.

### Log Files

Log files (`logs/{run-id}/{plan}.{host}.out.log` and `.err.log`) contain:

1. **Action Delimiters** (for executed actions):
   ```
   ====================
   JOB: job-name, ACTION: [index] type
   STARTED: YYYY-MM-DD HH:MM:SS
   --------------------

   (command output)
   ```

2. **Skip Messages** (plain text, no delimiters):
   ```
   Skipping /etc/app/config.yml (already up to date)
   ```

3. **Command Output** (stderr/stdout from SSH commands)

**Important:** Log files contain **plain text only** - no ANSI color codes or escape sequences.

## Implementation Guidelines

### Adding New Actions

When creating new actions that can skip execution:

1. **Write to logs** using `runtime.Stdout` (plain text)
2. **Write to console** using `runtime.ConsoleStdout` with:
   - Action format: `Action [index] type (name)`
   - Available via `runtime.ActionDesc`
   - Blue ○ symbol for skipped state

**Example:**
```go
// Log: plain text
fmt.Fprintf(runtime.Stdout, "Skipping %s (reason)\n", path)

// Console: formatted with symbol
if runtime.ConsoleStdout != nil {
    fmt.Fprintf(runtime.ConsoleStdout, "[%s] %s○%s Action %s: skipped (%s reason)\n",
        runtime.Host.Name, ctc.ForegroundBlue, ctc.Reset, runtime.ActionDesc, path)
}
```

### Adding New Lifecycle States

If adding new states to existing entities:

1. Choose appropriate symbol from Unicode
2. Assign color based on semantic meaning:
   - Yellow: Transient/in-progress states
   - Green: Success states
   - Blue: Neutral/informational states
   - Red: Error/failure states
3. Update this document with the new state
4. Maintain consistent format pattern

### Error Messages

Error messages should:
- Include enough context to understand what failed
- Be concise (one line when possible)
- Include error details after a dash: `failed - reason`

**Good:**
```
[web-01] ● Action [0] run: failed - command execution failed: exit status 1
[web-02] ◆ Job "deploy": failed - action 2 failed: permission denied
```

**Bad:**
```
[web-01] Failed!
[web-02] Error
```

## Visual Examples

### Successful Deployment

```
==========
Plan: deploy
==========

Run ID: hades-20260207-104532
Started: 2026-02-07T10:45:32+01:00

Step 1/1: Deploy application
  Job: install-app
  Targets: web-servers
  Status: □ Started
  Started: 2026-02-07 10:45:32

[web-01] ◇ Job "install-app": starting
[web-01] ◌ Action [0] run (install): in progress
[web-01] ● Action [0] run (install): completed
[web-01] ◌ Action [1] copy (config): in progress
[web-01] ○ Action [1] copy (config): skipped (/etc/app.conf already up to date)
[web-01] ◆ Job "install-app": completed

[web-02] ◇ Job "install-app": starting
[web-02] ◌ Action [0] run (install): in progress
[web-02] ● Action [0] run (install): completed
[web-02] ◌ Action [1] copy (config): in progress
[web-02] ● Action [1] copy (config): completed
[web-02] ◆ Job "install-app": completed

  Status: ■ Completed

✓ Plan completed successfully
Duration: 12.5s
```

### Failed Deployment

```
Step 1/1: Deploy application
  Job: install-app
  Targets: web-servers
  Status: □ Started
  Started: 2026-02-07 10:45:32

[web-01] ◇ Job "install-app": starting
[web-01] ◌ Action [0] run: in progress
[web-01] ● Action [0] run: failed - command execution failed: exit status 127
[web-01] ◆ Job "install-app": failed - action 0 failed: command execution failed

  Status: ■ Failed
```

### Skipped Job (Guard)

```
Step 1/1: Install Caddy
  Job: install-caddy
  Targets: web-servers
  Status: □ Started
  Started: 2026-02-07 10:45:32

[web-01] ◇ Job "install-caddy": skipped (guard failed)

[web-02] ◇ Job "install-caddy": starting
[web-02] ◌ Action [0] run: in progress
[web-02] ● Action [0] run: completed
[web-02] ◆ Job "install-caddy": completed

  Status: ■ Completed
```

## Code Locations

- **Executor**: `hades/executor/executor.go` - Main lifecycle message output
- **Actions**: `hades/actions/*.go` - Action-specific skip messages
- **UI**: `hades/ui/output.go` - Shared UI helper methods
- **Colors**: `github.com/wzshiming/ctc` - Color constants

## Summary

Following these guidelines ensures:
- ✅ Consistent visual language across all deployment operations
- ✅ Quick visual scanning of terminal output
- ✅ Clear separation between status (terminal) and details (logs)
- ✅ Maintainable codebase with predictable output patterns
