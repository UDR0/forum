FROM golang:1.24
WORKDIR /app
COPY . .
RUN go build -o forum
EXPOSE 8080
CMD ["./forum"]

