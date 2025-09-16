#!/bin/bash

echo "🚀 MAX Panel Installer"
echo "======================"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    echo "Visit: https://docs.docker.com/compose/install/"
    exit 1
fi

# Check if Go is installed (for local development)
if command -v go &> /dev/null; then
    echo "✅ Go is installed"
    GO_VERSION=$(go version | cut -d' ' -f3)
    echo "   Version: $GO_VERSION"
else
    echo "⚠️  Go is not installed (only needed for development)"
fi

echo ""
echo "Choose installation option:"
echo "1) Dependencies  - Check & install system requirements"
echo "2) Panel        - Install main panel with domain setup"
echo "3) Wings        - Install Wings daemon with node token"
echo "4) Cloudflare   - Setup Cloudflare Zero Trust tunnel"
echo "5) System       - Check system information & compatibility"
echo "6) Uninstall    - Remove all components"
echo "7) Exit         - Quit installer"

read -p "Enter your choice (1-7): " choice

case $choice in
    1)
        echo "🔍 Checking dependencies..."
        
        # Check Docker
        if ! command -v docker &> /dev/null; then
            echo "❌ Docker not found. Installing Docker..."
            curl -fsSL https://get.docker.com -o get-docker.sh
            sudo sh get-docker.sh
            sudo usermod -aG docker $USER
            echo "✅ Docker installed. Please log out and back in."
        else
            echo "✅ Docker found"
        fi
        
        # Check Docker Compose
        if ! command -v docker-compose &> /dev/null; then
            echo "❌ Docker Compose not found. Installing..."
            sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
            sudo chmod +x /usr/local/bin/docker-compose
            echo "✅ Docker Compose installed"
        else
            echo "✅ Docker Compose found"
        fi
        
        # Check Go (optional)
        if command -v go &> /dev/null; then
            echo "✅ Go found: $(go version)"
        else
            echo "⚠️ Go not found (optional for development)"
        fi
        
        echo "✅ Dependencies check complete!"
        ;;
        
    2)
        echo "🚀 Installing MAX Panel..."
        
        # Get domain
        read -p "Enter panel domain (e.g., panel.yourdomain.com): " panel_domain
        
        # Create data directory
        mkdir -p data
        
        # Build and start panel
        docker-compose up -d --build
        
        echo "✅ MAX Panel is starting!"
        echo "🌐 Access it at: http://$panel_domain or http://localhost:8080"
        
        # Create admin user
        echo ""
        echo "Creating admin user..."
        read -p "Email: " admin_email
        read -p "Username: " admin_username
        read -s -p "Password: " admin_password
        echo
        read -p "First Name: " admin_first_name
        read -p "Last Name: " admin_last_name
        
        echo "✅ Admin user will be created on first panel access"
        echo "📊 View logs: docker-compose logs -f"
        ;;
        
    3)
        echo "🛫 Installing Wings daemon..."
        
        # Get domain and token
        read -p "Enter Wings domain (e.g., node1.yourdomain.com): " wings_domain
        read -p "Enter node token from panel: " node_token
        
        # Download Wings
        echo "Downloading Wings..."
        curl -L -o /usr/local/bin/wings "https://github.com/pterodactyl/wings/releases/latest/download/wings_linux_$([[ "$(uname -m)" == "x86_64" ]] && echo "amd64" || echo "arm64")"
        chmod u+x /usr/local/bin/wings
        
        # Create Wings config
        mkdir -p /etc/pterodactyl
        cat > /etc/pterodactyl/config.yml << EOF
debug: false
api:
  host: 0.0.0.0
  port: 8080
  ssl:
    enabled: false
  upload_limit: 100
token: $node_token
system:
  data: /var/lib/pterodactyl/volumes
  sftp:
    bind_port: 2022
allowed_mounts: []
remote: https://$wings_domain
EOF
        
        # Create systemd service
        cat > /etc/systemd/system/wings.service << EOF
[Unit]
Description=Pterodactyl Wings Daemon
After=docker.service
Requires=docker.service
PartOf=docker.service

