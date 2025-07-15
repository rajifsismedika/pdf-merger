# PDF Merge Service

A Go-based web service for downloading and merging PDF documents from multiple URLs.

## Features

- Concurrent PDF downloads for improved performance
- Automatic URL encoding and fallback mechanisms
- In-memory PDF merging using pdfcpu
- Graceful error handling and non-PDF content skipping
- RESTful API with JSON responses
- Health monitoring and logging

## Prerequisites

- Go 1.23 or higher
- Network access to download PDFs from external URLs

## Installation

### 1. Clone and Build

```bash
git clone <repository-url>
cd api_merge_report2

# Install dependencies
go mod tidy

# Build the binary
go build -o bin/pdfmerge-server ./cmd/server
```

### 2. Configuration

Create a `.env` file:

```env
PORT=8585
BASE_URL=http://localhost
```

## Running the Service

### Development Mode

```bash
# Run directly from source
go run cmd/server/main.go

# Or run the built binary
./bin/pdfmerge-server

# Run with custom port
./bin/pdfmerge-server -port 9090
```

### Production Deployment

For production environments, it's recommended to use a process supervisor for automatic restarts, monitoring, and service management.

#### Option 1: systemd (Linux - Recommended)

1. **Create service user:**
```bash
sudo useradd --system --no-create-home --shell /bin/false pdfmerge
```

2. **Setup directory structure:**
```bash
sudo mkdir -p /opt/pdfmerge/{bin,logs}
sudo cp bin/pdfmerge-server /opt/pdfmerge/bin/
sudo cp .env /opt/pdfmerge/
sudo chown -R pdfmerge:pdfmerge /opt/pdfmerge
sudo chmod +x /opt/pdfmerge/bin/pdfmerge-server
```

3. **Install systemd service:**
```bash
sudo cp pdfmerge.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable pdfmerge
sudo systemctl start pdfmerge
```

4. **Service management:**
```bash
sudo systemctl status pdfmerge    # Check status
sudo systemctl restart pdfmerge   # Restart service
sudo systemctl stop pdfmerge      # Stop service
sudo journalctl -u pdfmerge -f    # View logs
```

#### Option 2: Supervisor (Cross-platform)

1. **Install supervisor:**
```bash
# Ubuntu/Debian
sudo apt-get install supervisor

# CentOS/RHEL
sudo yum install supervisor
```

2. **Setup configuration:**
```bash
sudo cp supervisor-pdfmerge.conf /etc/supervisor/conf.d/
sudo supervisorctl reread
sudo supervisorctl update
sudo supervisorctl start pdfmerge
```

3. **Service management:**
```bash
sudo supervisorctl status pdfmerge     # Check status
sudo supervisorctl restart pdfmerge    # Restart service
sudo supervisorctl stop pdfmerge       # Stop service
sudo supervisorctl tail pdfmerge       # View logs
```

#### Option 3: Docker

1. **Build and run:**
```bash
# Using docker-compose
docker-compose up -d

# Or with plain Docker
docker build -t pdfmerge-service .
docker run -d --name pdfmerge --restart=always -p 8585:8585 pdfmerge-service
```

## API Usage

### Merge PDFs Endpoint

**POST /merge**

Request body:
```json
{
  "urls": [
    "https://example.com/document1.pdf",
    "https://example.com/document2.pdf"
  ]
}
```

Response:
- Success: Returns merged PDF binary data with `Content-Type: application/pdf`
- Error: Returns JSON error response

### Report Endpoint

**GET /report/{id}**

Returns report data for the specified ID.

## Testing

### Using curl

Create `curl-format.txt` file:
```txt
   time_namelookup:  %{time_namelookup}s\n
        time_connect:  %{time_connect}s\n
     time_appconnect:  %{time_appconnect}s\n
    time_pretransfer:  %{time_pretransfer}s\n
       time_redirect:  %{time_redirect}s\n
  time_starttransfer:  %{time_starttransfer}s\n
                     ----------\n
          time_total:  %{time_total}s\n
```

Test commands:
```bash
curl -w "@curl-format.txt" -o /dev/null http://localhost:8585/report/1
curl -w "@curl-format.txt" -o /dev/null http://localhost:8585/report/1000
```

### Using the PHP Client

See the `php-client/` directory for a PHP implementation example.

## Build Options

### Standard Build
```bash
go build -o bin/pdfmerge-server ./cmd/server
```

### Static Binary (for containers/Alpine)
```bash
CGO_ENABLED=0 go build -o bin/pdfmerge-server ./cmd/server
```

### Cross-platform Builds
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/pdfmerge-server-linux ./cmd/server

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/pdfmerge-server.exe ./cmd/server

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/pdfmerge-server-macos ./cmd/server
```

## Command Line Options

```bash
./bin/pdfmerge-server -h
```

Available options:
- `-port string`: Port to run the server on (overrides PORT env var)

## Configuration Priority

1. Command line flags (highest priority)
2. Environment variables
3. `.env` file values
4. Default values (lowest priority)

## Troubleshooting

### Common Issues

1. **Port already in use**: Change the port using `-port` flag or PORT environment variable
2. **Permission denied**: Ensure the binary has execute permissions (`chmod +x`)
3. **Service fails to start**: Check logs using `journalctl -u pdfmerge` (systemd) or `supervisorctl tail pdfmerge`

### Logs

- **systemd**: `sudo journalctl -u pdfmerge -f`
- **Supervisor**: `sudo supervisorctl tail pdfmerge`
- **Docker**: `docker logs pdfmerge`

## Security Considerations

The systemd service configuration includes security hardening:
- Runs as non-root user
- Private temporary directories
- Protected file system access
- No new privileges allowed

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request