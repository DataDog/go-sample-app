FROM golang:1.20 AS builder
WORKDIR /notes
COPY . ./
RUN go build -o notes ./main.go

FROM debian:stable-slim
COPY --from=builder /notes/notes ./
RUN mkdir data
RUN chmod +x ./notes

CMD ["./notes"]
