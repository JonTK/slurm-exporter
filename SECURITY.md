# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | Yes                |
| < 0.3   | No                 |

## Reporting a Vulnerability

**For security vulnerabilities, DO NOT create a public GitHub issue.**

Instead, please report security issues through:

1. **GitHub Security Advisories** (Recommended)
   - Navigate to the [Security tab](https://github.com/jontk/slurm-exporter/security/advisories)
   - Click "Report a vulnerability"
   - Fill out the security advisory form

### What to Include

- **Description** of the vulnerability
- **Steps to reproduce** the issue
- **Potential impact** and attack scenarios
- **Affected versions** if known
- **Suggested fix** if you have one

### Response Timeline

| Timeline | Action |
|----------|--------|
| **72 hours** | Initial acknowledgment of report |
| **7 days** | Initial assessment and severity classification |
| **30 days** | Security patch released (for high/critical issues) |

## Security Measures

### Authentication

- **Multiple Auth Methods**: JWT tokens, API keys, basic auth
- **JWT User Token Auth**: Sets both `X-SLURM-USER-NAME` and `X-SLURM-USER-TOKEN` headers
- **Secret File Support**: Tokens and passwords can be read from files with permission checks

### Transport Security

- **TLS Support**: Configurable TLS for the metrics endpoint with cipher suite selection
- **SLURM TLS**: Client certificate support for SLURM API connections
- **Rate Limiting**: Configurable request throttling to prevent API abuse

### Container Security

- **Minimal Base Images**: Scratch, Alpine, and Distroless variants available
- **Non-root Execution**: Containers run as non-privileged users
- **Multi-stage Builds**: Separate build and runtime environments

### Development Security

- **Static Analysis**: Gosec security scanning in CI
- **Vulnerability Scanning**: govulncheck for Go dependency vulnerabilities
- **Branch Protection**: Required reviews and status checks

## Security Best Practices

### For Users

1. **Use JWT auth with username**: Configure both `username` and `token` in auth config
2. **Use token files**: Prefer `token_file` over inline `token` values
3. **Enable TLS**: Configure TLS for production metrics endpoints
4. **Restrict permissions**: Token files should have `0600` permissions
5. **Use `validation.allow_insecure_connections: false`**: Only allow HTTPS for non-localhost connections

### For Operators

1. **Network Segmentation**: Isolate the exporter appropriately
2. **Rotate Tokens**: Regularly rotate JWT tokens
3. **Monitor Access**: Enable logging to track metrics endpoint access
4. **Principle of Least Privilege**: Create dedicated SLURM users for the exporter

---

**Last Updated**: February 2026
