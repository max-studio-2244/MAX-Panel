let eggs = [];
let users = [];
let editingEgg = null;

// Load data on page load
document.addEventListener('DOMContentLoaded', function() {
    loadEggs();
    loadUsers();
});

async function loadEggs() {
    try {
        const response = await fetch('/api/admin/eggs');
        eggs = await response.json();
        renderEggs();
    } catch (error) {
        console.error('Failed to load eggs:', error);
    }
}

async function loadUsers() {
    try {
        const response = await fetch('/api/admin/users');
        users = await response.json();
        populateUserSelect();
    } catch (error) {
        console.error('Failed to load users:', error);
    }
}

function renderEggs() {
    const grid = document.getElementById('eggs-grid');
    grid.innerHTML = '';

    eggs.forEach(egg => {
        const eggCard = document.createElement('div');
        eggCard.className = 'bg-gray-800 rounded-lg p-6 border border-gray-700';
        eggCard.innerHTML = `
            <div class="flex justify-between items-start mb-4">
                <div>
                    <h3 class="text-xl font-semibold text-blue-400">${egg.name}</h3>
                    <p class="text-gray-400 text-sm">${egg.category}</p>
                </div>
                <div class="flex space-x-2">
                    <button onclick="editEgg('${egg.id}')" class="text-blue-400 hover:text-blue-300">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button onclick="deleteEgg('${egg.id}')" class="text-red-400 hover:text-red-300">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
            
            <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                    <span class="text-gray-400">Game:</span>
                    <span>${egg.game}</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-400">Version:</span>
                    <span>${egg.version || 'Latest'}</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-400">Memory:</span>
                    <span>${egg.min_memory}MB - ${egg.max_memory}MB</span>
                </div>
            </div>
            
            <p class="text-gray-300 text-sm mt-3">${egg.description || 'No description'}</p>
            
            <div class="mt-4 pt-4 border-t border-gray-700">
                <button onclick="showAssignModal('${egg.id}')" class="w-full bg-green-600 hover:bg-green-700 px-4 py-2 rounded-lg text-sm">
                    <i class="fas fa-server mr-2"></i>Create Server
                </button>
            </div>
        `;
        grid.appendChild(eggCard);
    });
}

function populateUserSelect() {
    const select = document.querySelector('select[name="owner_id"]');
    select.innerHTML = '<option value="">Select User</option>';
    
    users.forEach(user => {
        const option = document.createElement('option');
        option.value = user.id;
        option.textContent = `${user.username} (${user.email})`;
        select.appendChild(option);
    });
}

function showCreateEggModal() {
    editingEgg = null;
    document.getElementById('modal-title').textContent = 'Create New Egg';
    document.getElementById('egg-form').reset();
    document.getElementById('egg-modal').classList.remove('hidden');
    document.getElementById('egg-modal').classList.add('flex');
}

function editEgg(eggId) {
    editingEgg = eggs.find(e => e.id === eggId);
    if (!editingEgg) return;

    document.getElementById('modal-title').textContent = 'Edit Egg';
    
    const form = document.getElementById('egg-form');
    form.name.value = editingEgg.name;
    form.description.value = editingEgg.description || '';
    form.game.value = editingEgg.game;
    form.version.value = editingEgg.version || '';
    form.build_number.value = editingEgg.build_number || '';
    form.category.value = editingEgg.category;
    form.image.value = editingEgg.image || '';
    form.min_memory.value = editingEgg.min_memory;
    form.max_memory.value = editingEgg.max_memory;
    form.start_command.value = editingEgg.start_command;
    form.stop_command.value = editingEgg.stop_command || '';
    form.install_script.value = editingEgg.install_script || '';
    form.config_files.value = editingEgg.config_files || '';
    form.environment.value = editingEgg.environment || '';
    form.ports.value = editingEgg.ports || '';

    document.getElementById('egg-modal').classList.remove('hidden');
    document.getElementById('egg-modal').classList.add('flex');
}

function hideEggModal() {
    document.getElementById('egg-modal').classList.add('hidden');
    document.getElementById('egg-modal').classList.remove('flex');
}

async function saveEgg(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const eggData = {
        name: formData.get('name'),
        description: formData.get('description'),
        game: formData.get('game'),
        version: formData.get('version'),
        build_number: formData.get('build_number'),
        category: formData.get('category'),
        image: formData.get('image'),
        min_memory: parseInt(formData.get('min_memory')),
        max_memory: parseInt(formData.get('max_memory')),
        start_command: formData.get('start_command'),
        stop_command: formData.get('stop_command'),
        install_script: formData.get('install_script'),
        config_files: formData.get('config_files'),
        environment: formData.get('environment'),
        ports: formData.get('ports')
    };

    try {
        let response;
        if (editingEgg) {
            response = await fetch(`/api/admin/eggs/${editingEgg.id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(eggData)
            });
        } else {
            response = await fetch('/api/admin/eggs', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(eggData)
            });
        }

        if (response.ok) {
            hideEggModal();
            loadEggs();
            showNotification(editingEgg ? 'Egg updated successfully' : 'Egg created successfully', 'success');
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to save egg', 'error');
        }
    } catch (error) {
        showNotification('Network error', 'error');
    }
}

async function deleteEgg(eggId) {
    if (!confirm('Are you sure you want to delete this egg?')) return;

    try {
        const response = await fetch(`/api/admin/eggs/${eggId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            loadEggs();
            showNotification('Egg deleted successfully', 'success');
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to delete egg', 'error');
        }
    } catch (error) {
        showNotification('Network error', 'error');
    }
}

function showAssignModal(eggId) {
    const egg = eggs.find(e => e.id === eggId);
    if (!egg) return;

    document.getElementById('assign-egg-id').value = eggId;
    document.querySelector('#assign-form input[name="memory"]').value = egg.min_memory;
    document.querySelector('#assign-form input[name="memory"]').min = egg.min_memory;
    document.querySelector('#assign-form input[name="memory"]').max = egg.max_memory;

    // Set default port based on game
    const defaultPorts = {
        'minecraft': 25565,
        'nodejs': 3000,
        'csgo': 27015,
        'rust': 28015
    };
    document.querySelector('#assign-form input[name="port"]').value = defaultPorts[egg.game] || 25565;

    document.getElementById('assign-modal').classList.remove('hidden');
    document.getElementById('assign-modal').classList.add('flex');
}

function hideAssignModal() {
    document.getElementById('assign-modal').classList.add('hidden');
    document.getElementById('assign-modal').classList.remove('flex');
}

async function createServerFromEgg(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const serverData = {
        name: formData.get('name'),
        egg_id: formData.get('egg_id'),
        port: parseInt(formData.get('port')),
        memory: parseInt(formData.get('memory')),
        owner_id: formData.get('owner_id') || null
    };

    try {
        const response = await fetch('/api/admin/servers/create-from-egg', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(serverData)
        });

        if (response.ok) {
            hideAssignModal();
            showNotification('Server created successfully', 'success');
            // Optionally redirect to server management
            setTimeout(() => {
                window.location.href = '/admin.html';
            }, 2000);
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to create server', 'error');
        }
    } catch (error) {
        showNotification('Network error', 'error');
    }
}

function showNotification(message, type) {
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 px-6 py-3 rounded-lg text-white z-50 ${
        type === 'success' ? 'bg-green-600' : 'bg-red-600'
    }`;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 3000);
}