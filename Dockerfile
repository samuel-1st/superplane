# Use Node base image (change if your stack is different)
FROM node:20

# Set working directory
WORKDIR /app

# Install make
RUN apt-get update && apt-get install -y make && rm -rf /var/lib/apt/lists/*

# Copy project files
COPY . .

# Run setup
RUN make dev.setup

# Expose app port
EXPOSE 8000

# Start app
CMD ["make", "dev.start"]
