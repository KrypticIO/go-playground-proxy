# Go Playground Proxy

A simple proxy service that enables [Zulip code playgrounds](https://zulip.com/help/code-blocks#code-playgrounds) to work with the official [Go Playground](https://go.dev/play/).

## Why This Exists

Zulip's code playground feature expects a URL template with a `{code}` parameter, but the official Go Playground doesn't support direct URL parameters. Instead, it requires uploading code via an API to get a shareable link.

This proxy bridges that gap by:
1. Accepting code via URL parameter from Zulip
2. Uploading it to the official Go Playground
3. Redirecting users to the generated playground URL

## Quick Start

### Using Docker (Recommended)

```bash
# Pull and run from GitHub Container Registry
docker run -d -p 8080:8080 ghcr.io/yourusername/go-playground-proxy:latest
```

### Using Docker Compose

```yaml
services:
  go-playground-proxy:
    image: ghcr.io/yourusername/go-playground-proxy:latest
    ports:
      - "8080:8080"
    restart: unless-stopped
```

### Building from Source

```bash
git clone https://github.com/yourusername/go-playground-proxy.git
cd go-playground-proxy
go run main.go
```

## Zulip Configuration

1. Go to your Zulip organization settings → **Code playgrounds**
2. Add a new playground with:
    - **Language:** `Go`
    - **Name:** `Go Playground`
    - **URL template:** `http://your-proxy-server:8080/?code={code}`

## Usage

Once configured, Go code blocks in Zulip will show a playground button on hover:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

Clicking the button opens the code in the official Go Playground.

## API Endpoints

- `GET /?code={encoded_code}` - Main proxy endpoint
- `GET /health` - Health check endpoint

## Features

- ✅ Uses official Go Playground infrastructure
- ✅ Proper URL decoding and error handling
- ✅ Health check endpoint for monitoring
- ✅ Lightweight Docker image (~10MB)
- ✅ Multi-architecture support (amd64, arm64)
- ✅ Comprehensive logging

## Development

### Prerequisites

- Go 1.24 or later
- Docker (optional)

### Running locally

```bash
go mod download
go run main.go
```

The service will be available at `http://localhost:8080`.

### Testing

```bash
# Test with a simple Go program
curl "http://localhost:8080/?code=package%20main%0A%0Aimport%20%22fmt%22%0A%0Afunc%20main()%20%7B%0A%20%20%20%20fmt.Println(%22Hello,%20World!%22)%0A%7D"
```

### Building Docker image

```bash
docker build -t go-playground-proxy .
```

## Deployment Options

### 1. Cloud Run (Google Cloud)
```bash
gcloud run deploy go-playground-proxy \
  --image ghcr.io/yourusername/go-playground-proxy:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

### 2. Railway
```bash
railway login
railway link
railway up
```

### 3. Fly.io
```bash
fly deploy
```

### 4. Self-hosted
Use the provided `docker-compose.yaml` or deploy directly with Docker.

## Configuration

Environment variables:

- `PORT` - Server port (default: 8080)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Go Playground](https://go.dev/play/) team for the excellent service
- [Zulip](https://zulip.com/) for the code playground feature