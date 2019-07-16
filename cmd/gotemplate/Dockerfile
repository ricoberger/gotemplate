FROM alpine:latest

ARG REVISION
ARG VERSION

LABEL maintainer="Rico Berger"
LABEL git.ref=$REVISION
LABEL git.version=$VERSION
LABEL git.url="https://github.com/ricoberger/gotemplate"

RUN apk add --no-cache --update curl ca-certificates
HEALTHCHECK --interval=10s --timeout=3s --retries=3 CMD curl --fail http://localhost:8080/_health || exit 1

RUN addgroup -g 1000 gotemplate && \
    adduser -D -u 1000 -G gotemplate gotemplate
USER gotemplate

COPY ./bin/gotemplate-linux-amd64  /bin/gotemplate
EXPOSE 8080

ENTRYPOINT  [ "/bin/gotemplate" ]
