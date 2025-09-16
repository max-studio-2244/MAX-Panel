// Admin Dashboard JavaScript

let currentUser = null;
let eggs = [];
let nodes = [];
let users = [];

// Initialize admin dashboard
document.addEventListener('DOMContentLoaded', function() {
    loadDashboardStats();
    loadUsers();
    loadEggs();
    loadNodes();
    loadSettings();
    loadActivityLogs();
});

// Show admin section
function showAdminSection(section) {
    // Hide all sections
    document.querySelectorAll('.admin-section').forEach(el => el.classList.add('hidden'));
    
    // Show selected section
    document.getElementById(`admin-${section}`).classList.remove('hidden');
    
    // Update navigation
    document.querySelectorAll('.admin-nav-btn').forEach(btn => {
        btn.classList.remove('active', 'border-blue-500', 'text-blue-400');
        btn.classList.add('text-gray-400');
    });
    
    event.target.classList.add('active', 'border-blue-500', 'text-blue-400');
    event.target.classList.remove('text-gray-400');
}

// Load dashboard statistics
async function loadDashboardStats() {
    try {
        const [serversRes, usersRes, nodesRes] = await Promise.all([
            fetch('/api/servers'),
            fetch('/api/admin/users'),
            fetch('/api/admin/nodes')
        ]);
        
        const servers = await serversRes.json();
        const usersData = await usersRes.json();
        const nodesData = await nodesRes.json();
        
        document.getElementById('total-servers').textContent = servers.length;
        document.getElementById('total-users').textContent = usersData.length;
        
        // Calculate resource usage
        let totalMemory = 0;
        let totalDisk = 0;
        servers.forEach(server => {
            totalMemory += server.memory || 0;
            totalDisk += server.disk || 0;
        });
        
        document.getElementById('memory-used').textContent = `${(totalMemory / 1024).toFixed(1)} GB`;
        document.getElementById('disk-used').textContent = `${(totalDisk / 1024).toFixed(1)} GB`;
        
    } catch (error) {
        console.error('Failed to load dashboard stats:', error);
    }
}

// Load users
async function loadUsers() {
    try {
        const response = await fetch('/api/admin/users');
        users = await response.json();
        renderUsers();
    } catch (error) {
        console.error('Failed to load users:', error);
    }
}

// Render users table
function renderUsers() {
    const tbody = document.getElementById('users-table');
    tbody.innerHTML = users.map(user => `
        <tr class="border-t border-gray-700">
            <td class="px-6 py-4">
                <div class="flex items-center">
                    <div class="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center mr-3">
                        ${user.first_name.charAt(0)}${user.last_name.charAt(0)}
                    </div>
                    <div>
                        <div class="font-medium">${user.first_name} ${user.last_name}</div>
                        <div class="text-gray-400 text-sm">@${user.username}</div>
                    </div>
                </div>
            </td>
            <td class="px-6 py-4">${user.email}</td>
            <td class="px-6 py-4">
                <span class="px-2 py-1 text-xs rounded-full ${user.is_admin ? 'bg-red-500' : 'bg-blue-500'}">
                    ${user.is_admin ? 'Admin' : 'User'}
                </span>
            </td>
            <td class="px-6 py-4 text-gray-400">${new Date(user.created_at).toLocaleDateString()}</td>
            <td class="px-6 py-4">
                <div class="flex space-x-2">
                    <button onclick="editUser('${user.id}')" class="text-blue-400 hover:text-blue-300">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button onclick="deleteUser('${user.id}')" class="text-red-400 hover:text-red-300">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </td>
        </tr>
    `).join('');
}

// Load game eggs
async function loadEggs() {
    try {
        const response = await fetch('/api/admin/eggs');
        eggs = await response.json();
        renderEggs();
    } catch (error) {
        console.error('Failed to load eggs:', error);
        // Load default eggs if API fails
        loadDefaultEggs();
    }
}

