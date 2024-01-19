FROM golang:1.21-alpine AS gobuild

WORKDIR /build
COPY cmd ./cmd
COPY internal ./internal
COPY quotes.json ./
COPY tokens.json ./
COPY sunswap.json ./
COPY go.mod go.sum ./
RUN apk --no-cache add git mercurial ca-certificates

RUN git config --global url.https://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@github.com/kattana-io.insteadOf https://github.com/kattana-io
RUN export GOPRIVATE=github.com/kattana-io

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/app ./cmd/main.go

FROM alpine:latest
RUN apk add tzdata
RUN echo "UTC" > /etc/timezone
WORKDIR /root/
COPY --from=gobuild /build/.bin/app .
COPY --from=gobuild /build/quotes.json .
COPY --from=gobuild /build/tokens.json .
COPY --from=gobuild /build/sunswap.json .
EXPOSE 8080
CMD ["./app"]
