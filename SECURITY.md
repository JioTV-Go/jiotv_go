# Security Policy

## Supported Versions

Security fixes are applied to the latest released version and the active `main` for stable and `develop` branches. Older releases are not supported unless a maintainer explicitly confirms otherwise.

## Reporting a Vulnerability

Do **not** open a public issue for a suspected vulnerability.

Use [GitHub private vulnerability reporting](https://github.com/jiotv-go/jiotv_go/security/advisories/new). Include:

- affected version, commit, or deployment method;
- impact and attack prerequisites;
- reproducible steps or proof of concept;
- relevant configuration details and sanitized logs; and
- a proposed fix, if available.

Do not include Jio credentials, OTPs, access tokens, API keys, private URLs, or other secrets. Redact them before submitting the report.

If private reporting is unavailable, contact a repository maintainer privately through email rather than using the public issue tracker.

## Response and Disclosure

Maintainers will assess reports privately, reproduce the issue, and coordinate a fix before disclosure. Please allow reasonable time for triage and remediation; avoid public disclosure until maintainers confirm a fix or agree on a disclosure date.

## Scope

Relevant reports include vulnerabilities in the Go server, web UI, release artifacts, container images, CI configuration, and handling of credentials, DRM, encrypted stream URLs, or upstream proxy requests. Reports limited to third-party JioTV service availability or account support are out of scope unless this repository's code creates the vulnerability.
