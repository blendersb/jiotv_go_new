FROM golang:1.25-alpine

ENV GO111MODULE=on \
    GOEXPERIMENT=jsonv2,greenteagc \
    JIOTV_DEBUG=true \
    JIOTV_PATH_PREFIX="/app/.jiotv_go"
RUN useradd -u 10001 -m user

WORKDIR /app

RUN mkdir "/build"

# Copy source files from the host computer to the container
COPY . .
# Expose your port
EXPOSE 5001

# Switch to non-root user
USER 10001
RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon -build="go build -o build/jiotv_go ." -command="build/jiotv_go serve --public" -include="*.html"
