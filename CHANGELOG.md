# Changelog

Breaking changes are listed first in each entry. Deployers: read the entry before you apply a Railway "update available" notice. Your wiki data lives on the volume at `/wiki/data` and is kept across updates.

## 1.9.6-railway.1 (2026-11-24)

- First release of the Railway template.
- Wiki-Go 1.9.6 (`leomoonstudios/wiki-go:1.9.6@sha256:642fe29fcedbaede3bacc58aba7e3070bf554990ff57143ed0b77edaf5d26d9e`).
- `wiki-go-init` start helper: creates the admin from `WIKI_ADMIN_USERNAME`/`WIKI_ADMIN_PASSWORD` on first boot (no default `admin`/`admin`), replaces a leftover default admin password, listens on `$PORT`, trusts Railway's proxy range for client IPs, hands the Railway volume to the image's non-root user and exits non-zero on any setup error.
