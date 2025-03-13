ARG RELEASE_TAG=develop 

FROM golang:latest-alpine as builder

ARG RELEASE_TAG

RUN apk update && apk add --no-cache git make ca-certificates tzdata && update-ca-certificates

ENV USER=appuser
ENV UID=10001

# See https://stackoverflow.com/a/55757473/12429735
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"
WORKDIR $GOPATH/src/resonatecoop/id

COPY go.mod .
RUN git clone --branch ${RELEASE_TAG} --single-branch --depth 1 https://github.com/resonatecoop/id

ENV GO111MODULE=on
RUN go mod download && go mod verify

COPY . .

RUN make install
RUN make generate

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' -a \
    -o /go/bin/id .

FROM scratch

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /go/bin/id /go/bin/id

# Use an unprivileged user.
USER appuser:appuser

ENTRYPOINT ["/go/bin/id"]
