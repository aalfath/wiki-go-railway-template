# Wiki-Go for Railway: the official image, pinned by version and digest, plus a
# small first-boot helper (init/) that creates the admin from generated
# variables instead of upstream's default admin/admin.

FROM docker.io/library/golang:1.27.3-alpine3.24@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 AS init
WORKDIR /src
COPY init/ ./
RUN CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags="-s -w" -o /wiki-go-init .

FROM docker.io/leomoonstudios/wiki-go:1.9.6@sha256:642fe29fcedbaede3bacc58aba7e3070bf554990ff57143ed0b77edaf5d26d9e

COPY --from=init /wiki-go-init /usr/local/bin/wiki-go-init

# Starts as root only to hand the Railway volume (mounted at /wiki/data, owned by
# root) to appuser; wiki-go-init then drops to appuser and execs Wiki-Go.
USER root
WORKDIR /wiki

# Railway's edge rewrites X-Forwarded-For to "<visitor>, <edge IP>" (edge IPs
# vary), so every hop is trusted and Wiki-Go's login ban keys on the visitor.
ENV PORT=8080 \
    TZ=UTC \
    CONFIGFILE=/wiki/data/config.yaml \
    WIKI_TRUSTED_PROXIES=0.0.0.0/0,::/0 \
    GOMEMLIMIT=384MiB

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/wiki-go-init"]
