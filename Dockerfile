FROM golang:1.27.3-alpine3.24 AS build
RUN apk add --no-cache git
WORKDIR /workspace
ENV GO111MODULE=on
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o webhook -ldflags '-w -extldflags "-static"' .

# ------------------------------
FROM alpine:3.24.4
RUN apk add --no-cache ca-certificates
COPY --from=build /workspace/webhook /usr/local/bin/webhook
ENTRYPOINT ["webhook"]
