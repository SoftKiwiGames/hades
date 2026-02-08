# Hades Environment Variable Guide

Environment variables in Hades follow a strict contract system to ensure safety and predictability.

## Environment Variable Contract

Every job declares an **environment contract** specifying which variables it expects:

```yaml
jobs:
  deploy:
    env:
      VERSION:              # Required (no default)
      REGION:               # Required (no default)
      MODE:
        default: production # Optional (has default)
    actions:
      - run: echo "Deploying ${VERSION} to ${REGION}"
```

### Required Variables

Variables **without a default** are **required**:

```yaml
env:
  VERSION:      # Must be provided
  REGION:       # Must be provided
```

Hades will **fail validation** if required variables are not provided.

### Optional Variables

Variables **with a default** are **optional**:

```yaml
env:
  MODE:
    default: production
  LOG_LEVEL:
    default: info
```

If not provided, the default value is used.

## Environment Merging Priority

Environment variables merge with the following priority (highest to lowest):

1. **CLI flags** (`-e KEY=VALUE`)
2. **Step-level env** (in plan)
3. **Plan-level env** (in plan)
4. **Job defaults**

### Example

```yaml
jobs:
  deploy:
    env:
      MODE:
        default: production  # Priority 4: Default
      VERSION:

plans:
  staging:
    env:
      VERSION: v1.0.0        # Priority 3: Plan-level
    steps:
      - name: deploy-1
        job: deploy
        env:
          MODE: staging      # Priority 2: Step overrides plan and default

      - name: deploy-2
        job: deploy
        # Inherits: VERSION=v1.0.0 (from plan), MODE=production (from default)
```

```bash
# Priority 1: CLI overrides everything
hades run staging -e MODE=development -e VERSION=v2.0.0
# Result: MODE=development, VERSION=v2.0.0
```

### Plan-Level Variables (NEW)

You can now define environment variables at the plan level to avoid repetition:

```yaml
plans:
  multi-region:
    env:                     # ← Plan-level env (applies to ALL steps)
      VERSION: v1.0.0
      ENV: production
    steps:
      - name: deploy-us
        job: deploy
        env:
          REGION: us-east-1  # Step adds REGION, inherits VERSION and ENV

      - name: deploy-eu
        job: deploy
        env:
          REGION: eu-west-1  # Step adds REGION, inherits VERSION and ENV
          VERSION: v1.0.1    # Step can override plan-level VERSION
```

This is much cleaner than repeating `VERSION` and `ENV` in every step!

## Built-in Variables (HADES_*)

Hades automatically injects these variables for every job:

| Variable | Description | Example |
|----------|-------------|---------|
| `HADES_RUN_ID` | Unique run identifier | `abc-123-def` |
| `HADES_PLAN` | Plan name | `production-deploy` |
| `HADES_TARGET` | Target group | `app-servers` |
| `HADES_HOST_NAME` | Current host name | `app-1` |
| `HADES_HOST_ADDR` | Current host address | `192.168.1.10` |

**Important**: You **cannot** define or override `HADES_*` variables. Attempts to do so will fail validation.

```yaml
# ❌ INVALID - Will fail
env:
  HADES_RUN_ID: custom

# ❌ INVALID - Will fail
hades run deploy -e HADES_PLAN=custom
```

## Validation Rules

### Rule 1: All Required Variables Must Be Provided

```yaml
jobs:
  deploy:
    env:
      VERSION:    # Required

plans:
  # ❌ INVALID - Missing VERSION
  bad-plan:
    steps:
      - name: deploy
        job: deploy
        # No env provided

  # ✅ VALID - VERSION provided
  good-plan:
    steps:
      - name: deploy
        job: deploy
        env:
          VERSION: v1.0.0
```

### Rule 2: No Unknown Variables Allowed

```yaml
jobs:
  deploy:
    env:
      VERSION:    # Only VERSION is defined

plans:
  # ❌ INVALID - UNKNOWN_VAR not in contract
  bad-plan:
    steps:
      - name: deploy
        job: deploy
        env:
          VERSION: v1.0.0
          UNKNOWN_VAR: value  # Not defined in job
```

This prevents typos and ensures all variables are intentional.

### Rule 3: No HADES_* Variables

```yaml
# ❌ INVALID - Job cannot define HADES_* vars
jobs:
  bad-job:
    env:
      HADES_CUSTOM: value

# ❌ INVALID - Step cannot define HADES_* vars
steps:
  - name: deploy
    job: deploy
    env:
      HADES_RUN_ID: custom

# ❌ INVALID - CLI cannot provide HADES_* vars
$ hades run deploy -e HADES_PLAN=custom
```

## OS Environment Variable Expansion

Use `${VAR}` syntax to reference OS environment variables:

