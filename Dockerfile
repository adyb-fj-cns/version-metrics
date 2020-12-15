FROM library/golang as builder
RUN mkdir -p /build
WORKDIR /build
ADD . /build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-d -w -s' -o version-metrics .


FROM alpine
WORKDIR /app
COPY --from=builder /build/version-metrics .
EXPOSE 8000
ENTRYPOINT ./version-metrics