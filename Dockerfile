# Build stage
FROM golang:1.24 AS builder

WORKDIR /build

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app

# Final stage - use Ubuntu instead of Alpine for better Chrome support
FROM ubuntu:22.04

# Install basic dependencies first
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    wget \
    gnupg \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

# Add Chrome repository BEFORE trying to install chrome
RUN wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | apt-key add - \
    && echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list

# Now install Chrome and all other dependencies
RUN apt-get update && apt-get install -y \
    google-chrome-stable \
    fonts-liberation \
    libappindicator3-1 \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcups2 \
    libdrm2 \
    libgbm1 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libxcomposite1 \
    libxdamage1 \
    libxrandr2 \
    libxss1 \
    xdg-utils \
    libgconf-2-4 \
    && rm -rf /var/lib/apt/lists/*

# Create app directory
WORKDIR /app

# Copy built application
COPY --from=builder /build/app /app/app

# Add non-root user for security
RUN useradd -m appuser && chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8080

# Ensure the app is executable
RUN chmod +x /app/app

# Run the application directly
ENTRYPOINT ["/app/app"]