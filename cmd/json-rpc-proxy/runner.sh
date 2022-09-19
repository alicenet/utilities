#!/usr/bin/env sh

set -eu

/usr/bin/envsubst '${PORT} ${REMOTE_PATH} ${CACHE_TIME}' < ./default.conf > /etc/nginx/conf.d/default.conf

exec "$@"
