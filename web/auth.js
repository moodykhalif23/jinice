const API_BASE = 'http://localhost:8080';

// Form switching functionality
document.addEventListener('DOMContentLoaded', () => {
  setupEventListeners();
});

function setupEventListeners() {
  // Tab switching
  document.getElementById('tab-login').addEventListener('click', () => showTab('login'));
  document.getElementById('tab-register').addEventListener('click', () => showTab('register'));

  // Switch buttons within forms
  document.getElementById('switch-to-register').addEventListener('click', () => showTab('register'));
  document.getElementById('switch-to-login').addEventListener('click', () => showTab('login'));

  // User type selection
  document.querySelectorAll('.user-type-option').forEach(option => {
    option.addEventListener('click', handleUserTypeSelection);
  });

  // Form submissions
  document.getElementById('login-form').addEventListener('submit', handleLogin);
  document.getElementById('register-form').addEventListener('submit', handleRegister);

  // Auto-focus first input on page load
  document.getElementById('login-email').focus();
}

function showTab(tabName) {
  // Update tabs
  document.querySelectorAll('.auth-tab').forEach(tab => tab.classList.remove('active'));
  document.getElementById(`tab-${tabName}`).classList.add('active');

  // Update forms
  document.querySelectorAll('.auth-form').forEach(form => form.classList.remove('active'));
  document.getElementById(`${tabName}-form`).classList.add('active');

  // Focus appropriate input
  setTimeout(() => {
    const inputToFocus = tabName === 'login' ? 'login-email' : 'register-name';
    document.getElementById(inputToFocus).focus();
  }, 100);
}

function handleUserTypeSelection(e) {
  const selectedOption = e.currentTarget;
  const userType = selectedOption.dataset.type;

  // Update selected state
  document.querySelectorAll('.user-type-option').forEach(option => {
    option.classList.remove('selected');
  });
  selectedOption.classList.add('selected');

  // Show/hide organization field
  const organizationField = document.getElementById('organization-field');
  const organizationInput = document.getElementById('register-organization');

  if (userType === 'event_owner') {
    organizationField.style.display = 'block';
    organizationInput.required = false; // Optional for event owners initially
  } else {
    organizationField.style.display = 'none';
    organizationInput.required = false;
    organizationInput.value = ''; // Clear value when hidden
  }
}

async function handleLogin(e) {
  e.preventDefault();

  const email = document.getElementById('login-email').value.trim();
  const password = document.getElementById('login-password').value;
  const submitBtn = document.getElementById('login-btn');

  if (!email || !password) {
    showError('Please fill in all fields');
    return;
  }

  toggleLoading(submitBtn, true);

  try {
    const response = await fetch(`${API_BASE}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ email, password })
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Login failed');
    }

    const data = await response.json();
    const userType = data.user.type;

    // Store auth token and user data
    sessionStorage.setItem('authToken', data.token);
    sessionStorage.setItem('currentUser', JSON.stringify(data.user));

    showSuccess('Login successful! Redirecting...');

    // Redirect based on user type
    setTimeout(() => {
      if (userType === 'business_owner') {
        window.location.href = 'business-owner.html';
      } else if (userType === 'event_owner') {
        window.location.href = 'manage-events.html';
      } else {
        window.location.href = 'index.html'; // fallback
      }
    }, 1500);

  } catch (error) {
    console.error('Login error:', error);
    showError(error.message || 'Login failed. Please check your credentials and try again.');
  } finally {
    toggleLoading(submitBtn, false);
  }
}

async function handleRegister(e) {
  e.preventDefault();

  const selectedType = document.querySelector('.user-type-option.selected');
  if (!selectedType) {
    showError('Please select an account type');
    return;
  }

  const userType = selectedType.dataset.type;
  const name = document.getElementById('register-name').value.trim();
  const email = document.getElementById('register-email').value.trim();
  const password = document.getElementById('register-password').value;
  const organization = document.getElementById('register-organization').value.trim();
  const submitBtn = document.getElementById('register-btn');

  if (!name || !email || !password) {
    showError('Please fill in all required fields');
    return;
  }

  if (password.length < 6) {
    showError('Password must be at least 6 characters long');
    return;
  }

  // Validate email format
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) {
    showError('Please enter a valid email address');
    return;
  }

  toggleLoading(submitBtn, true);

  try {
    const requestData = {
      name,
      email,
      password,
      type: userType,
      company: userType === 'business_owner' ? name.split(' ')[0] + "'s Business" : organization, // Default fallback
      phone: '' // Optional field, leave empty
    };

    const response = await fetch(`${API_BASE}/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(requestData)
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Registration failed');
    }

    const data = await response.json();

    // Store auth token and user data
    sessionStorage.setItem('authToken', data.token);
    sessionStorage.setItem('currentUser', JSON.stringify(data.user));

    showSuccess(`Welcome to JiNice! Redirecting...`);

    // Redirect based on user type
    setTimeout(() => {
      if (data.user.type === 'business_owner') {
        window.location.href = 'business-owner.html';
      } else if (data.user.type === 'event_owner') {
        window.location.href = 'manage-events.html';
      } else {
        window.location.href = 'index.html';
      }
    }, 2000);

  } catch (error) {
    console.error('Registration error:', error);
    showError(error.message || 'Registration failed. Please try again.');
  } finally {
    toggleLoading(submitBtn, false);
  }
}

function toggleLoading(button, loading) {
  const originalText = button.textContent;

  if (loading) {
    button.disabled = true;
    button.textContent = 'Please wait...';
    button.style.opacity = '0.7';
  } else {
    button.disabled = false;
    button.textContent = originalText;
    button.style.opacity = '1';
  }
}

function showSuccess(message) {
  removeNotifications();
  const notification = createNotification(message, 'success');
  document.body.appendChild(notification);

  setTimeout(() => {
    notification.remove();
  }, 3000);
}

function showError(message) {
  removeNotifications();
  const notification = createNotification(message, 'error');
  document.body.appendChild(notification);

  // Shake animation for errors
  notification.style.animation = 'shake 0.5s ease-in-out';

  setTimeout(() => {
    notification.remove();
  }, 5000);
}

function createNotification(message, type) {
  const notification = document.createElement('div');
  notification.style.cssText = `
    position: fixed;
    top: 20px;
    right: 20px;
    padding: 16px 20px;
    border-radius: 12px;
    font-weight: 500;
    font-size: 14px;
    z-index: 10000;
    min-width: 300px;
    max-width: 500px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
    animation: slideIn 0.3s ease-out;
  `;

  if (type === 'success') {
    notification.style.backgroundColor = '#d4edda';
    notification.style.color = '#155724';
    notification.style.borderLeft = '4px solid #28a745';
  } else {
    notification.style.backgroundColor = '#f8d7da';
    notification.style.color = '#721c24';
    notification.style.borderLeft = '4px solid #dc3545';
  }

  notification.innerHTML = `
    <div style="display: flex; align-items: center;">
      <span style="flex: 1;">${message}</span>
      <button onclick="this.parentElement.parentElement.remove()" style="
        background: none;
        border: none;
        font-size: 18px;
        cursor: pointer;
        opacity: 0.7;
        margin-left: 10px;
      ">&times;</button>
    </div>
  `;

  return notification;
}

function removeNotifications() {
  const notifications = document.querySelectorAll('body > div');
  notifications.forEach(notification => {
    if (notification.style.position === 'fixed') {
      notification.remove();
    }
  });
}

// Add CSS animations
const animationStyles = `
@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}
`;

const styleSheet = document.createElement('style');
styleSheet.textContent = animationStyles;
document.head.appendChild(styleSheet);
