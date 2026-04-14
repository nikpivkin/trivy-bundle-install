# trivy-bundle-install

A [Trivy](https://github.com/aquasecurity/trivy) plugin that downloads and installs a checks bundle to the Trivy cache directory.

## Installation

```bash
trivy plugin install github.com/nikpivkin/trivy-bundle-install
```

## Usage

```bash
trivy bundle-install --bundle-url <url>
```

### Flags

| Flag | Description |
|------|-------------|
| `--bundle-url` | URL or path of the checks bundle to download (required) |
| `--cache-dir` | Trivy cache directory (default: auto-detected from trivy) |

> **Note:** Support for S3, GCS and other sources is planned.

### Examples

```bash
# From OCI registry
trivy bundle-install --bundle-url oci://ghcr.io/aquasecurity/trivy-checks:2

# From HTTP
trivy bundle-install --bundle-url https://example.com/bundle.tar.gz

# From local path
trivy bundle-install --bundle-url file:///path/to/bundle
```

After installing the bundle, pass `--skip-check-update` to Trivy to prevent it from overwriting the bundle with an update:

```bash
trivy conf --skip-check-update myproject
```
