FROM scratch
COPY github-listener /github-listener
ENTRYPOINT ["/github-listener"]
