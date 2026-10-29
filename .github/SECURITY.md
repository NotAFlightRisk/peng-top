# Security policy

Thanks for helping keep this project and the people using it safe 💛

---

## Supported versions

Our version history is linear, and we use semantic versioning. Only the latest release gets patched, so check you're on it before reporting anything.

---

## What it does and doesn't protect

peng-top is read-only. It reads system stats and serves them, and there's no way to kill, signal or change anything from the UI or the API.

Anyone who can reach it can see everything it shows, though, and that includes your process list and every process's command line (which is where secrets passed as flags tend to end up). So:

- It listens on `127.0.0.1` by default. Binding to anything else (like the Docker image does) is your call, so put it behind a firewall, a VPN or a reverse proxy
- Set `PENGTOP_AUTH=user:pass` (or `--auth`) for basic auth, and only send that over HTTPS
- Unless you've set a login, it only answers requests whose `Host` header is an IP address, `localhost`, the machine's own name or one you've listed in `PENGTOP_ALLOWED_HOSTS`. That stops a random web page using DNS rebinding to read it from your browser (with a login set, the browser won't hand it over anyway)
- The Docker image runs as root with the host's filesystem mounted read-only at `/host`, because reading other users' processes needs it. If you'd rather not, drop the `/host` mount (and `--pid host`) and it'll only see the container

A bug in any of those (a way round auth or the `Host` check, a write path, reading files outside what's shown) is a vulnerability. Seeing your own processes after you've exposed it to the internet without auth isn't.

---

## Reporting a vulnerability

> [!IMPORTANT]
> Please don't report anything publicly (issues, PRs, discussions) without giving us at least 30 days to respond and handle it.

Acceptable channels:

1. **GitHub** (preferred) - open an [advisory](https://github.com/NotAFlightRisk/peng-top/security/advisories/new) from the Security tab
2. **Email** - [security@peng.ly](mailto:security@peng.ly) (PGP: [`A8431F9F332FB0CD`](https://github.com/NotAFlightRisk.gpg))

Include the type of issue, the affected version and file paths, steps to reproduce, a PoC if you've got one, and what an attacker could actually achieve with it.

Keep it short, and don't over-state the impact - it's not helpful. Using AI is fine, just say that you have.

### What happens next

I'll try to acknowledge and triage within 48 hours. If it's valid, I'll get a fix out within 2 weeks and publish the advisory, and I'll keep you updated throughout.

---

## Disclosure

Coordinated, so give us a fair chance to ship a fix before going public. Happy to credit you in the advisory and release notes, or leave you out of it - your call.

---

## Safe harbour

Report in good faith, stick to this policy, and steer clear of privacy violations, data destruction and service disruption, and I won't pursue or support legal action against you.