```yaml
plans:
  deploy:
    steps:
      - name: deploy
        job: deploy
        env:
          VERSION: ${GIT_TAG}      # Expands from OS env
          COMMIT: ${GIT_COMMIT}    # Expands from OS env
```

```bash
# Set in shell
export GIT_TAG=v1.2.3
export GIT_COMMIT=abc123

# Run plan
hades run deploy
# VERSION will be v1.2.3, COMMIT will be abc123
```

Expansion happens **once** before execution. Missing OS variables cause immediate failure.

## Common Patterns

### Pattern 1: Environment-Specific Defaults

```yaml
jobs:
  deploy:
    env:
      MODE:
        default: production
      REPLICAS:
        default: "3"
      LOG_LEVEL:
        default: warn

plans:
  # Production: use all defaults
  prod-deploy:
    steps:
      - name: deploy
        job: deploy
        env:
          VERSION: v1.0.0

  # Staging: override some defaults
  staging-deploy:
    steps:
      - name: deploy
        job: deploy
        env:
          VERSION: v1.0.0
          MODE: staging
          REPLICAS: "1"
          LOG_LEVEL: debug
```

### Pattern 2: CLI Override Everything

```yaml
# Define minimal plan
plans:
  flexible-deploy:
    steps:
      - name: deploy
        job: deploy

# Provide everything via CLI
$ hades run flexible-deploy \
  -e VERSION=v2.0.0 \
  -e REGION=us-west \
  -e MODE=production
```

### Pattern 3: Version from Git

```yaml
plans:
  auto-version:
    steps:
      - name: deploy
        job: deploy
        env:
          VERSION: ${GIT_TAG}
```

```bash
# Automatically use git tag
export GIT_TAG=$(git describe --tags)
hades run auto-version
```

### Pattern 4: Different Envs Per Step

```yaml
plans:
  multi-region:
    steps:
      - name: deploy-us
        job: deploy
        targets: [us-servers]
        env:
          VERSION: v1.0.0
          REGION: us-east-1

      - name: deploy-eu
        job: deploy
        targets: [eu-servers]
        env:
          VERSION: v1.0.0
          REGION: eu-west-1
```

## Error Messages

### Missing Required Variable

```
Error: environment validation failed: step 0 (deploy-app):
required environment variable "VERSION" not provided
```

**Fix**: Provide the variable in step env or via CLI:
```bash
hades run deploy -e VERSION=v1.0.0
```

### Unknown Variable

```
Error: environment validation failed: step 0 (deploy-app):
unknown environment variable "UNKNOWN_VAR" (not defined in job env contract)
```

**Fix**: Either add the variable to the job's env contract, or remove it from the step.

### HADES_* Override Attempt

```
Error: failed to expand environment variables:
user cannot define HADES_* environment variables: HADES_RUN_ID
```

**Fix**: Remove the `HADES_*` variable. These are auto-injected and cannot be overridden.

### Missing OS Variable

```
Error: failed to expand environment variables:
missing OS environment variables: GIT_TAG
```

**Fix**: Set the OS variable before running:
```bash
export GIT_TAG=v1.0.0
hades run deploy
```

## Best Practices

1. **Explicit Contracts**: Define all variables your job uses in the env contract
2. **Sensible Defaults**: Provide defaults for optional/environment-specific variables
3. **Required for Critical**: Make version/tag variables required (no default)
4. **Use OS Expansion**: Leverage `${VAR}` for CI/CD integration
5. **Document Variables**: Add comments explaining what each variable does
6. **Fail Fast**: Let validation catch errors before SSH connections

## Example: Complete Contract

```yaml
jobs:
  production-deploy:
    # All variables documented and typed
    env:
      # Required: Must be provided for every deployment
      VERSION:              # Semantic version (e.g., v1.2.3)
      DEPLOY_USER:          # User performing deployment

      # Optional: Environment-specific
      REGION:
        default: us-east-1  # AWS region
      MODE:
        default: production # Deployment mode

      # Optional: Operational settings
      REPLICAS:
        default: "3"        # Number of instances
      HEALTH_CHECK_TIMEOUT:
        default: "30s"      # Health check timeout

    actions:
      - run: |
          echo "Deploying ${VERSION}"
          echo "Region: ${REGION}"
          echo "Mode: ${MODE}"
          echo "User: ${DEPLOY_USER}"
```

## Troubleshooting

**Variables not expanding**: Check that you're using `${VAR}` syntax, not `$VAR`.

**Default not working**: Ensure the default is a string (use quotes for numbers: `"123"`).

**Validation passing but value wrong**: Check priority - CLI > step > defaults.

**Need to skip validation**: Not supported. All jobs must satisfy their contract.
