ARG GO_IMAGE=golang:1.23-alpine
FROM ${GO_IMAGE} AS builder

RUN go install github.com/jackc/tern/v2@latest

FROM postgres:17-alpine

WORKDIR /app

COPY --from=builder /go/bin/tern /usr/local/bin/tern

COPY ./migrations /app/migrations
COPY ./configuration/tern.conf /app/tern.conf
COPY ./scripts/run.sh /app/run.sh

RUN chmod +rx /app/run.sh && \
    chmod -R +r /app/migrations /app/tern.conf

CMD ["/app/run.sh"]