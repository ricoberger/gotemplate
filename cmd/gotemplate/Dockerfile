FROM golang:1.18.3 as api
WORKDIR /gotemplate
COPY go.mod go.sum /gotemplate/
RUN go mod download
COPY . .
RUN export CGO_ENABLED=0 && make build

FROM alpine:3.16.0
RUN apk update && apk add --no-cache ca-certificates
RUN mkdir /gotemplate
COPY --from=api /gotemplate/bin/gotemplate /gotemplate
WORKDIR /gotemplate
USER nobody
ENTRYPOINT  [ "/gotemplate/gotemplate" ]
