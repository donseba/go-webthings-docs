FROM alpine:latest

# Install timezone data
RUN apk --no-cache add tzdata

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled binary from the current directory into the container
COPY . .

# Expose port (for example, 8080) for the application
EXPOSE 8080

# Command to run the application
ENTRYPOINT ["./website"]