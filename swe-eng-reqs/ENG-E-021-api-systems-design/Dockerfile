FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN go test ./... && CGO_ENABLED=0 go build -o /out/eng-e-021 .

FROM alpine:3.20
RUN apk add --no-cache wget ca-certificates
COPY --from=build /out/eng-e-021 /eng-e-021
EXPOSE 8080
USER nobody
ENTRYPOINT ["/eng-e-021"]