[Service]
User=root
WorkingDirectory=/etc/pterodactyl
LimitNOFILE=4096
PIDFile=/var/run/wings/daemon.pid
ExecStart=/usr/local/bin/wings
Restart=on-failure
StartLimitInterval=180
StartLimitBurst=30
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
        
        # Start Wings
        systemctl enable --now wings
        
        echo "✅ Wings daemon installed and started!"
        echo "📊 Check status: systemctl status wings"
        ;;
        
    4)
        echo "☁️ Setting up Cloudflare tunnel..."
        
        # Download cloudflared
        echo "Downloading Cloudflare tunnel..."
        curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o cloudflared
        chmod +x cloudflared
        sudo mv cloudflared /usr/local/bin/
        
        echo ""
        echo "Setup Instructions:"
        echo "1. Go to Cloudflare Zero Trust Dashboard"
        echo "2. Navigate to Access > Tunnels"
        echo "3. Create a new tunnel"
        echo "4. Copy the tunnel token"
        echo "5. Configure these routes:"
        echo "   - panel.yourdomain.com → localhost:8080"
        echo "   - node1.yourdomain.com → localhost:8080"
        echo ""
        
        read -p "Enter your tunnel token: " tunnel_token
        
        # Install tunnel service
        cloudflared service install $tunnel_token
        
        echo "✅ Cloudflare tunnel configured!"
        echo "🔒 Your services are now securely accessible through Cloudflare"
        ;;
        
    5)
        echo "📊 Checking system information..."
        
        echo ""
        echo "=== SYSTEM INFORMATION ==="
        echo "OS: $(uname -s) $(uname -r)"
        echo "Architecture: $(uname -m)"
        echo "Hostname: $(hostname)"
        echo "Uptime: $(uptime -p 2>/dev/null || uptime)"
        
        echo ""
        echo "=== HARDWARE INFORMATION ==="
        echo "CPU: $(grep 'model name' /proc/cpuinfo | head -1 | cut -d':' -f2 | xargs 2>/dev/null || echo 'Unknown')"
        echo "CPU Cores: $(nproc 2>/dev/null || echo 'Unknown')"
        echo "Memory: $(free -h | grep '^Mem:' | awk '{print $2}' 2>/dev/null || echo 'Unknown')"
        echo "Disk Space: $(df -h / | tail -1 | awk '{print $2}' 2>/dev/null || echo 'Unknown')"
        
        echo ""
        echo "=== SOFTWARE VERSIONS ==="
        if command -v docker &> /dev/null; then
            echo "Docker: $(docker --version | cut -d' ' -f3 | cut -d',' -f1)"
        else
            echo "Docker: Not installed"
        fi
        
        if command -v docker-compose &> /dev/null; then
            echo "Docker Compose: $(docker-compose --version | cut -d' ' -f3 | cut -d',' -f1)"
        else
            echo "Docker Compose: Not installed"
        fi
        
        if command -v go &> /dev/null; then
            echo "Go: $(go version | cut -d' ' -f3)"
        else
            echo "Go: Not installed"
        fi
        
        echo ""
        echo "=== NETWORK INFORMATION ==="
        echo "IP Address: $(curl -s ifconfig.me 2>/dev/null || echo 'Unable to detect')"
        echo "Local IP: $(hostname -I | awk '{print $1}' 2>/dev/null || echo 'Unknown')"
        
        echo ""
        echo "=== COMPATIBILITY CHECK ==="
        
        # Check minimum requirements
        memory_gb=$(free -g | grep '^Mem:' | awk '{print $2}' 2>/dev/null || echo 0)
        if [ "$memory_gb" -ge 2 ]; then
            echo "✅ Memory: ${memory_gb}GB (Minimum 2GB required)"
        else
            echo "❌ Memory: ${memory_gb}GB (Minimum 2GB required)"
        fi
        
        disk_gb=$(df -BG / | tail -1 | awk '{print $4}' | sed 's/G//' 2>/dev/null || echo 0)
        if [ "$disk_gb" -ge 10 ]; then
            echo "✅ Disk Space: ${disk_gb}GB available (Minimum 10GB required)"
        else
            echo "❌ Disk Space: ${disk_gb}GB available (Minimum 10GB required)"
        fi
        
        if command -v docker &> /dev/null; then
            echo "✅ Docker: Installed"
        else
            echo "❌ Docker: Not installed (Required)"
        fi
        
        echo ""
        echo "✅ System information check complete!"
        ;;
        
    6)
        echo "🗑️ Uninstalling MAX Panel..."
        
        # Stop services
        docker-compose down -v
        
        # Remove containers and images
        docker system prune -af
        
        # Remove Wings if installed
        if systemctl is-active --quiet wings; then
            systemctl stop wings
            systemctl disable wings
            rm -f /etc/systemd/system/wings.service
            rm -f /usr/local/bin/wings
            rm -rf /etc/pterodactyl
        fi
        
        # Remove Cloudflare tunnel
        if command -v cloudflared &> /dev/null; then
            cloudflared service uninstall
            rm -f /usr/local/bin/cloudflared
        fi
        
        # Remove data
        read -p "Remove all data? (y/N): " remove_data
        if [[ $remove_data =~ ^[Yy]$ ]]; then
            rm -rf data/
            rm -f panel.db
        fi
        
        echo "✅ MAX Panel uninstalled successfully!"
        ;;
        
    7)
        echo "Goodbye!"
        exit 0
        ;;
        
    *)
        echo "❌ Invalid choice"
        exit 1
        ;;
esac

echo ""
echo "🎉 Installation complete!"
echo ""
echo "📚 Quick Start Guide:"
echo "1. Access the web interface"
echo "2. Click 'New Server' to create a game server"
echo "3. Select your game type (Minecraft, CS:GO, etc.)"
echo "4. Configure resources and click 'Create'"
echo "5. Start your server and enjoy!"
echo ""
echo "🔧 Troubleshooting:"
echo "- Make sure Docker is running"
echo "- Check firewall settings for the ports you use"
echo "- View logs for any errors"
echo ""
echo "📖 Documentation: https://github.com/your-repo/game-panel"