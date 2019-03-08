FROM golang:1.12-alpine as build
ENV GOOS linux
ENV GOARCH 386
WORKDIR /usr/src/pipe-to-me
COPY go.mod .
RUN go mod download
COPY *.go ./
RUN go build

FROM busybox
COPY --from=build /usr/src/pipe-to-me/pipe-to-me /pipe-to-me
EXPOSE 8080
ENTRYPOINT ["/pipe-to-me", "-httpaddr", ":8080"]