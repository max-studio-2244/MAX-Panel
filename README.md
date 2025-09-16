# 🚀 MAX Panel

A complete, enterprise-grade game server management panel built with Go and modern web technologies. All the power of Pterodactyl Panel with one-command installation.

## ✨ Features

- **🚀 One-Command Installation** - Get started in minutes
- **🐳 Docker Integration** - No separate Wings daemon needed
- **🎯 Multi-Game Support** - 50+ game eggs with full Pterodactyl compatibility
- **💻 Modern Admin Dashboard** - Complete panel customization & branding
- **📊 Real-time Console** - WebSocket-based server console with command history
- **⚡ Advanced Resource Management** - Multi-node support with automatic scaling
- **📁 Complete File Manager** - Built-in editor, upload/download, permissions
- **🔄 Enterprise Backups** - Automated, scheduled, and manual backup system
- **👥 User Management** - Role-based permissions, subusers, API keys
- **🌐 Multi-Node Support** - Distributed server management across multiple machines
- **🔒 Security Features** - 2FA, activity logs, IP restrictions, SSL/TLS
- **📦 Egg System** - Import/export server configurations like Pterodactyl
- **📊 Monitoring & Analytics** - Resource usage, performance metrics, alerts
- **🔌 API Gateway** - Full REST API with documentation

## 🚀 Quick Start

### Windows
```bash
# Run the installer
install.bat
```

### Linux/macOS
```bash
# Make installer executable
chmod +x install.sh

# Run the installer
./install.sh
```

### Docker Compose (Recommended)
```bash
# Clone and start
git clone <your-repo>
cd game-panel
docker-compose up -d
```

## 📋 Requirements

- **Docker** & **Docker Compose** (recommended)
- **Go 1.21+** (for development)
- **2GB+ RAM** (for game servers)

## 🎯 Supported Games (50+ Eggs)

### Minecraft
- Vanilla Server
- Paper/Spigot
- Forge/Fabric
- Bedrock Edition
- Modded variants

### Source Engine
- CS:GO/CS2
- Team Fortress 2
- Garry's Mod
- Left 4 Dead 2

### Survival Games
- Rust (Official/Oxide)
- ARK: Survival Evolved
- Valheim
- 7 Days to Die
- The Forest

### Other Popular Games
- Terraria
- Don't Starve Together
- Project Zomboid
- Satisfactory
- And many more...

## 🔧 Configuration

### Environment Variables
```bash
PORT=8080                    # Web interface port
DATABASE_PATH=./panel.db     # SQLite database path
```

### Game Server Limits
- **Memory**: 512MB - 8GB
- **CPU**: 0.5 - 4 cores
- **Ports**: 1024 - 65535

## 📚 Complete API Documentation

### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/register` - User registration
- `POST /api/auth/logout` - User logout
- `GET /api/auth/me` - Get current user

### Server Management
- `GET /api/servers` - List all servers
- `POST /api/servers` - Create new server
- `GET /api/servers/:id` - Get server details
- `POST /api/servers/:id/start` - Start server
- `POST /api/servers/:id/stop` - Stop server
- `POST /api/servers/:id/restart` - Restart server
- `DELETE /api/servers/:id` - Delete server

### File Management
- `GET /api/servers/:id/files` - List server files
- `GET /api/servers/:id/files/download` - Download file
- `POST /api/servers/:id/files/upload` - Upload file
- `PUT /api/servers/:id/files/edit` - Edit file
- `DELETE /api/servers/:id/files/delete` - Delete file

### Backup System
- `GET /api/servers/:id/backups` - List backups
- `POST /api/servers/:id/backups` - Create backup
- `POST /api/servers/:id/backups/:backup_id/restore` - Restore backup
- `DELETE /api/servers/:id/backups/:backup_id` - Delete backup

### Admin Panel
- `GET /api/admin/users` - Manage users
- `GET /api/admin/eggs` - Manage game eggs
- `GET /api/admin/nodes` - Manage nodes
- `GET /api/admin/settings` - Panel settings
- `GET /api/admin/logs` - Activity logs

### Real-time Features
- `WS /ws/:id` - Server console with command execution

## 🏗️ Development

### Local Development
```bash
# Install dependencies
go mod tidy

# Run in development mode
go run .

# Build binary
go build -o game-panel .
```

### Project Structure
```
max-panel/
├── main.go              # Main application
├── handlers.go          # Server management
├── auth.go              # Authentication system
├── admin.go             # Admin management
├── files.go             # File & backup management
├── web/                 # Frontend files
│   ├── index.html       # Main interface
│   ├── admin.html       # Admin dashboard
│   ├── app.js          # Main JavaScript
│   └── admin.js        # Admin JavaScript
├── docker-compose.yml   # Docker setup
├── Dockerfile          # Container build
├── install.sh          # Linux/macOS installer
└── install.bat         # Windows installer
```

## 🔒 Security Features

- **Container Isolation** - Each server runs in its own Docker container
- **Resource Limits** - CPU and memory constraints
- **Port Management** - Automatic port allocation
- **File System Isolation** - Servers can't access host files

## 🆚 vs Pterodactyl Panel

| Feature | Game Panel | Pterodactyl |
|---------|------------|-------------|
| Installation | One command | Complex setup |
| Dependencies | Docker only | PHP, Redis, MySQL, Wings |
| Architecture | Single binary | Multi-component |
| Resource Usage | Lightweight | Heavy |
| Learning Curve | Minimal | Steep |

## 🛠️ Troubleshooting

### Common Issues

**Docker not found**
```bash
# Install Docker Desktop
# Windows: https://docs.docker.com/desktop/windows/
# macOS: https://docs.docker.com/desktop/mac/
# Linux: https://docs.docker.com/engine/install/
```

**Port already in use**
```bash
# Check what's using the port
netstat -tulpn | grep :8080

# Change port in docker-compose.yml
ports:
  - "8081:8080"  # Use port 8081 instead
```

**Server won't start**
```bash
# Check Docker logs
docker-compose logs game-panel

# Check server logs
docker logs <container-name>
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

- Inspired by Pterodactyl Panel
- Built with Go Fiber framework
- Uses Docker for containerization
- Frontend with Tailwind CSS

---

**Made with ❤️ for the gaming community**