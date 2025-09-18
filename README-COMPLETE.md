# 🚀 MAX Panel VPS - Complete Production System

**The most advanced game server management panel with full egg system, Cloudflare integration, and multi-user support.**

## ✅ **Complete Feature Set**

### 🎮 **Game Server Management**
- **Real Server Execution** - Actually runs game servers with processes
- **Live Console** - Real-time WebSocket console with command execution
- **Resource Monitoring** - CPU, memory, disk usage tracking
- **Auto-restart** - Automatic server restart on crash
- **Port Management** - Automatic port allocation and conflict detection

### 🥚 **Advanced Egg System**
- **Custom Eggs** - Create eggs for any game/application
- **Version Control** - Set specific game versions and build numbers
- **Install Scripts** - Automated server setup and file downloads
- **Config Templates** - Dynamic configuration file generation
- **Environment Variables** - Customizable server environment
- **Memory Limits** - Min/max memory constraints per egg
- **Multi-Category** - Organize eggs by game categories

### 👥 **User Management & Assignments**
- **Multi-User Support** - Admin and regular user roles
- **Server Assignments** - Assign servers to specific users
- **Role-Based Access** - Owner, admin, user permissions
- **User Dashboard** - Personalized server view per user

### 🌐 **Cloudflare Integration**
- **Automatic DNS** - Creates A and wildcard records
- **SSL Certificates** - Generates Cloudflare Origin certificates
- **DDoS Protection** - Cloudflare proxy protection
- **Custom Domains** - Connect any domain instantly
- **HTTPS Redirect** - Automatic HTTP to HTTPS redirect

### 📁 **File Management**
- **Real File System** - Browse actual server directories
- **File Editor** - Edit configuration files in-browser
- **Upload/Download** - File transfer capabilities
- **Security** - Path traversal protection
- **Permissions** - File permission management

### 💾 **Backup System**
- **ZIP Backups** - Complete server backups in ZIP format
- **Scheduled Backups** - Automatic backup creation
- **Restore Function** - One-click backup restoration
- **Size Tracking** - Backup size monitoring

## 🛠️ **Installation Methods**

### **Method 1: One-Command Install**
```bash
curl -sSL https://install.maxpanel.com | sudo bash -s panel.yourdomain.com
```

### **Method 2: Docker Compose**
```bash
git clone https://github.com/maxpanel/panel.git
cd panel
docker-compose -f docker-compose.vps.yml up -d
```

### **Method 3: Manual Build**
```bash
go mod tidy
go build -o panel vps-*.go
./panel
```

## 🌐 **Cloudflare Setup**

### **Step 1: Get API Token**
1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens)
2. Create Custom Token with Zone:Edit permissions
3. Copy the token

### **Step 2: Find Zone ID**
1. Go to your domain overview in Cloudflare
2. Copy Zone ID from right sidebar

### **Step 3: Configure in Panel**
1. Access Admin → Cloudflare Setup
2. Enter domain, email, API token, and Zone ID
3. Click "Setup Cloudflare"
4. System automatically creates DNS records and SSL certificates

### **What Gets Created:**
- `A` record: `panel.yourdomain.com` → Your VPS IP (Proxied)
- `A` record: `*.panel.yourdomain.com` → Your VPS IP (DNS Only)
- SSL certificate for domain and wildcard
- Nginx reverse proxy configuration
- Automatic HTTPS redirect

## 🥚 **Egg Management**

### **Creating Custom Eggs**
1. Go to Admin → Egg Management
2. Click "Create Egg"
3. Fill in details:
   - **Name**: Display name for the egg
   - **Game**: Game type identifier
   - **Version**: Specific version (e.g., "1.20.1", "latest")
   - **Build Number**: Specific build if applicable
   - **Category**: Organization category
   - **Start Command**: Command with variables like `{MEMORY}`, `{PORT}`
   - **Install Script**: Bash script to download/setup files
   - **Config Files**: JSON object with filename → content mapping
   - **Environment**: JSON object with environment variables
   - **Memory Limits**: Min/max memory constraints

