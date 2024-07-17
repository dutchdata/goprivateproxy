# goprivateproxy
[![GoDoc](https://pkg.go.dev/badge/github.com/dutchdata/goprivateproxy.svg)](https://pkg.go.dev/github.com/dutchdata/goprivateproxy)

A Go-based reverse proxy and AWS PrivateLink replacement running as a systemd service on EC2 instances. Includes rate limiting, bot blocking, and dynamic routing based on subdomains and paths.

## Requirements
- Go 1.20 or later

## Features

- Reverse proxy for handling incoming requests
- Rate limiting to prevent abuse
- Bot blocking based on user-agent strings
- Dynamic routing based on subdomains and paths
- Configurable via a YAML configuration file

## Installation

1. **Clone the repository**:
    ```sh
    git clone https://github.com/dutchdata/goprivateproxy.git
    cd goprivateproxy
    ```

2. **Create a configuration file**:
    ```sh
    cp config.yaml.example config.yaml
    ```

3. **Edit config.yaml according to your subnet and requirements:**

    The configuration file supports the following fields:

    - `port`: The port on which the proxy server listens.
    - `limiter`: Rate limiting configuration, with `rps` (requests per second) and `burst` values.
    - `botBlockList`: A list of user-agent substrings to block.
    - `permittedBots`: A list of user-agent substrings to allow.
    - `otherRoutes`: A list of routes for subdomains and paths.
    - `ip`: Target IP address.
    - `port`: Target port.
    - `path`: Path or subdomain for routing.
    - `defaultRoute`: Default route configuration.
    - `ip`: Default target IP address.
    - `port`: Default target port.


## Getting Started

**Fetch config and create the server**

```go
package main

import "github.com/dutchdata/goprivateproxy"

func main() {
	config := GetConfig()
	server := NewServer(config)
	server.Start()
}
```

**Build the binary**

```sh
go build
```

**Run to test your configuration**

```sh
./goprivateproxy -config config.yaml
```

**Set up as a systemd service** 

`/etc/systemd/system/goprivateproxy.service`

```ini
[Unit]
Description=Go Proxy Service

[Service]
ExecStart=/home/ec2-user/goprivateproxy/goprivateproxy -config /home/ec2-user/goprivateproxy/config.yaml
Restart=always
User=ec2-user
Group=ec2-user
WorkingDirectory=/home/ec2-user/goprivateproxy/

[Install]
WantedBy=multi-user.target
```

**Start the service**

```sh
sudo systemctl daemon-reload
sudo systemctl enable goprivateproxy
sudo systemctl start goprivateproxy
```

**Check logs**
```sh
journalctl -u goprivateproxy.service -f
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
