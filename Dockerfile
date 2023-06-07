# Base image with Docker client
FROM docker

# Install Docker Compose
RUN apk add --no-cache py-pip
RUN pip install docker-compose
