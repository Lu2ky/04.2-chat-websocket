FROM golang:latest AS builder

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