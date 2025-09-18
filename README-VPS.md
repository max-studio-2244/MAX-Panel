# MAX Panel VPS - Full Featured Game Server Management

A complete, production-ready game server management panel that runs on any VPS with full Docker support, real console access, file management, and custom domain connectivity.

## 🚀 Features

### ✅ Fully Working Features
- **Real Game Servers** - Actually runs Minecraft, Node.js, and other game servers
- **Live Console** - Real-time server console with command execution
- **File Manager** - Browse, edit, upload, download server files
- **Backup System** - Create and restore server backups
- **User Management** - Multi-user support with authentication
- **Custom Domain** - Connect your own domain with SSL
- **Node System** - Connect multiple VPS nodes
- **Resource Monitoring** - Real CPU, memory, disk usage
- **WebSocket Console** - Real-time log streaming
- **Security** - File access controls, user permissions

### 🎮 Supported Games
- **Minecraft** (Java Edition) - Full server with plugins
- **Node.js Applications** - Custom web/game servers
- **Extensible** - Easy to add more game types

## 🛠️ Installation

### Quick Install (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/maxpanel/installer/main/install-vps.sh | sudo bash -s yourdomain.com
```

### Manual Installation
```bash
# Clone repository
git clone https://github.com/maxpanel/panel.git
cd panel

# Build and start
docker-compose -f docker-compose.vps.yml up -d
```

### Custom Domain Setup
```bash
# Install with your domain
curl -sSL https://install.maxpanel.com | sudo bash -s panel.yourdomain.com

# Or set domain after installation
docker-compose -f docker-compose.vps.yml down
echo "DOMAIN=panel.yourdomain.com" >> .env
docker-compose -f docker-compose.vps.yml up -d
```

## 🌐 Domain Configuration

### DNS Setup
Point your domain to your VPS IP:
```
A    panel.yourdomain.com    -> YOUR_VPS_IP
```

### SSL Certificate
The installer automatically generates self-signed certificates. For production, use Let's Encrypt:
```bash
# Install certbot
apt install certbot

# Get certificate
certbot certonly --standalone -d panel.yourdomain.com

# Copy certificates
cp /etc/letsencrypt/live/panel.yourdomain.com/fullchain.pem /opt/maxpanel/ssl/cert.pem
cp /etc/letsencrypt/live/panel.yourdomain.com/privkey.pem /opt/maxpanel/ssl/key.pem

# Restart panel
docker-compose -f /opt/maxpanel/docker-compose.vps.yml restart nginx
```

## 🔗 Node System

### Add External Nodes
1. Go to Admin → Nodes
2. Click "Add Node"
3. Enter node details
4. Run the provided install command on your new VPS

### Node Installation Command
```bash
curl -sSL https://raw.githubusercontent.com/maxpanel/node/main/install.sh | bash -s -- \
  --token=YOUR_NODE_TOKEN \
  --host=panel.yourdomain.com \
  --port=8080
```

## 📁 File Structure
```
/opt/maxpanel/
├── panel                    # Main binary
├── web/                     # Web interface
├── servers/                 # Game server files
│   ├── server-id-1/        # Individual server directories
│   └── server-id-2/
├── backups/                 # Server backups
├── logs/                    # Panel logs
├── ssl/                     # SSL certificates
├── docker-compose.vps.yml   # Docker configuration
└── .env                     # Environment variables
```

## 🎮 Creating Game Servers

### Minecraft Server
1. Click "Create Server"
2. Select "Minecraft"
3. Set port (25565), memory (2GB recommended)
4. Server will download and start automatically
5. Access console for real-time management

### Node.js Server
1. Click "Create Server"
2. Select "Node.js"
3. Upload your application files
4. Edit server.js as needed
5. Start server and access via assigned port

## 🔧 Management

### Start/Stop Panel
```bash
cd /opt/maxpanel
docker-compose -f docker-compose.vps.yml up -d    # Start
docker-compose -f docker-compose.vps.yml down     # Stop
docker-compose -f docker-compose.vps.yml restart  # Restart
```

### View Logs
```bash
docker-compose -f docker-compose.vps.yml logs -f maxpanel
```

### Update Panel
```bash
curl -sSL https://install.maxpanel.com | sudo bash
```

### Backup Panel Data
```bash
tar -czf maxpanel-backup-$(date +%Y%m%d).tar.gz /opt/maxpanel
```

## 🔒 Security

### Default Credentials
- Username: `admin`
- Password: `admin123`
- **Change immediately after installation!**

### Firewall Setup
```bash
# Allow panel access
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 8080/tcp

# Allow game server ports
ufw allow 25565/tcp  # Minecraft
ufw allow 27015/tcp  # Source games
ufw allow 3000:3010/tcp  # Custom games

ufw enable
```

### User Management
- Create users via Admin panel
- Assign servers to specific users
- Role-based permissions (Admin/User)

## 📊 Monitoring

### Resource Usage
- Real-time CPU, memory, disk monitoring
- Per-server resource tracking
- Historical usage graphs

### Server Status
- Live server status indicators
- Automatic restart on crash
- Performance alerts

## 🆘 Troubleshooting

### Panel Won't Start
```bash
# Check Docker status
systemctl status docker

# Check logs
docker-compose -f docker-compose.vps.yml logs

# Restart services
docker-compose -f docker-compose.vps.yml restart
```

### Game Server Issues
```bash
# Check server logs via web console or:
docker-compose -f docker-compose.vps.yml exec maxpanel cat /app/servers/SERVER_ID/server.log

# Check port availability
netstat -tulpn | grep :25565
```

### Domain/SSL Issues
```bash
# Test domain resolution
nslookup panel.yourdomain.com

# Check SSL certificate
openssl x509 -in /opt/maxpanel/ssl/cert.pem -text -noout

# Restart nginx
docker-compose -f docker-compose.vps.yml restart nginx
```

## 🔄 Updates

The panel supports automatic updates:
```bash
# Check for updates
curl -s https://api.maxpanel.com/version

# Update to latest
curl -sSL https://install.maxpanel.com | sudo bash
```

## 📞 Support

- Documentation: https://docs.maxpanel.com
- Issues: https://github.com/maxpanel/panel/issues
- Discord: https://discord.gg/maxpanel

---

**Production Ready** • **Docker Native** • **Multi-Node** • **Custom Domain** • **Real Game Servers**