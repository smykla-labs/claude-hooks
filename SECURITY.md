# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
|---------|--------------------|
| latest  | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it by emailing the maintainers directly rather than opening a public issue.

**Please do not report security vulnerabilities through public GitHub issues.**

To report a security vulnerability:

1. **Email**: Contact the project maintainers directly
2. **Provide details**: Include as much information as possible about the vulnerability:

   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

3. **Wait for response**: You should receive a response within 48 hours
4. **Coordinated disclosure**: Please allow time for the vulnerability to be fixed before public disclosure

## Security Update Process

1. **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
2. **Investigation**: We will investigate and validate the reported vulnerability
3. **Fix**: We will develop and test a fix
4. **Release**: We will release a security update
5. **Disclosure**: We will publicly disclose the vulnerability after the fix is released

## Best Practices

When using Claude Hooks:

- Keep your installation up to date with the latest version
- Review hook configurations regularly
- Use GPG signing for commits (`-S` flag)
- Enable signoff for commits (`-s` flag)
- Follow the principle of least privilege when configuring validators
- Review logs regularly: `~/.claude/hooks/dispatcher.log`

## Security Features

Claude Hooks includes several security features:

- **Command validation**: Validates commands before execution
- **Path protection**: Blocks writes to sensitive paths like `/tmp`
- **Commit signing enforcement**: Requires GPG-signed commits
- **Signoff validation**: Enforces Developer Certificate of Origin
- **Timeout protection**: Prevents long-running validations from hanging

## Known Security Considerations

- Claude Hooks runs as a PreToolUse hook and has access to command inputs
- Logs are stored in plaintext at `~/.claude/hooks/dispatcher.log`
- Git operations may execute external commands

## Acknowledgments

We appreciate the security research community's efforts to responsibly disclose vulnerabilities.