### **Example Minecraft Paper Egg**
```json
{
  "name": "Minecraft Paper 1.20.1",
  "game": "minecraft",
  "version": "1.20.1",
  "build_number": "196",
  "category": "Minecraft",
  "start_command": "java -Xms{MEMORY}M -Xmx{MEMORY}M -jar paper.jar nogui",
  "install_script": "#!/bin/bash\nwget https://api.papermc.io/v2/projects/paper/versions/1.20.1/builds/196/downloads/paper-1.20.1-196.jar -O paper.jar",
  "config_files": {
    "server.properties": "server-port={PORT}\nmotd=MAX Panel Server\nonline-mode=false\ndifficulty=easy",
    "eula.txt": "eula=true"
  },
  "environment": {
    "EULA": "true",
    "TYPE": "PAPER"
  },
  "min_memory": 1024,
  "max_memory": 8192
}
```

### **Creating Servers from Eggs**
1. In Egg Management, click "Create Server" on any egg
2. Set server name, port, memory (within egg limits)
3. Assign to specific user (optional)
4. Server is automatically created with egg configuration

## 👥 **User Assignment System**

### **Assigning Servers to Users**
1. Create server from egg
2. Select user in "Assign to User" dropdown
3. User gets access to manage that server
4. Admin can view all assignments in server details

### **User Roles**
- **Owner**: Full server control, can assign others
- **Admin**: Server management, no user assignment
- **User**: Basic server control (start/stop/console)

## 📊 **Monitoring & Stats**

### **Real-time Monitoring**
- CPU usage percentage
- Memory consumption
- Disk space usage
- Server uptime
- Process status

### **Server Console**
- Live log streaming via WebSocket
- Command execution
- Command history
- Console clearing
- Server status commands

## 🔧 **Advanced Configuration**

### **Environment Variables**
```bash
PORT=8080                    # Panel port
PANEL_NAME="MAX Panel"       # Panel display name
DOMAIN=panel.yourdomain.com  # Custom domain
```

### **Docker Compose Override**
```yaml
version: '3.8'
services:
  maxpanel:
    environment:
      - PANEL_NAME=My Game Panel
      - DOMAIN=games.mydomain.com
    ports:
      - "8080:8080"
      - "25565-25575:25565-25575"  # Minecraft servers
      - "27015-27025:27015-27025"  # Source games
```

## 🔒 **Security Features**

### **File System Security**
- Path traversal protection
- Server directory isolation
- File permission validation
- Upload size limits

### **User Security**
- bcrypt password hashing
- Session token validation
- Role-based access control
- API rate limiting

### **Network Security**
- Cloudflare DDoS protection
- SSL/TLS encryption
- Secure WebSocket connections
- IP-based restrictions

## 📈 **Scaling & Performance**

### **Multi-Node Support**
- Connect multiple VPS servers
- Distribute servers across nodes
- Load balancing
- Centralized management

### **Resource Optimization**
- Efficient process management
- Memory usage optimization
- Disk space monitoring
- Automatic cleanup

## 🆘 **Troubleshooting**

### **Common Issues**

**Cloudflare Setup Fails**
```bash
# Check API token permissions
curl -X GET "https://api.cloudflare.com/client/v4/user/tokens/verify" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Verify Zone ID
curl -X GET "https://api.cloudflare.com/client/v4/zones/YOUR_ZONE_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Server Won't Start**
```bash
# Check server logs
docker-compose -f docker-compose.vps.yml logs maxpanel

# Check server directory
ls -la /opt/maxpanel/servers/SERVER_ID/

# Check Java installation (for Minecraft)
java -version
```

**Domain Not Accessible**
```bash
# Check DNS propagation
nslookup panel.yourdomain.com

# Check SSL certificate
openssl s_client -connect panel.yourdomain.com:443

# Check Nginx status
docker-compose -f docker-compose.vps.yml logs nginx
```

## 🔄 **Updates & Maintenance**

### **Update Panel**
```bash
curl -sSL https://install.maxpanel.com | sudo bash
```

### **Backup Panel Data**
```bash
tar -czf maxpanel-backup-$(date +%Y%m%d).tar.gz /opt/maxpanel
```

### **Monitor Resources**
```bash
# Check disk usage
df -h /opt/maxpanel

# Check memory usage
free -h

# Check running servers
docker ps
```

## 📞 **Support & Community**

- **Documentation**: https://docs.maxpanel.com
- **GitHub**: https://github.com/maxpanel/panel
- **Discord**: https://discord.gg/maxpanel
- **Issues**: https://github.com/maxpanel/panel/issues

---

**🎮 Production Ready • 🌐 Cloudflare Integrated • 🥚 Advanced Eggs • 👥 Multi-User • 🔒 Secure**