// Load default eggs
function loadDefaultEggs() {
    eggs = [
        {
            id: 'minecraft-vanilla',
            name: 'Minecraft Vanilla',
            description: 'Official Minecraft server',
            author: 'Mojang',
            image: 'itzg/minecraft-server:latest',
            category: 'Minecraft',
            start_command: 'java -Xms{{SERVER_MEMORY}}M -Xmx{{SERVER_MEMORY}}M -jar server.jar nogui'
        },
        {
            id: 'minecraft-paper',
            name: 'Minecraft Paper',
            description: 'High performance Minecraft server',
            author: 'PaperMC',
            image: 'itzg/minecraft-server:latest',
            category: 'Minecraft',
            start_command: 'java -Xms{{SERVER_MEMORY}}M -Xmx{{SERVER_MEMORY}}M -jar paper.jar nogui'
        },
        {
            id: 'csgo',
            name: 'Counter-Strike: Global Offensive',
            description: 'CS:GO dedicated server',
            author: 'Valve',
            image: 'steamcmd/steamcmd:latest',
            category: 'Source Engine',
            start_command: './srcds_run -game csgo -console -usercon +game_type 0 +game_mode 1 +mapgroup mg_active +map de_dust2'
        },
        {
            id: 'rust',
            name: 'Rust Server',
            description: 'Rust dedicated server',
            author: 'Facepunch Studios',
            image: 'didstopia/rust-server:latest',
            category: 'Survival',
            start_command: './RustDedicated -batchmode +server.port {{SERVER_PORT}} +server.identity "rust" +rcon.port {{RCON_PORT}} +rcon.web 1'
        }
    ];
    renderEggs();
}

// Render eggs grid
function renderEggs() {
    const grid = document.getElementById('eggs-grid');
    grid.innerHTML = eggs.map(egg => `
        <div class="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <div class="flex items-center justify-between mb-4">
                <h3 class="text-lg font-semibold">${egg.name}</h3>
                <span class="text-xs px-2 py-1 bg-blue-500 rounded-full">${egg.category}</span>
            </div>
            <p class="text-gray-400 text-sm mb-4">${egg.description}</p>
            <div class="text-xs text-gray-500 mb-4">
                <div>Author: ${egg.author}</div>
                <div>Image: ${egg.image}</div>
            </div>
            <div class="flex space-x-2">
                <button onclick="editEgg('${egg.id}')" class="flex-1 bg-blue-600 hover:bg-blue-700 px-3 py-2 rounded text-sm">
                    <i class="fas fa-edit mr-1"></i>Edit
                </button>
                <button onclick="exportEgg('${egg.id}')" class="bg-green-600 hover:bg-green-700 px-3 py-2 rounded text-sm">
                    <i class="fas fa-download"></i>
                </button>
                <button onclick="deleteEgg('${egg.id}')" class="bg-red-600 hover:bg-red-700 px-3 py-2 rounded text-sm">
                    <i class="fas fa-trash"></i>
                </button>
            </div>
        </div>
    `).join('');
}

// Load nodes
async function loadNodes() {
    try {
        const response = await fetch('/api/admin/nodes');
        nodes = await response.json();
        renderNodes();
    } catch (error) {
        console.error('Failed to load nodes:', error);
        // Show default local node
        nodes = [{
            id: 'local',
            name: 'Local Node',
            description: 'Local Docker daemon',
            host: 'localhost',
            status: 'online',
            memory_total: 8192,
            memory_allocated: 2048,
            disk_total: 100000,
            disk_allocated: 25000
        }];
        renderNodes();
    }
}

