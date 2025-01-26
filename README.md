# ipx
Origin IP Discovery Tool

## Installation
Run the following command to install the latest version:
```bash
go install -v github.com/0xAFz/ipx@latest
```
#### Usage
```bash
Origin IP Discovery Tool

Usage:
  ipx [flags]

Flags:
      --cidr string     CIDR range to scan for IPs (e.g., 192.168.1.0/24)
      --delta int       Allowed range for content length difference
      --domain string   Domain name to use in the Host header
  -h, --help            help for ipx
```
```bash
ipx --cidr 192.168.1.0/24 --domain example.com
```
