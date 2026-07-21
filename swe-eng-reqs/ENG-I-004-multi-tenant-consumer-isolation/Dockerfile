FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN go test ./... && CGO_ENABLED=0 go build -o /out/eng-i-004 .

FROM alpine:3.20
RUN apk add --no-cache wget ca-certificates
COPY --from=build /out/eng-i-004 /eng-i-004
EXPOSE 8080
USER nobody
ENTRYPOINT ["/eng-i-004"]
