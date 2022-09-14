FROM scratch
COPY indexer-worker /indexer-worker
ENTRYPOINT ["/indexer-worker"]
