##Build stage 1
#FROM golang:1.24-alpine AS gobuilder
#
## Set working directory inside container
#WORKDIR /app
#
## Copy the local Go file into the container
#COPY . .
#
## Build the Go application
#RUN go build -o server main.go
#
## Use a minimal image to run the binary
#FROM alpine:latest
#
## Copy the built binary from the builder stage
#COPY --from=gobuilder /app/server /server
#COPY --from=gobuilder /app/Cert /Cert
#
## Expose the port your server listens on (adjust if needed)
##EXPOSE 8080
#EXPOSE 8090
#
## Command to run the binary
#ENTRYPOINT ["/server"]



#Build stage 2
FROM heroiclabs/nakama-pluginbuilder:latest AS builder

ENV GO111MODULE=on
ENV CGO_ENABLED=1

WORKDIR /backend
COPY . .

RUN go version  # Optional: to verify Go version inside the image
RUN go build --trimpath --buildmode=plugin -o ./backend.so

# Final stage
FROM heroiclabs/nakama:latest

# Create a non-root user (e.g., nakamauser) and set it as the default user
RUN useradd -ms /bin/bash nakamauser

# Set up working directories and ensure the user has the necessary permissions
COPY --from=builder /backend/backend.so /nakama/data/modules/
COPY --from=builder /backend/local.yml /nakama/data/
COPY --from=builder /backend/*.json /nakama/data/modules/

COPY --from=builder /backend/Cert /Cert

# Adjust the ownership of copied files to the nakamauser
RUN chown -R nakamauser:nakamauser /nakama/data

# Switch to the non-root user
USER nakamauser

# Run Nakama as the non-root user
CMD ["nakama", "migration", "up"]