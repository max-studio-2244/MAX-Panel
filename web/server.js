// Server Management JavaScript

let serverId = null;
let serverData = null;
let consoleWs = null;
let currentPath = '/';

// Initialize server management
document.addEventListener('DOMContentLoaded', function() {
    // Get server ID from URL
    const urlParams = new URLSearchParams(window.location.search);
    serverId = urlParams.get('id');
    
    if (!serverId) {
        window.location.href = '/';
        return;
    }
    
    loadServerData();
    connectConsole();
    loadFiles();
    loadBackups();
    
    // Auto-refresh server stats every 5 seconds
    setInterval(updateServerStats, 5000);
    
    // Handle console input
    document.getElementById('console-input').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            sendCommand();
        }
    });
});

// Load server data
async function loadServerData() {
    try {
        const response = await fetch(`/api/servers/${serverId}`);
        serverData = await response.json();
        
        document.getElementById('server-name').textContent = serverData.name;
        document.getElementById('server-ip').textContent = serverData.host || 'localhost';
        document.getElementById('server-port').textContent = serverData.port;
        document.getElementById('sidebar-ip').textContent = serverData.host || 'localhost';
        document.getElementById('sidebar-port').textContent = serverData.port;
        
        updateServerStatus(serverData.status);
        updateServerButtons(serverData.status);
        
        // Load server settings
        document.getElementById('server-name-input').value = serverData.name;
        document.getElementById('memory-limit-input').value = serverData.memory;
        document.getElementById('server-port-input').value = serverData.port;
        
    } catch (error) {
        console.error('Failed to load server data:', error);
        showNotification('Failed to load server data', 'error');
    }
}

// Update server status display
function updateServerStatus(status) {
    const statusElement = document.getElementById('server-status');
    const sidebarStatus = document.getElementById('sidebar-status');
    
    statusElement.className = 'px-3 py-1 rounded-full text-sm';
    
    switch (status) {
        case 'running':
            statusElement.classList.add('bg-green-500');
            statusElement.innerHTML = '<i class="fas fa-circle mr-1"></i>Online';
            sidebarStatus.textContent = 'Online';
            sidebarStatus.className = 'text-green-400';
            break;
        case 'starting':
            statusElement.classList.add('bg-yellow-500');
            statusElement.innerHTML = '<i class="fas fa-circle mr-1"></i>Starting';
            sidebarStatus.textContent = 'Starting';
            sidebarStatus.className = 'text-yellow-400';
            break;
        case 'stopping':
            statusElement.classList.add('bg-orange-500');
            statusElement.innerHTML = '<i class="fas fa-circle mr-1"></i>Stopping';
            sidebarStatus.textContent = 'Stopping';
            sidebarStatus.className = 'text-orange-400';
            break;
        default:
            statusElement.classList.add('bg-red-500');
            statusElement.innerHTML = '<i class="fas fa-circle mr-1"></i>Offline';
            sidebarStatus.textContent = 'Offline';
            sidebarStatus.className = 'text-red-400';
    }
}

// Update server control buttons
function updateServerButtons(status) {
    const startBtn = document.getElementById('start-btn');
    const stopBtn = document.getElementById('stop-btn');
    const restartBtn = document.getElementById('restart-btn');
    const killBtn = document.getElementById('kill-btn');
    
    if (status === 'running') {
        startBtn.disabled = true;
        startBtn.classList.add('opacity-50', 'cursor-not-allowed');
        stopBtn.disabled = false;
        stopBtn.classList.remove('opacity-50', 'cursor-not-allowed');
        restartBtn.disabled = false;
        restartBtn.classList.remove('opacity-50', 'cursor-not-allowed');
        killBtn.disabled = false;
        killBtn.classList.remove('opacity-50', 'cursor-not-allowed');
    } else {
        startBtn.disabled = false;
        startBtn.classList.remove('opacity-50', 'cursor-not-allowed');
        stopBtn.disabled = true;
        stopBtn.classList.add('opacity-50', 'cursor-not-allowed');
        restartBtn.disabled = true;
        restartBtn.classList.add('opacity-50', 'cursor-not-allowed');
        killBtn.disabled = true;
        killBtn.classList.add('opacity-50', 'cursor-not-allowed');
    }
}

