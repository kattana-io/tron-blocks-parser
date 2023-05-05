FROM golang:1.20-alpine AS gobuild

WORKDIR /build
COPY cmd ./cmd
COPY internal ./internal
COPY quotes.json ./
COPY tokens.json ./
COPY go.mod go.sum ./
RUN apk --no-cache add git mercurial ca-certificates
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/app ./cmd/main.go

FROM alpine:latest
RUN apk add tzdata
RUN echo "UTC" > /etc/timezone
WORKDIR /root/
COPY --from=gobuild /build/.bin/app .
COPY --from=gobuild /build/quotes.json .
COPY --from=gobuild /build/tokens.json .
EXPOSE 8080
CMD ["./app"]
