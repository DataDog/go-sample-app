FROM golang:1.20 AS builder
WORKDIR /users
COPY . ./
RUN go build -o users ./main.go

FROM debian:stable-slim
COPY --from=builder /users/users ./
RUN mkdir data
RUN chmod +x ./users

CMD ["./users"]
