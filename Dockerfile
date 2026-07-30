FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/air ./cmd/air

FROM alpine:3.21
RUN apk add --no-cache git ca-certificates
COPY --from=build /out/air /usr/local/bin/air
ENTRYPOINT ["air"]
CMD ["--help"]
