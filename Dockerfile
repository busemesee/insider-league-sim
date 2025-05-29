FROM golang:1.20-alpine

# Ensure git is available for module fetching
RUN apk update && apk add --no-cache git

WORKDIR /app

# Copy entire project
COPY . .

# Fetch dependencies and generate go.sum
RUN go mod tidy

# Build the application
RUN go build -o insider-league-simulation cmd/server/main.go

EXPOSE 8080
CMD ["./insider-league-simulation"]
