[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/new/template/wiki-go?utm_medium=integration&utm_source=button&utm_campaign=wiki-go)

# Wiki-Go on Railway

One-click Railway template for [Wiki-Go](https://github.com/leomoon-studios/wiki-go) (LeoMoon Wiki-Go), a flat-file Markdown wiki written in Go. No database: pages, versions, attachments, users and settings are files on a Railway volume.

## What it deploys

| Service | Source | What it does |
|---|---|---|
| Wiki-Go | this repo's `Dockerfile` (official `leomoonstudios/wiki-go:1.9.6`, pinned by digest) | The wiki, on a public Railway domain, health check `/login`, volume at `/wiki/data` |

The `Dockerfile` adds one small helper, `wiki-go-init` (source in [`init/`](init/main.go)), which runs before Wiki-Go on every start:

- **First boot:** writes `data/config.yaml` with one admin user from `WIKI_ADMIN_USERNAME` / `WIKI_ADMIN_PASSWORD` (bcrypt), so upstream's default `admin`/`admin` account is never created.
- **Every boot:** listens on `$PORT`, trusts Railway's proxy range (`100.64.0.0/10`) for client IPs so the login ban works per visitor, keeps cookies Secure, and replaces any admin password that is still the upstream default.
- Railway volumes are owned by root, so it gives `/wiki/data` to the image's `appuser` (uid 1000), drops to that user and then execs Wiki-Go. Any setup error exits non-zero, so Railway restarts the deploy instead of running an unsafe wiki.

## Variables

| Variable | Default | Notes |
|---|---|---|
| `WIKI_ADMIN_PASSWORD` | `${{secret(24)}}` | Generated admin password. Read it in the service's Variables tab. |
| `WIKI_ADMIN_USERNAME` | `admin` | Admin login name created on first boot. |
| `PORT` | `8080` | Port Wiki-Go listens on (the Railway domain points here). |
| `WIKI_RESET_ADMIN_PASSWORD` | not set | Optional: set to `true` for one deploy to apply a changed `WIKI_ADMIN_PASSWORD` to the existing admin. |

Everything else (wiki title, private mode, comments, upload size, languages, users) is set in Wiki-Go's admin panel and stored in `/wiki/data/config.yaml`.

## After deploying

1. Open the public URL and sign in as `admin` with the `WIKI_ADMIN_PASSWORD` value.
2. Change the password from the user menu if you like, and turn on **Private wiki** in the settings if pages shouldn't be public (anyone can read pages until you do; only logged-in users can edit).

## Resources

Fits Railway's Free plan (0.5 GB RAM, 1 vCPU). `GOMEMLIMIT=384MiB` keeps the Go heap inside that limit.

## Updates

The service builds from this repo, so when we bump the pinned Wiki-Go version on `main`, Railway shows deployers an "update available" notice; you choose when to apply it. Read [CHANGELOG.md](CHANGELOG.md) first. Data on the volume is kept.

## Licence

Wiki-Go is GPL-3.0 (© LeoMoon Studios). This wrapper is GPL-3.0 as well.
