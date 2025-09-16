// Global variables
let servers = [];
let currentConsoleWs = null;
let currentUser = null;

// Initialize app
document.addEventListener('DOMContentLoaded', function() {
    checkAuthentication();
    setInterval(loadServers, 5000); // Refresh every 5 seconds
});

// Check if user is authenticated
async function checkAuthentication() {
    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/login.html';
        return;
    }
    
    try {
        const response = await fetch('/api/auth/me', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (response.ok) {
            currentUser = await response.json();
            updateUIForUser();
            loadServers();
        } else {
            localStorage.removeItem('auth_token');
            window.location.href = '/login.html';
        }
    } catch (error) {
        localStorage.removeItem('auth_token');
        window.location.href = '/login.html';
    }
}

// Update UI based on user role
function updateUIForUser() {
    if (currentUser) {
        document.getElementById('username-display').textContent = currentUser.username;
        
        if (currentUser.is_admin) {
            // Show admin features
            document.getElementById('create-server-btn').style.display = 'block';
            document.getElementById('admin-link').style.display = 'block';
            document.getElementById('empty-admin-actions').style.display = 'block';
        } else {
            // Hide admin features for regular users
            document.getElementById('create-server-btn').style.display = 'none';
            document.getElementById('admin-link').style.display = 'none';
            document.getElementById('empty-admin-actions').style.display = 'none';
        }
    }
}

// Load servers from API
async function loadServers() {
    const token = localStorage.getItem('auth_token');
    if (!token) return;
    
    try {
        const response = await fetch('/api/servers', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (response.ok) {
            servers = await response.json();
            renderServers();
        } else if (response.status === 401) {
            localStorage.removeItem('auth_token');
            window.location.href = '/login.html';
        }
    } catch (error) {
        console.error('Failed to load servers:', error);
    }
}

// Render servers grid
function renderServers() {
    const grid = document.getElementById('servers-grid');
    const emptyState = document.getElementById('empty-state');
    
    if (servers.length === 0) {
        grid.innerHTML = '';
        emptyState.classList.remove('hidden');
        return;
    }
    
    emptyState.classList.add('hidden');
    
    grid.innerHTML = servers.map(server => `
        <div class="bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-700">
            <div class="flex items-center justify-between mb-4">
                <div class="flex items-center">
                    <div class="w-3 h-3 rounded-full mr-3 ${getStatusColor(server.status)}"></div>
                    <h3 class="text-lg font-semibold">${server.name}</h3>
                </div>
                <span class="text-xs px-2 py-1 rounded-full bg-gray-700 text-gray-300">${server.game.toUpperCase()}</span>
            </div>
            
            <div class="space-y-2 text-sm text-gray-400 mb-4">
                <div class="flex justify-between">
                    <span>Status:</span>
                    <span class="capitalize ${server.status === 'running' ? 'text-green-400' : 'text-red-400'}">${server.status}</span>
                </div>
                <div class="flex justify-between">
                    <span>Port:</span>
                    <span>${server.port}</span>
                </div>
                <div class="flex justify-between">
                    <span>Memory:</span>
                    <span>${server.memory}MB</span>
                </div>
                <div class="flex justify-between">
                    <span>CPU:</span>
                    <span>${server.cpu} cores</span>
                </div>
            </div>
            
            <div class="flex space-x-2">
                <button onclick="openServerManager('${server.id}')" class="flex-1 bg-blue-600 hover:bg-blue-700 px-3 py-2 rounded text-sm">
                    <i class="fas fa-cog mr-1"></i>Manage
                </button>
                ${server.status === 'running' 
                    ? `<button onclick="stopServer('${server.id}')" class="bg-red-600 hover:bg-red-700 px-3 py-2 rounded text-sm">
                         <i class="fas fa-stop"></i>
                       </button>`
                    : `<button onclick="startServer('${server.id}')" class="bg-green-600 hover:bg-green-700 px-3 py-2 rounded text-sm">
                         <i class="fas fa-play"></i>
                       </button>`
                }
                <button onclick="restartServer('${server.id}')" class="bg-yellow-600 hover:bg-yellow-700 px-3 py-2 rounded text-sm">
                    <i class="fas fa-redo"></i>
                </button>
                <button onclick="deleteServer('${server.id}')" class="bg-red-600 hover:bg-red-700 px-3 py-2 rounded text-sm">
                    <i class="fas fa-trash"></i>
                </button>
            </div>
        </div>
    `).join('');
}

// Get status indicator color
function getStatusColor(status) {
    switch (status) {
        case 'running': return 'bg-green-500';
        case 'stopped': return 'bg-red-500';
        case 'starting': return 'bg-yellow-500';
        default: return 'bg-gray-500';
    }
}

// Show create server modal
function showCreateModal() {
    document.getElementById('create-modal').classList.remove('hidden');
    document.getElementById('create-modal').classList.add('flex');
}

// Hide create server modal
function hideCreateModal() {
    document.getElementById('create-modal').classList.add('hidden');
    document.getElementById('create-modal').classList.remove('flex');
    document.getElementById('create-form').reset();
}

// Create new server
async function createServer(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const serverData = {
        name: formData.get('name'),
        game: formData.get('game'),
        port: parseInt(formData.get('port')),
        memory: parseInt(formData.get('memory')),
        cpu: parseFloat(formData.get('cpu'))
    };
    
    try {
        const response = await fetch('/api/servers', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(serverData)
        });
        
        if (response.ok) {
            hideCreateModal();
            loadServers();
            showNotification('Server created successfully!', 'success');
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to create server', 'error');
        }
    } catch (error) {
        showNotification('Failed to create server', 'error');
    }
}

