// Authentication JavaScript

let currentUser = null;

// Initialize auth page
document.addEventListener('DOMContentLoaded', function() {
    // Check if user is already logged in
    const token = localStorage.getItem('auth_token');
    if (token) {
        // Verify token and redirect if valid
        verifyToken(token);
    }
});

// Show different forms
function showLogin() {
    hideAllForms();
    document.getElementById('login-form').classList.remove('hidden');
}

function showRegister() {
    hideAllForms();
    document.getElementById('register-form').classList.remove('hidden');
}

function showForgotPassword() {
    hideAllForms();
    document.getElementById('forgot-form').classList.remove('hidden');
}

function showTwoFA() {
    hideAllForms();
    document.getElementById('twofa-form').classList.remove('hidden');
}

function hideAllForms() {
    document.getElementById('login-form').classList.add('hidden');
    document.getElementById('register-form').classList.add('hidden');
    document.getElementById('forgot-form').classList.add('hidden');
    document.getElementById('twofa-form').classList.add('hidden');
}

// Toggle password visibility
function togglePassword() {
    const passwordInput = document.getElementById('password');
    const toggleIcon = document.getElementById('password-toggle');
    
    if (passwordInput.type === 'password') {
        passwordInput.type = 'text';
        toggleIcon.classList.remove('fa-eye');
        toggleIcon.classList.add('fa-eye-slash');
    } else {
        passwordInput.type = 'password';
        toggleIcon.classList.remove('fa-eye-slash');
        toggleIcon.classList.add('fa-eye');
    }
}

// Handle login
async function handleLogin(event) {
    event.preventDefault();
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const remember = document.getElementById('remember').checked;
    
    try {
        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username,
                password: password
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            // Check if 2FA is enabled
            if (data.user.two_factor_enabled) {
                currentUser = data.user;
                showTwoFA();
                showNotification('Please enter your 2FA code', 'info');
            } else {
                // Login successful
                localStorage.setItem('auth_token', data.token);
                if (remember) {
                    localStorage.setItem('remember_user', 'true');
                }
                
                showNotification('Login successful!', 'success');
                
                // Redirect based on user role
                if (data.user.is_admin) {
                    window.location.href = '/admin.html';
                } else {
                    window.location.href = '/';
                }
            }
        } else {
            showNotification(data.error || 'Login failed', 'error');
        }
    } catch (error) {
        showNotification('Network error. Please try again.', 'error');
    }
}

// Handle registration
async function handleRegister(event) {
    event.preventDefault();
    
    const firstName = document.getElementById('reg-first-name').value;
    const lastName = document.getElementById('reg-last-name').value;
    const username = document.getElementById('reg-username').value;
    const email = document.getElementById('reg-email').value;
    const password = document.getElementById('reg-password').value;
    const confirmPassword = document.getElementById('reg-confirm-password').value;
    
    // Validate passwords match
    if (password !== confirmPassword) {
        showNotification('Passwords do not match', 'error');
        return;
    }
    
    // Validate password strength
    if (password.length < 8) {
        showNotification('Password must be at least 8 characters long', 'error');
        return;
    }
    
    try {
        const response = await fetch('/api/auth/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                first_name: firstName,
                last_name: lastName,
                username: username,
                email: email,
                password: password,
                is_admin: false
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showNotification('Account created successfully! Please sign in.', 'success');
            showLogin();
            
            // Pre-fill login form
            document.getElementById('username').value = username;
        } else {
            showNotification(data.error || 'Registration failed', 'error');
        }
    } catch (error) {
        showNotification('Network error. Please try again.', 'error');
    }
}

// Handle forgot password
async function handleForgotPassword(event) {
    event.preventDefault();
    
    const email = document.getElementById('forgot-email').value;
    
    try {
        const response = await fetch('/api/auth/forgot-password', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                email: email
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showNotification('Password reset link sent to your email!', 'success');
            showLogin();
        } else {
            showNotification(data.error || 'Failed to send reset link', 'error');
        }
    } catch (error) {
        showNotification('Network error. Please try again.', 'error');
    }
}

// Handle 2FA verification
async function handleTwoFA(event) {
    event.preventDefault();
    
    const code = document.getElementById('twofa-code').value;
    
    if (code.length !== 6) {
        showNotification('Please enter a 6-digit code', 'error');
        return;
    }
    
    try {
        const response = await fetch('/api/auth/verify-2fa', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                user_id: currentUser.id,
                code: code
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            localStorage.setItem('auth_token', data.token);
            showNotification('2FA verification successful!', 'success');
            
            // Redirect based on user role
            if (currentUser.is_admin) {
                window.location.href = '/admin.html';
            } else {
                window.location.href = '/';
            }
        } else {
            showNotification(data.error || '2FA verification failed', 'error');
        }
    } catch (error) {
        showNotification('Network error. Please try again.', 'error');
    }
}

// Verify existing token
async function verifyToken(token) {
    try {
        const response = await fetch('/api/auth/me', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (response.ok) {
            const user = await response.json();
            
            // Redirect based on user role
            if (user.is_admin) {
                window.location.href = '/admin.html';
            } else {
                window.location.href = '/';
            }
        } else {
            // Token invalid, remove it
            localStorage.removeItem('auth_token');
        }
    } catch (error) {
        // Network error, stay on login page
        localStorage.removeItem('auth_token');
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
    }, 4000);
}

// Auto-format 2FA code input
document.addEventListener('DOMContentLoaded', function() {
    const twofaInput = document.getElementById('twofa-code');
    if (twofaInput) {
        twofaInput.addEventListener('input', function(e) {
            // Only allow numbers
            e.target.value = e.target.value.replace(/[^0-9]/g, '');
            
            // Auto-submit when 6 digits entered
            if (e.target.value.length === 6) {
                setTimeout(() => {
                    document.querySelector('#twofa-form form').dispatchEvent(new Event('submit'));
                }, 500);
            }
        });
    }
});