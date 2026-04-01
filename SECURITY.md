# Security policy

## Supported versions

Only the **latest release** (or `main`) receives security fixes. Prefer upgrading to the newest tagged version.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| older   | :x:                |

## Reporting a vulnerability

**Do not** open a public GitHub issue for security vulnerabilities.

Please report them **privately** using one of these:

1. [GitHub Security Advisories](https://github.com/samirkoirala/devops-doctor/security/advisories/new) — *Preferred* if you have a GitHub account.
2. Or contact the maintainer via GitHub profile / email listed there, with subject line including `SECURITY: devops-doctor`.

Include:

- Description of the issue and impact
- Steps to reproduce (if possible)
- Affected versions or `go` / OS context if relevant

You should receive an initial response within a few days. We will coordinate a fix and disclosure timeline with you.

## Scope

This project is a **local CLI** that runs diagnostic commands on the operator’s machine. It does not expose a network service. Reports about “RCE via malicious compose file” on the same machine as the operator are generally out of scope unless the tool introduces an unexpected trust boundary; still, we welcome discussion on the [issue tracker](https://github.com/samirkoirala/devops-doctor/issues) for design improvements.

## Recommendations for users

- Install from this repository or official releases; verify tags when fetching with `go install …@vx.y.z`.
- Run with least privilege; the tool may invoke `docker`, `kubectl`, etc., with your user’s permissions.