// Start server
async function startServer(id) {
    try {
        const response = await fetch(`/api/servers/${id}/start`, { method: 'POST' });
        if (response.ok) {
            showNotification('Server starting...', 'success');
            loadServers();
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to start server', 'error');
        }
    } catch (error) {
        showNotification('Failed to start server', 'error');
    }
}

// Stop server
async function stopServer(id) {
    try {
        const response = await fetch(`/api/servers/${id}/stop`, { method: 'POST' });
        if (response.ok) {
            showNotification('Server stopping...', 'success');
            loadServers();
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to stop server', 'error');
        }
    } catch (error) {
        showNotification('Failed to stop server', 'error');
    }
}

// Restart server
async function restartServer(id) {
    try {
        const response = await fetch(`/api/servers/${id}/restart`, { method: 'POST' });
        if (response.ok) {
            showNotification('Server restarting...', 'success');
            loadServers();
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to restart server', 'error');
        }
    } catch (error) {
        showNotification('Failed to restart server', 'error');
    }
}

// Delete server
async function deleteServer(id) {
    if (!confirm('Are you sure you want to delete this server? This action cannot be undone.')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/servers/${id}`, { method: 'DELETE' });
        if (response.ok) {
            showNotification('Server deleted successfully!', 'success');
            loadServers();
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to delete server', 'error');
        }
    } catch (error) {
        showNotification('Failed to delete server', 'error');
    }
}

// Show console modal
function showConsole(serverId) {
    document.getElementById('console-modal').classList.remove('hidden');
    document.getElementById('console-modal').classList.add('flex');
    
    // Connect to WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/${serverId}`;
    
    currentConsoleWs = new WebSocket(wsUrl);
    
    const output = document.getElementById('console-output');
    output.innerHTML = '';
    
    currentConsoleWs.onmessage = function(event) {
        const line = document.createElement('div');
        line.textContent = event.data;
        output.appendChild(line);
        output.scrollTop = output.scrollHeight;
    };
    
    currentConsoleWs.onerror = function(error) {
        const line = document.createElement('div');
        line.textContent = 'WebSocket error: ' + error;
        line.className = 'text-red-400';
        output.appendChild(line);
    };
    
    // Handle Enter key in console input
    document.getElementById('console-input').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            sendCommand();
        }
    });
}

// Hide console modal
function hideConsoleModal() {
    document.getElementById('console-modal').classList.add('hidden');
    document.getElementById('console-modal').classList.remove('flex');
    
    if (currentConsoleWs) {
        currentConsoleWs.close();
        currentConsoleWs = null;
    }
    
    document.getElementById('console-input').value = '';
}

// Send command to server
function sendCommand() {
    const input = document.getElementById('console-input');
    const command = input.value.trim();
    
    if (command && currentConsoleWs && currentConsoleWs.readyState === WebSocket.OPEN) {
        currentConsoleWs.send(command);
        
        // Show command in output
        const output = document.getElementById('console-output');
        const line = document.createElement('div');
        line.innerHTML = `<span class="text-blue-400">$ ${command}</span>`;
        output.appendChild(line);
        output.scrollTop = output.scrollHeight;
        
        input.value = '';
    }
}

// Open server management panel
function openServerManager(serverId) {
    window.open(`/server.html?id=${serverId}`, '_blank');
}

// User menu functions
function toggleUserMenu() {
    const menu = document.getElementById('user-menu');
    menu.classList.toggle('hidden');
}

function showProfile() {
    showNotification('Profile settings coming soon!', 'info');
    document.getElementById('user-menu').classList.add('hidden');
}

function logout() {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('remember_user');
    window.location.href = '/login.html';
}

// Show notification
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