FROM golang:1.26.7-alpine3.24 AS builder
#Multistage.build
WORKDIR /ws
COPY . .
RUN go mod tidy
RUN go test -v -cover
RUN CGO_ENABLED=0 go build -o websocket .

FROM alpine:3.24
COPY --from=builder /ws/websocket .
RUN chmod +x websocket
EXPOSE 9012
CMD ["./websocket"]
