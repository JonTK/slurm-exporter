# Security Policy

## 🛡️ Supported Versions

We actively maintain security updates for the following versions:

| Version | Supported          | End of Life    |
| ------- | ------------------ | -------------- |
| 1.x.x   | ✅ Yes             | TBD            |
| 0.x.x   | ⚠️ Security only   | 2024-12-31     |

## 🚨 Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please follow our responsible disclosure process:

### 📧 How to Report

**For security vulnerabilities, DO NOT create a public GitHub issue.**

Instead, please report security issues through one of these secure channels:

1. **GitHub Security Advisories** (Recommended)
   - Navigate to the [Security tab](https://github.com/jontk/slurm-exporter/security/advisories)
   - Click "Report a vulnerability"
   - Fill out the security advisory form

2. **Email** (Alternative)
   - Send an email to: security@slurm-exporter.io
   - Include "SECURITY" in the subject line
   - Encrypt with our PGP key if possible

### 📝 What to Include

Please provide as much information as possible:

- **Description** of the vulnerability
- **Steps to reproduce** the issue
- **Potential impact** and attack scenarios
- **Affected versions** if known
- **Suggested fix** if you have one
- **Your contact information** for follow-up

### ⏰ Response Timeline

| Timeline | Action |
|----------|--------|
| **24 hours** | Initial acknowledgment of report |
| **72 hours** | Initial assessment and severity classification |
| **7 days** | Detailed investigation and response plan |
| **30 days** | Security patch released (for high/critical issues) |
| **90 days** | Public disclosure (after patch is available) |

### 🏆 Security Researchers

We appreciate security researchers who help keep our project safe:

- **Recognition**: Contributors will be credited in our security advisories (unless they prefer anonymity)
- **Hall of Fame**: Acknowledged security researchers are listed in our documentation
- **Responsible Disclosure**: We follow coordinated vulnerability disclosure practices

## 🔒 Security Measures

### Code Security

- **Static Analysis**: Automated security scanning with Gosec, StaticCheck
- **Dependency Scanning**: Regular vulnerability checks with Nancy, Snyk
- **Secret Detection**: TruffleHog and GitLeaks scans
- **License Compliance**: Automated license compatibility checks

### Container Security

- **Vulnerability Scanning**: Trivy and Grype container scans
- **Minimal Base Images**: Using distroless/alpine for reduced attack surface
- **Non-root Execution**: Containers run as non-privileged users
- **Multi-stage Builds**: Separate build and runtime environments

### Infrastructure Security

- **Secure Defaults**: Security-first configuration templates
- **TLS Encryption**: All communications encrypted in transit
- **Access Controls**: Principle of least privilege
- **Audit Logging**: Comprehensive security event logging

### Development Security

- **Branch Protection**: Required reviews and status checks
- **Signed Commits**: Encouragement of GPG-signed commits
- **Security Training**: Regular security awareness for maintainers
- **Threat Modeling**: Regular security architecture reviews

## 🔧 Security Features

### Authentication & Authorization

- **Multiple Auth Methods**: JWT, API keys, certificates, basic auth
- **Token Validation**: Comprehensive token verification
- **Permission Scoping**: Fine-grained access controls
- **Session Management**: Secure session handling

### Data Protection

- **Input Validation**: Comprehensive input sanitization
- **Output Encoding**: Proper encoding of all outputs
- **Error Handling**: Secure error messages (no information leakage)
- **Logging Security**: Sensitive data excluded from logs

### Network Security

- **TLS Configuration**: Strong cipher suites and protocols
- **Rate Limiting**: Protection against DoS attacks
- **CORS Policies**: Appropriate cross-origin restrictions
- **Security Headers**: Comprehensive HTTP security headers

### Monitoring & Detection

- **Security Metrics**: Built-in security monitoring
- **Circuit Breakers**: Protection against cascading failures
- **Health Checks**: Comprehensive system health monitoring
- **Alerting**: Real-time security event notifications

## 🚀 Security Best Practices

### For Users

1. **Keep Updated**: Always use the latest stable version
2. **Secure Configuration**: Follow our security hardening guides
3. **Network Security**: Deploy in secure network environments
4. **Access Controls**: Implement proper authentication and authorization
5. **Monitoring**: Enable security monitoring and alerting
6. **Backup Security**: Secure backup and recovery procedures

### For Developers

1. **Secure Coding**: Follow OWASP secure coding practices
2. **Input Validation**: Validate all inputs rigorously
3. **Error Handling**: Implement secure error handling
4. **Secrets Management**: Never commit secrets to version control
5. **Dependency Updates**: Keep dependencies updated
6. **Security Testing**: Include security tests in your workflow

### For Operators

1. **Deployment Security**: Use secure deployment practices
2. **Network Segmentation**: Isolate SLURM Exporter appropriately
3. **Monitoring**: Implement comprehensive security monitoring
4. **Incident Response**: Have an incident response plan
5. **Regular Audits**: Conduct regular security assessments
6. **Documentation**: Maintain security documentation

## 📚 Security Resources

### Documentation

- [Security Configuration Guide](docs/security-configuration.md)
- [Threat Model](docs/security-threat-model.md)
- [Security Architecture](docs/security-architecture.md)
- [Incident Response Playbook](docs/security-incident-response.md)

### External Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [Go Security Guidelines](https://github.com/OWASP/Go-SCP)
- [Container Security Guide](https://owasp.org/www-project-docker-top-10/)

### Tools & Scanners

- [Gosec](https://github.com/securego/gosec) - Go security analyzer
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) - Go vulnerability scanner
- [Trivy](https://github.com/aquasecurity/trivy) - Container vulnerability scanner
- [TruffleHog](https://github.com/trufflesecurity/trufflehog) - Secret scanner

## 🏅 Security Hall of Fame

We thank the following security researchers for their responsible disclosure:

*No security issues reported yet - be the first to help secure SLURM Exporter!*

## 📞 Contact Information

- **Security Team**: security@slurm-exporter.io
- **General Contact**: hello@slurm-exporter.io
- **Project Maintainers**: [GitHub Team](https://github.com/orgs/jontk/teams/maintainers)

## 📜 Security Advisory History

All security advisories are published at:
- [GitHub Security Advisories](https://github.com/jontk/slurm-exporter/security/advisories)
- [Security RSS Feed](https://github.com/jontk/slurm-exporter/security/advisories.atom)

---

**Last Updated**: December 2024  
**Next Review**: March 2025

*This security policy is regularly reviewed and updated. For questions or suggestions about this policy, please open a GitHub discussion or contact our security team.*