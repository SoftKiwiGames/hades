# Hades Registry Guide

Registries provide immutable artifact storage for deployments. Use registries to store built artifacts and deploy them across your infrastructure.

## Supported Registry Types

### Filesystem Registry

Stores artifacts on the local filesystem or shared network storage.

```yaml
registries:
  prod:
    type: filesystem
    path: /var/hades/registry

  staging:
    type: filesystem
    path: /tmp/staging-registry
```

### S3 Registry (Future)

S3 registry support is planned but not yet implemented. Use filesystem registries for now.

## Registry Operations

### Push Action

Push an artifact from the artifact manager to a registry.

```yaml
- push:
    registry: prod          # Registry name
    artifact: app-binary    # Artifact name (from job artifacts)
    name: myapp            # Name in registry
    tag: ${VERSION}        # Tag/version (supports env vars)
```

**Requirements:**
- Artifact must be loaded in the job's `artifacts` section
- Registry must be defined in `registries` section
- Tag must be unique (registries are immutable)

### Pull Action

Pull an artifact from a registry and copy it to a remote host.

```yaml
- pull:
    registry: prod         # Registry name
    name: myapp           # Artifact name in registry
    tag: ${VERSION}       # Tag/version (supports env vars)
    to: /opt/app/binary   # Destination path on remote host
```

**Features:**
- Automatically copies to remote host via SSH
- Sets file mode to 0644 by default
- Supports environment variable expansion in all fields

## Complete Example

```yaml
registries:
  prod:
    type: filesystem
    path: /var/hades/artifacts

jobs:
  publish:
    artifacts:
      app-binary:
        path: ./build/myapp
    actions:
      - push:
          registry: prod
          artifact: app-binary
          name: myapp
          tag: ${VERSION}

  deploy:
    env:
      VERSION:
    actions:
      - mkdir:
          path: /opt/myapp
          mode: 0755
      - pull:
          registry: prod
          name: myapp
          tag: ${VERSION}
          to: /opt/myapp/app
      - run: chmod +x /opt/myapp/app
      - run: systemctl restart myapp

plans:
  release:
    steps:
      - name: publish
        job: publish
        targets: [build-server]
        env:
          VERSION: v1.0.0

      - name: deploy-canary
        job: deploy
        targets: [app-servers]
        limit: 1
        env:
          VERSION: v1.0.0

      - name: deploy-all
        job: deploy
        targets: [app-servers]
        env:
          VERSION: v1.0.0
```

## Best Practices

1. **Versioning**: Use semantic versioning for tags (v1.0.0, v1.0.1, etc.)
2. **Immutability**: Never reuse tags. Once published, tags are permanent.
3. **Separation**: Use separate registries for staging and production
4. **Checksums**: Registries automatically verify artifact integrity
5. **Cleanup**: Implement external cleanup policies (registries don't auto-delete)

## Workflow Patterns

### Build → Publish → Deploy

```yaml
steps:
  - name: build
    job: build
    targets: [build-server]

  - name: publish
    job: publish
    targets: [build-server]

  - name: deploy
    job: deploy
    targets: [app-servers]
```

### Canary with Registry

```yaml
steps:
  - name: publish
    job: publish
    targets: [build-server]
    env:
      VERSION: ${TAG}

  - name: canary
    job: deploy
    targets: [app-servers]
    limit: 1
    env:
      VERSION: ${TAG}

  - name: gate
    job: confirm
    targets: [build-server]

  - name: rollout
    job: deploy
    targets: [app-servers]
    parallelism: "3"
    env:
      VERSION: ${TAG}
```

### Multi-Environment

```yaml
registries:
  staging:
    type: filesystem
    path: /var/registry/staging

  prod:
    type: filesystem
    path: /var/registry/prod

# Promote from staging to prod
- pull:
    registry: staging
    name: myapp
    tag: ${VERSION}
    to: /tmp/artifact

- push:
    registry: prod
    artifact: myapp
    name: myapp
    tag: ${VERSION}
```

## Troubleshooting

**Artifact not found**: Ensure the artifact is defined in the job's `artifacts` section before using `push`.

**Already exists error**: Registries are immutable. Use a new tag or version number.

**Registry not found**: Check that the registry is defined in the `registries` section at the top level of your hadesfile.

**Permission denied**: Ensure the registry path is writable by the user running Hades.
