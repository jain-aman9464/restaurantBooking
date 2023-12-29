# Use an official Go runtime as a parent image
FROM golang:1.16-alpine

# Set the working directory inside the container
WORKDIR /go/src/app

# Copy the local package files to the container's workspace
COPY . .

# Install any dependencies
RUN go get -u -v ./...

# Build the Go application
RUN go build -o app

# Expose the port that your application will run on
EXPOSE 8080

# Set environment variables for RabbitMQ
ENV RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/"

# Run both the main app and email worker in the background
CMD ["sh", "-c", "./app & ./email_worker"]
