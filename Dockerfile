# Pinned gcr.io/distroless/static-debian12:nonroot using pinny
FROM gcr.io/distroless/static-debian12@sha256:43a5ce527e9def017827d69bed472fb40f4aaf7fe88c356b23556a21499b1c04 
COPY bolt-cli /usr/bin/bolt-cli
ENTRYPOINT ["/usr/bin/bolt-cli"]
LABEL org.opencontainers.image.source https://github.com/koalalab-inc/bolt-cli
