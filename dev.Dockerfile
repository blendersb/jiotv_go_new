# Use the official Golang image based on Alpine Linux
FROM golang:1.25-alpine

# Set environment variables
ENV GO111MODULE=on \
    GOEXPERIMENT=jsonv2,greenteagc \
    JIOTV_DEBUG=true \
    JIOTV_PATH_PREFIX="/app/.jiotv_go"

# Create a non-root user with UID 10001 (make sure this is not already in use on your system)
RUN adduser -D -u 10001 user

# Set the working directory to /app
WORKDIR /app

# Create a build directory to compile your Go app
RUN mkdir "/build"

# Copy the source code from your host machine to the container's working directory
COPY . .

# Expose the port on which the service will run
EXPOSE 5001

# Install necessary Go packages for the app
RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

# Switch to the non-root user for security reasons
USER 10001

# Define the entrypoint to use CompileDaemon for automatic rebuilds and serve the app
ENTRYPOINT ["CompileDaemon", "-build=go build -o build/jiotv_go .", "-command=build/jiotv_go serve --public", "-include=*.html"]