// Connect to console WebSocket
function connectConsole() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/${serverId}`;
    
    consoleWs = new WebSocket(wsUrl);
    
    const output = document.getElementById('console-output');
    
    consoleWs.onopen = function() {
        addConsoleMessage('Connected to server console', 'system');
    };
    
    consoleWs.onmessage = function(event) {
        addConsoleMessage(event.data, 'server');
        output.scrollTop = output.scrollHeight;
    };
    
    consoleWs.onerror = function(error) {
        addConsoleMessage('Console connection error: ' + error, 'error');
    };
    
    consoleWs.onclose = function() {
        addConsoleMessage('Console disconnected', 'system');
        // Try to reconnect after 5 seconds
        setTimeout(connectConsole, 5000);
    };
}

// Add message to console
function addConsoleMessage(message, type = 'server') {
    const output = document.getElementById('console-output');
    const line = document.createElement('div');
    
    const timestamp = new Date().toLocaleTimeString();
    
    switch (type) {
        case 'command':
            line.innerHTML = `<span class="text-blue-400">[${timestamp}] $ ${message}</span>`;
            break;
        case 'system':
            line.innerHTML = `<span class="text-yellow-400">[${timestamp}] [SYSTEM] ${message}</span>`;
            break;
        case 'error':
            line.innerHTML = `<span class="text-red-400">[${timestamp}] [ERROR] ${message}</span>`;
            break;
        default:
            line.innerHTML = `<span class="text-green-400">[${timestamp}] ${message}</span>`;
    }
    
    output.appendChild(line);
    output.scrollTop = output.scrollHeight;
}

// Send command to server
function sendCommand(command) {
    const input = document.getElementById('console-input');
    const cmd = command || input.value.trim();
    
    if (cmd && consoleWs && consoleWs.readyState === WebSocket.OPEN) {
        consoleWs.send(cmd);
        addConsoleMessage(cmd, 'command');
        input.value = '';
    }
}

// Server control functions
async function startServer() {
    try {
        const response = await fetch(`/api/servers/${serverId}/start`, { method: 'POST' });
        if (response.ok) {
            showNotification('Server starting...', 'success');
            updateServerStatus('starting');
            updateServerButtons('starting');
        }
    } catch (error) {
        showNotification('Failed to start server', 'error');
    }
}

async function stopServer() {
    try {
        const response = await fetch(`/api/servers/${serverId}/stop`, { method: 'POST' });
        if (response.ok) {
            showNotification('Server stopping...', 'success');
            updateServerStatus('stopping');
            updateServerButtons('stopping');
        }
    } catch (error) {
        showNotification('Failed to stop server', 'error');
    }
}

async function restartServer() {
    try {
        const response = await fetch(`/api/servers/${serverId}/restart`, { method: 'POST' });
        if (response.ok) {
            showNotification('Server restarting...', 'success');
            updateServerStatus('starting');
            updateServerButtons('starting');
        }
    } catch (error) {
        showNotification('Failed to restart server', 'error');
    }
}

async function killServer() {
    if (confirm('Are you sure you want to force kill the server? This may cause data loss.')) {
        try {
            const response = await fetch(`/api/servers/${serverId}/kill`, { method: 'POST' });
            if (response.ok) {
                showNotification('Server killed', 'success');
                updateServerStatus('stopped');
                updateServerButtons('stopped');
            }
        } catch (error) {
            showNotification('Failed to kill server', 'error');
        }
    }
}

// Update server stats (CPU, Memory, Disk)
async function updateServerStats() {
    try {
        const response = await fetch(`/api/servers/${serverId}/stats`);
        if (response.ok) {
            const stats = await response.json();
            
            // Update CPU usage
            document.getElementById('cpu-usage').textContent = `${stats.cpu || 0}%`;
            document.getElementById('cpu-bar').style.width = `${stats.cpu || 0}%`;
            
            // Update Memory usage
            const memUsed = stats.memory_used || 0;
            const memTotal = serverData?.memory || 1024;
            const memPercent = (memUsed / memTotal) * 100;
            document.getElementById('memory-usage').textContent = `${memUsed} MB / ${memTotal} MB`;
            document.getElementById('memory-bar').style.width = `${memPercent}%`;
            
            // Update Disk usage
            const diskUsed = stats.disk_used || 0;
            const diskTotal = serverData?.disk || 5000;
            const diskPercent = (diskUsed / diskTotal) * 100;
            document.getElementById('disk-usage').textContent = `${diskUsed} MB / ${diskTotal} MB`;
            document.getElementById('disk-bar').style.width = `${diskPercent}%`;
            
            // Update uptime
            if (stats.uptime) {
                document.getElementById('sidebar-uptime').textContent = formatUptime(stats.uptime);
            }
        }
    } catch (error) {
        // Silently fail - stats are not critical
    }
}

// Format uptime
function formatUptime(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (hours > 0) {
        return `${hours}h ${minutes}m`;
    } else {
        return `${minutes}m`;
    }
}

// Tab management
function showTab(tabName) {
    // Hide all tabs
    document.querySelectorAll('.tab-content').forEach(tab => tab.classList.add('hidden'));
    
    // Show selected tab
    document.getElementById(`tab-${tabName}`).classList.remove('hidden');
    
    // Update tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active', 'border-blue-500', 'text-blue-400');
        btn.classList.add('text-gray-400');
    });
    
    event.target.classList.add('active', 'border-blue-500', 'text-blue-400');
    event.target.classList.remove('text-gray-400');
    
    // Load tab-specific data
    switch (tabName) {
        case 'files':
            loadFiles();
            break;
        case 'backups':
            loadBackups();
            break;
    }
}

// File management
async function loadFiles(path = '/') {
    try {
        const response = await fetch(`/api/servers/${serverId}/files?path=${encodeURIComponent(path)}`);
        const files = await response.json();
        
        currentPath = path;
        document.getElementById('current-path').textContent = path;
        
        const fileList = document.getElementById('file-list');
        fileList.innerHTML = '';
        
        // Add parent directory link if not at root
        if (path !== '/') {
            const parentPath = path.split('/').slice(0, -1).join('/') || '/';
            fileList.appendChild(createFileItem('..', parentPath, true, true));
        }
        
        // Add files and directories
        files.forEach(file => {
            fileList.appendChild(createFileItem(file.name, file.path, file.is_dir, false));
        });
        
    } catch (error) {
        showNotification('Failed to load files', 'error');
    }
}

// Create file item element
function createFileItem(name, path, isDir, isParent) {
    const item = document.createElement('div');
    item.className = 'flex items-center justify-between p-3 bg-gray-700 rounded-lg hover:bg-gray-600';
    
    const icon = isDir ? 'fa-folder' : 'fa-file';
    const iconColor = isDir ? 'text-blue-400' : 'text-gray-400';
    
    item.innerHTML = `
        <div class="flex items-center space-x-3">
            <i class="fas ${icon} ${iconColor}"></i>
            <span class="text-white">${name}</span>
        </div>
        <div class="flex space-x-2">
            ${!isParent && !isDir ? `
                <button onclick="downloadFile('${path}')" class="text-blue-400 hover:text-blue-300">
                    <i class="fas fa-download"></i>
                </button>
                <button onclick="editFile('${path}')" class="text-green-400 hover:text-green-300">
                    <i class="fas fa-edit"></i>
                </button>
            ` : ''}
            ${!isParent ? `
                <button onclick="deleteFile('${path}')" class="text-red-400 hover:text-red-300">
                    <i class="fas fa-trash"></i>
                </button>
            ` : ''}
        </div>
    `;
    
    // Add click handler for directories
    if (isDir) {
        item.style.cursor = 'pointer';
        item.addEventListener('click', () => loadFiles(path));
    }
    
    return item;
}

// File operations
async function downloadFile(path) {
    window.open(`/api/servers/${serverId}/files/download?path=${encodeURIComponent(path)}`);
}

function editFile(path) {
    showNotification('File editor coming soon!', 'info');
}

async function deleteFile(path) {
    if (confirm(`Are you sure you want to delete ${path}?`)) {
        try {
            const response = await fetch(`/api/servers/${serverId}/files/delete?path=${encodeURIComponent(path)}`, {
                method: 'DELETE'
            });
            if (response.ok) {
                showNotification('File deleted successfully', 'success');
                loadFiles(currentPath);
            }
        } catch (error) {
            showNotification('Failed to delete file', 'error');
        }
    }
}

function uploadFile() {
    showNotification('File upload coming soon!', 'info');
}

function createFolder() {
    const name = prompt('Enter folder name:');
    if (name) {
        showNotification('Create folder coming soon!', 'info');
    }
}

// Backup management
async function loadBackups() {
    try {
        const response = await fetch(`/api/servers/${serverId}/backups`);
        const backups = await response.json();
        
        const backupList = document.getElementById('backup-list');
        backupList.innerHTML = '';
        
        if (backups.length === 0) {
            backupList.innerHTML = '<p class="text-gray-400">No backups found</p>';
            return;
        }
        
        backups.forEach(backup => {
            const item = document.createElement('div');
            item.className = 'flex items-center justify-between p-4 bg-gray-700 rounded-lg';
            
            item.innerHTML = `
                <div>
                    <div class="font-medium">${backup.name}</div>
                    <div class="text-sm text-gray-400">
                        ${new Date(backup.created_at).toLocaleString()} • 
                        ${backup.size ? formatBytes(backup.size) : 'Unknown size'} • 
                        <span class="px-2 py-1 rounded text-xs ${backup.status === 'completed' ? 'bg-green-500' : 'bg-yellow-500'}">
                            ${backup.status}
                        </span>
                    </div>
                </div>
                <div class="flex space-x-2">
                    ${backup.status === 'completed' ? `
                        <button onclick="restoreBackup('${backup.id}')" class="bg-blue-600 hover:bg-blue-700 px-3 py-2 rounded text-sm">
                            <i class="fas fa-undo mr-1"></i>Restore
                        </button>
                    ` : ''}
                    <button onclick="deleteBackup('${backup.id}')" class="bg-red-600 hover:bg-red-700 px-3 py-2 rounded text-sm">
                        <i class="fas fa-trash mr-1"></i>Delete
                    </button>
                </div>
            `;
            
            backupList.appendChild(item);
        });
        
    } catch (error) {
        showNotification('Failed to load backups', 'error');
    }
}

async function createBackup() {
    const name = prompt('Enter backup name (optional):');
    try {
        const response = await fetch(`/api/servers/${serverId}/backups`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: name || '' })
        });
        
        if (response.ok) {
            showNotification('Backup started', 'success');
            loadBackups();
        }
    } catch (error) {
        showNotification('Failed to create backup', 'error');
    }
}

async function restoreBackup(backupId) {
    if (confirm('Are you sure you want to restore this backup? This will overwrite current server files.')) {
        try {
            const response = await fetch(`/api/servers/${serverId}/backups/${backupId}/restore`, {
                method: 'POST'
            });
            if (response.ok) {
                showNotification('Backup restore started', 'success');
            }
        } catch (error) {
            showNotification('Failed to restore backup', 'error');
        }
    }
}

async function deleteBackup(backupId) {
    if (confirm('Are you sure you want to delete this backup?')) {
        try {
            const response = await fetch(`/api/servers/${serverId}/backups/${backupId}`, {
                method: 'DELETE'
            });
            if (response.ok) {
                showNotification('Backup deleted', 'success');
                loadBackups();
            }
        } catch (error) {
            showNotification('Failed to delete backup', 'error');
        }
    }
}

// Utility functions
function formatBytes(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 px-6 py-3 rounded-lg shadow-lg z-50 ${
        type === 'success' ? 'bg-green-600' : 
        type === 'error' ? 'bg-red-600' : 'bg-blue-600'
    }`;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 3000);
}