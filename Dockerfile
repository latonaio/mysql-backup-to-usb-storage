# syntax = docker/dockerfile:experimental
FROM golang:1.16.2 as builder

ENV GO111MODULE on
ENV GOPRIVATE "github.com/latonaio"
WORKDIR /go/src/github.com/latonaio

COPY go.mod .

RUN git config --global url."git@github.com:".insteadOf "https://github.com/"
RUN mkdir /root/.ssh/ && touch /root/.ssh/known_hosts && ssh-keyscan -t rsa bitbucket.org >> /root/.ssh/known_hosts
RUN --mount=type=secret,id=ssh,target=/root/.ssh/id_rsa go mod download

COPY . .

RUN go build -o mysql-backup-to-usb-storage mysql-backup-to-usb-storage/app/

# Runtime Container
FROM alpine:3.12
RUN apk add --no-cache libc6-compat tzdata lsblk
COPY --from=builder /go/src/github.com/latonaio/mysql-backup-to-usb-storage .
CMD ["./mysql-backup-to-usb-storage"]
