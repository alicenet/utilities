FROM scratch
COPY indexer-frontend /indexer-frontend
ENTRYPOINT ["/indexer-frontend"]
