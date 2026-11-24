# Deploy and Host Wiki-Go on Railway

Wiki-Go (LeoMoon Wiki-Go) is an open-source, flat-file wiki written in Go. Pages are plain Markdown files on disk, with no database. It has a Markdown editor with Mermaid diagrams and LaTeX math, full-text search, page version history, file attachments, comments, Kanban boards and admin, editor and viewer roles.

## About Hosting Wiki-Go

This template runs one Wiki-Go service, built from the official `leomoonstudios/wiki-go` 1.9.6 image pinned by digest, with a Railway volume at `/wiki/data`. Every page, version, attachment, user and setting lives in that volume, so they survive restarts and redeploys. A small start helper writes the first config: the admin account is `admin` with a generated password (the `WIKI_ADMIN_PASSWORD` variable on the service), so upstream's default `admin`/`admin` login never exists. It also sets the port, trusts Railway's proxy for the login ban, and keeps session cookies HTTPS-only. After deploying, open the public URL, sign in and change the password in the user menu.

## Common Use Cases

* Internal team wiki and documentation hub
* Personal knowledge base with Markdown, diagrams and math
* Lightweight project tracking with Kanban boards next to your docs

## Dependencies for Wiki-Go Hosting

* One Railway service built from the template repo (Wiki-Go 1.9.6)
* One Railway volume mounted at `/wiki/data` (no database needed)

### Deployment Dependencies

* Wiki-Go source and docs: https://github.com/leomoon-studios/wiki-go
* Wiki-Go releases: https://github.com/leomoon-studios/wiki-go/releases
* Template repo: https://github.com/aalfath/wiki-go-railway-template

### Implementation Details

* **Sign in:** user `admin`, password from the `WIKI_ADMIN_PASSWORD` variable on the Wiki-Go service (Variables tab). If you change that variable later, also set `WIKI_RESET_ADMIN_PASSWORD=true`, redeploy once, then remove it.
* **Private wiki:** pages are readable by anyone until you turn on "Private wiki" in the admin settings. Editing always needs a login.
* **Settings:** use the admin panel, or edit `/wiki/data/config.yaml`. The template rewrites only the port, proxy and listen settings on each start.
* **Memory:** runs within Railway's 0.5 GB Free plan limit (tested idle and while editing, searching and uploading).
* **Updates:** the service builds from the GitHub repo, so Railway shows an "update available" notice when we bump the pinned Wiki-Go version. Read the repo's CHANGELOG.md before applying it. Your data stays on the volume.

## Why Deploy Wiki-Go on Railway?

<!-- Recommended: Keep this section as shown below -->
Railway is a singular platform to deploy your infrastructure stack. Railway will host your infrastructure so you don't have to deal with configuration, while allowing you to vertically and horizontally scale it.

By deploying Wiki-Go on Railway, you are one step closer to supporting a complete full-stack application with minimal burden. Host your servers, databases, AI agents, and more on Railway.
<!-- End recommended section -->
