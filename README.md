# HellCat

A high–intensity VLESS pentesting-testing tool using xray-core and Go.

Use only in pentest your servers!

## Features

- Launch multiple xray-core instances automatically  
- Run hundreds of parallel HTTP download streams via SOCKS5  
- Config generation from VLESS links  
- Supports single `--url` or multiple via `--list`  

## Prerequisites

- Linux (tested on Ubuntu 22.04+)  
- `bash`, `curl`, `wget`, `unzip`  
- `git`  
- `sudo` privileges  

## Installation

1. **Clone the repo**  
   ```bash
   git clone https://your.repo/HellCat.git
   cd HellCat
2. **Run the installer**  
   ```bash
    chmod +x install.sh
    ./install.sh
3. **Build the HellCat binary** 
   ```bash 
    go mod init hellcat
    go mod tidy
    go build -o hellcat main.go
