# 🚀 MAX Panel - Game Server Management

Complete game server management panel with egg system, Cloudflare integration, and multi-user support.

## 🚀 Quick Start

### CodeSandbox (Demo)
[![Open in CodeSandbox](https://codesandbox.io/static/img/play-codesandbox.svg)](https://codesandbox.io/s/github/YOUR_USERNAME/max-panel)

### VPS Installation
```bash
curl -sSL https://install.maxpanel.com | sudo bash -s panel.yourdomain.com
```

### Local Development
```bash
git clone https://github.com/YOUR_USERNAME/max-panel.git
cd max-panel
go mod tidy
go run *.go
```

## 🔑 Default Login
- Username: `admin`
- Password: `admin123`

## ✨ Features
- 🎮 Real game server management
- 🥚 Advanced egg system with versions
- 👥 Multi-user support with assignments
- 🌐 Cloudflare domain integration
- 📁 Complete file manager
- 💾 Backup system
- 🖥️ Real-time console
- 📊 Resource monitoring

## 📚 Documentation
See [README-COMPLETE.md](README-COMPLETE.md) for full documentation.

## 🛠️ Tech Stack
- **Backend:** Go + Fiber
- **Database:** SQLite
- **Frontend:** HTML + Tailwind CSS + JavaScript
- **WebSocket:** Real-time console
- **Docker:** Production deployment