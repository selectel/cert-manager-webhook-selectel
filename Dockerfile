FROM golang:1.25.5-alpine3.23 AS build
RUN apk add --no-cache git
WORKDIR /workspace
ENV GO111MODULE=on
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o webhook -ldflags '-w -extldflags "-static"' .

# ------------------------------
FROM alpine:3.23
RUN apk add --no-cache ca-certificates
COPY --from=build /workspace/webhook /usr/local/bin/webhook
USER nobody
ENTRYPOINT ["webhook"]
