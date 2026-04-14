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

> **Note:** Currently only local file paths are supported. Support for HTTP, S3, GCS and other sources is planned.

### Examples

```bash
trivy bundle-install --bundle-url file:///path/to/bundle
```

After installing the bundle, pass `--skip-check-update` to Trivy to prevent it from overwriting the bundle with an update:

```bash
trivy conf --skip-check-update myproject
```
