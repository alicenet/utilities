FROM golang:1.18-alpine AS builder

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/app cmd/worker/main.go

FROM gcr.io/distroless/static

COPY --from=builder /go/bin/app /

CMD [ "/app" ]
