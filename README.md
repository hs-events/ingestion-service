# Ingestion Service - Hackathon Challenge

This is your team's ingestion service for the hackathon. Your goal is to optimize this service to handle high-volume event traffic from the control server while maintaining low latency and high reliability.

## Prerequisites

### Required Tools

Install the following tools before getting started:

**1. kubectl**
- Included with gcloud: `gcloud components install kubectl`
- Or standalone: `brew install kubectl` (macOS), `apt-get install kubectl` (Linux)

**2. Docker**
- macOS: `brew install --cask docker`
- Linux: `apt-get install docker.io docker-compose`
- Windows: `winget install Docker.DockerDesktop`

**3. Go**
- Version 1.21 or higher
- macOS: `brew install go`
- Linux: `apt-get install golang-go`
- Windows: Download from [golang.org](https://golang.org/dl/)

**4. Docker Compose**
- macOS: Included with Docker Desktop
- Linux: `apt-get install docker-compose`
- Windows: Included with Docker Desktop

**4. Teleport Connect**
- [Download based on your OS (Version = 15.5.4)](https://goteleport.com/download/client-tools/?version=15.5.4)
- Access the cluster https://k8s.hs-events.com/
- Username/Password will be shared through discord

### Verify Installation

```bash
gcloud --version
kubectl version --client
go version
docker --version
docker-compose --version
```