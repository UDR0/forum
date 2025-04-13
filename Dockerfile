FROM golang:1.24-alpine

RUN apk add --no-cache gcc musl-dev

ENV CGO_ENABLED=1

WORKDIR /app
COPY . .

RUN go build -o forum

EXPOSE 8080
CMD ["./forum"]



#docker build -t forum .
#docker run -v ${pwd}\forum.db:/app/forum.db -p 8080:8080 forum

