#!/bin/bash

# MAX Panel VPS Installer
echo "🚀 Installing MAX Panel VPS..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (use sudo)"
    exit 1
fi

# Update system
apt update && apt upgrade -y

# Install Docker
if ! command -v docker &> /dev/null; then
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    systemctl enable docker
    systemctl start docker
fi

# Install Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "Installing Docker Compose..."
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

# Create installation directory
mkdir -p /opt/maxpanel
cd /opt/maxpanel

# Download panel files
echo "Downloading MAX Panel files..."
curl -L https://github.com/maxpanel/releases/latest/download/maxpanel-vps.tar.gz | tar -xz

# Set permissions
chmod +x install-vps.sh

# Create environment file
cat > .env << EOF
DOMAIN=${1:-localhost}
PANEL_NAME=MAX Panel VPS
PORT=8080
EOF

# Generate SSL certificates (self-signed for now)
mkdir -p ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout ssl/key.pem \
    -out ssl/cert.pem \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=${1:-localhost}"

# Start services
echo "Starting MAX Panel..."
docker-compose -f docker-compose.vps.yml up -d

# Wait for services to start
sleep 10

# Show status
docker-compose -f docker-compose.vps.yml ps

echo ""
echo "✅ MAX Panel VPS installed successfully!"
echo ""
echo "🌐 Access your panel at:"
echo "   HTTP:  http://${1:-localhost}"
echo "   HTTPS: https://${1:-localhost}"
echo ""
echo "🔑 Default login:"
echo "   Username: admin"
echo "   Password: admin123"
echo ""
echo "📝 Configuration files:"
echo "   Panel: /opt/maxpanel"
echo "   Servers: /opt/maxpanel/servers"
echo "   Backups: /opt/maxpanel/backups"
echo ""
echo "🔧 Management commands:"
echo "   Start:   docker-compose -f /opt/maxpanel/docker-compose.vps.yml up -d"
echo "   Stop:    docker-compose -f /opt/maxpanel/docker-compose.vps.yml down"
echo "   Logs:    docker-compose -f /opt/maxpanel/docker-compose.vps.yml logs -f"
echo "   Update:  curl -sSL https://install.maxpanel.com | bash"
echo ""