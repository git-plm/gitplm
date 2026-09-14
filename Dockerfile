FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=docker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=${VERSION}" -o /out/gitplm .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /workspace

COPY --from=build /out/gitplm /usr/local/bin/gitplm

RUN addgroup -S gitplm && adduser -S -G gitplm gitplm \
	&& mkdir -p /home/gitplm \
	&& chown -R gitplm:gitplm /home/gitplm /workspace

USER gitplm

ENV HOME=/home/gitplm

EXPOSE 7654

ENTRYPOINT ["gitplm"]
CMD ["version"]