// Render nodes grid
function renderNodes() {
    const grid = document.getElementById('nodes-grid');
    grid.innerHTML = nodes.map(node => `
        <div class="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <div class="flex items-center justify-between mb-4">
                <h3 class="text-lg font-semibold">${node.name}</h3>
                <span class="px-2 py-1 text-xs rounded-full ${node.status === 'online' ? 'bg-green-500' : 'bg-red-500'}">
                    ${node.status}
                </span>
            </div>
            <p class="text-gray-400 text-sm mb-4">${node.description}</p>
            <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                    <span>Host:</span>
                    <span class="text-gray-300">${node.host}</span>
                </div>
                <div class="flex justify-between">
                    <span>Memory:</span>
                    <span class="text-gray-300">${node.memory_allocated}MB / ${node.memory_total}MB</span>
                </div>
                <div class="flex justify-between">
                    <span>Disk:</span>
                    <span class="text-gray-300">${(node.disk_allocated/1024).toFixed(1)}GB / ${(node.disk_total/1024).toFixed(1)}GB</span>
                </div>
            </div>
            <div class="mt-4 flex space-x-2">
                <button onclick="showNodeToken('${node.id}')" class="flex-1 bg-blue-600 hover:bg-blue-700 px-3 py-2 rounded text-sm">
                    <i class="fas fa-key mr-1"></i>Token
                </button>
                <button onclick="editNode('${node.id}')" class="bg-yellow-600 hover:bg-yellow-700 px-3 py-2 rounded text-sm">
                    <i class="fas fa-edit"></i>
                </button>
                <button onclick="deleteNode('${node.id}')" class="bg-red-600 hover:bg-red-700 px-3 py-2 rounded text-sm">
                    <i class="fas fa-trash"></i>
                </button>
            </div>
        </div>
    `).join('');
}

// Load panel settings
async function loadSettings() {
    try {
        const response = await fetch('/api/admin/settings');
        const settings = await response.json();
        
        document.getElementById('panel-name').value = settings.panel_name || 'MAX Panel';
        document.getElementById('primary-color').value = settings.primary_color || '#3B82F6';
        document.getElementById('theme').value = settings.theme || 'dark';
        document.getElementById('animations').checked = settings.animations_enabled !== false;
    } catch (error) {
        console.error('Failed to load settings:', error);
    }
}

// Save panel settings
async function saveSettings() {
    const settings = {
        panel_name: document.getElementById('panel-name').value,
        primary_color: document.getElementById('primary-color').value,
        theme: document.getElementById('theme').value,
        animations_enabled: document.getElementById('animations').checked
    };
    
    try {
        const response = await fetch('/api/admin/settings', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(settings)
        });
        
        if (response.ok) {
            showNotification('Settings saved successfully!', 'success');
        }
    } catch (error) {
        showNotification('Failed to save settings', 'error');
    }
}

// Load activity logs
async function loadActivityLogs() {
    try {
        const response = await fetch('/api/admin/logs');
        const logs = await response.json();
        renderActivityLogs(logs);
    } catch (error) {
        console.error('Failed to load activity logs:', error);
    }
}

// Render activity logs
function renderActivityLogs(logs) {
    const tbody = document.getElementById('logs-table');
    tbody.innerHTML = logs.map(log => `
        <tr class="border-t border-gray-700">
            <td class="px-6 py-4 text-gray-400">${new Date(log.created_at).toLocaleString()}</td>
            <td class="px-6 py-4">${log.user_id || 'System'}</td>
            <td class="px-6 py-4">${log.action}</td>
            <td class="px-6 py-4 text-gray-400">${log.ip_address || 'N/A'}</td>
        </tr>
    `).join('');
}

// Show node token
function showNodeToken(nodeId) {
    const node = nodes.find(n => n.id === nodeId);
    if (node) {
        const token = node.token || 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...';
        navigator.clipboard.writeText(token);
        showNotification('Node token copied to clipboard!', 'success');
    }
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

// Placeholder functions for modals and actions
function showCreateUserModal() { showNotification('Create user modal - Coming soon!'); }
function showCreateEggModal() { showNotification('Create egg modal - Coming soon!'); }
function showCreateNodeModal() { showNotification('Create node modal - Coming soon!'); }
function editUser(id) { showNotification('Edit user - Coming soon!'); }
function deleteUser(id) { showNotification('Delete user - Coming soon!'); }
function editEgg(id) { showNotification('Edit egg - Coming soon!'); }
function deleteEgg(id) { showNotification('Delete egg - Coming soon!'); }
function exportEgg(id) { showNotification('Export egg - Coming soon!'); }
function importEgg() { showNotification('Import egg - Coming soon!'); }
function editNode(id) { showNotification('Edit node - Coming soon!'); }
function deleteNode(id) { showNotification('Delete node - Coming soon!'); }