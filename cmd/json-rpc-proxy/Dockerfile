FROM openresty/openresty:alpine

RUN apk add gettext

COPY ./runner.sh ./default.conf ./

CMD ["/bin/sh", "runner.sh", "/usr/local/openresty/bin/openresty", "-g", "daemon off;"]
