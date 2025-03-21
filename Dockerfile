FROM golang:1.24
WORKDIR /app
COPY . .
RUN go build -o forum
CMD ["./forum"]
EXPOSE 8080
