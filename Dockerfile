FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:3.16.2

COPY tskeyservice /usr/local/bin/tskeyservice

RUN mkdir -p /data/tskeyservice
WORKDIR /data/tskeyservice

ENTRYPOINT ["/usr/local/bin/tskeyservice"]