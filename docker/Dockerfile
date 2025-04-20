FROM golang:1.24.1-bookworm AS builder
WORKDIR /app
COPY . .
RUN go build -o forum

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/forum .
CMD [ "./forum" ]

#docker build -t forum .
#docker run -v ${pwd}/forum.db:/app/forum.db -p 8080:8080 forum