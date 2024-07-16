# Use the official Golang image to create a build artifact.
FROM golang:1.22 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -v -o server ./cmd/app

# Use a Docker multi-stage build to create a lean production image.
FROM alpine:3.14
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/server .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
