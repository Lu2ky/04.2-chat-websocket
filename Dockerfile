FROM golang:1.26.8-alpine3.24 AS builder

WORKDIR /ws
COPY . .
RUN go mod tidy
RUN go test -v -cover
RUN go build -o websocket .


FROM alpine:latest
COPY --from=builder /ws/websocket .
RUN chmod u+x websocket
EXPOSE 9012
CMD ["./websocket"]