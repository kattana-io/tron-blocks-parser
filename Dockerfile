ARG PAT

FROM golang:1.21-alpine AS gobuild

WORKDIR /build

RUN apk --no-cache add git mercurial ca-certificates
COPY cmd ./cmd
COPY internal ./internal
COPY go.mod go.sum ./
COPY quotes.json ./
COPY tokens.json ./
COPY sunswap.json ./

RUN git config --global url.https://${PAT}@github.com/kattana-io.insteadOf https://github.com/kattana-io
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
