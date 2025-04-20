FROM golang:1.24.1
RUN apt-get update && apt-get install -y gcc libc-dev



# Enable CGO for SQLite
ENV CGO_ENABLED=1

# Set working directory
WORKDIR /app

# Copy your application files into the container
COPY . .

# Build the Go application
RUN go build -o forum

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./forum"]


#docker build -t forum .
#docker run -v ${pwd}\forum.db:/app/forum.db -p 8080:8080 forum

