const base = 'http://localhost:8080';
let currentUser = null;
let authToken = null;
let allBusinesses = [];

// Check login status on page load
document.addEventListener('DOMContentLoaded', () => {
  const savedToken = sessionStorage.getItem('authToken');
  const savedUser = sessionStorage.getItem('currentUser');

  if (savedToken && savedUser) {
    try {
      currentUser = JSON.parse(savedUser);
      authToken = savedToken;
      showDashboard();
      loadUserData();
    } catch (e) {
      console.error('Error parsing saved user data', e);
      showNotAuthenticated();
    }
  } else {
    showNotAuthenticated();
  }

  setupEventListeners();
});

function setupEventListeners() {
  // Form switches
  document.getElementById('show-register').addEventListener('click', showRegister);
  document.getElementById('show-login').addEventListener('click', showLogin);

  // Navigation auth buttons
  document.getElementById('show-login-nav').addEventListener('click', showLogin);
  document.getElementById('show-register-nav').addEventListener('click', showRegister);

  // Login form
  document.getElementById('login-form').addEventListener('submit', handleLogin);

  // Register form
  document.getElementById('register-form').addEventListener('submit', handleRegister);

  // Logout
  document.addEventListener('click', (e) => {
    if (e.target.id === 'logout-btn' || e.target.id === 'logout-nav') {
      handleLogout();
    }
  });

  // Other event listeners
  const refreshBtn = document.getElementById('btn-refresh-businesses');
  if (refreshBtn) {
    refreshBtn.addEventListener('click', () => {
      if (currentUser) {
        loadUserData();
      }
    });
  }

  // Add Business
  const addBusinessBtn = document.getElementById('btn-add-business');
  if (addBusinessBtn) {
    addBusinessBtn.addEventListener('click', handleAddBusiness);
  }
}

function showLogin() {
  document.getElementById('login-container').style.display = 'block';
  document.getElementById('register-container').style.display = 'none';
  document.getElementById('dashboard').style.display = 'none';
  document.getElementById('user-status').textContent = 'Please log in to manage your businesses';

  // Update navigation buttons
  document.getElementById('show-login-nav').style.display = 'none';
  document.getElementById('show-register-nav').style.display = 'none';
  document.getElementById('logout-nav').style.display = 'none';
}

function showRegister() {
  document.getElementById('login-container').style.display = 'none';
  document.getElementById('register-container').style.display = 'block';
  document.getElementById('dashboard').style.display = 'none';
  document.getElementById('user-status').textContent = 'Register as a business owner';

  // Update navigation buttons
  document.getElementById('show-login-nav').style.display = 'none';
  document.getElementById('show-register-nav').style.display = 'none';
  document.getElementById('logout-nav').style.display = 'none';
}

function showNotAuthenticated() {
  document.getElementById('login-container').style.display = 'none';
  document.getElementById('register-container').style.display = 'none';
  document.getElementById('dashboard').style.display = 'none';
  document.getElementById('user-status').textContent = 'Welcome to the Business Owner Portal. Please login or register to access your dashboard.';

  // Show navigation authentication buttons in top right
  const showLoginNav = document.getElementById('show-login-nav');
  const showRegisterNav = document.getElementById('show-register-nav');
  const logoutNav = document.getElementById('logout-nav');
  
  if (showLoginNav) showLoginNav.style.display = 'inline-block';
  if (showRegisterNav) showRegisterNav.style.display = 'inline-block';
  if (logoutNav) logoutNav.style.display = 'none';
}

function showDashboard() {
  document.getElementById('login-container').style.display = 'none';
  document.getElementById('register-container').style.display = 'none';
  document.getElementById('dashboard').style.display = 'block';
  document.getElementById('user-status').textContent = `Welcome, ${currentUser.name}! Manage your business listings and monitor performance.`;

  // Update navigation buttons - show logout when logged in
  document.getElementById('show-login-nav').style.display = 'none';
  document.getElementById('show-register-nav').style.display = 'none';
  document.getElementById('logout-nav').style.display = 'inline-block';
}

function handleLogin(e) {
  e.preventDefault();

  const email = document.getElementById('login-email').value;
  const password = document.getElementById('login-password').value;

  fetch(base + '/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  })
  .then(res => {
    if (!res.ok) {
      throw new Error('Login failed');
    }
    return res.json();
  })
  .then(data => {
    currentUser = data.user;
    authToken = data.token;

    // Save to sessionStorage (not localStorage for better security)
    sessionStorage.setItem('authToken', authToken);
    sessionStorage.setItem('currentUser', JSON.stringify(currentUser));

    showDashboard();
    loadUserData();
    alert('Login successful!');
  })
  .catch(err => {
    console.error('Login error:', err);
    alert('Login failed. Please check your credentials.');
  });
}

function handleRegister(e) {
  e.preventDefault();

  const name = document.getElementById('register-name').value;
  const email = document.getElementById('register-email').value;
  const password = document.getElementById('register-password').value;
  const company = document.getElementById('register-company').value;
  const phone = document.getElementById('register-phone').value;

  fetch(base + '/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, email, password, company, phone })
  })
  .then(res => {
    if (!res.ok) {
      throw new Error('Registration failed');
    }
    return res.json();
  })
  .then(data => {
    currentUser = data.user;
    authToken = data.token;

    // Save to sessionStorage (not localStorage for better security)
    sessionStorage.setItem('authToken', authToken);
    sessionStorage.setItem('currentUser', JSON.stringify(currentUser));

    showDashboard();
    loadUserData();
    alert('Registration successful! Welcome to your business owner portal.');
  })
  .catch(err => {
    console.error('Registration error:', err);
    alert('Registration failed. Please try again.');
  });
}

function handleLogout() {
  // Call logout endpoint to remove token from database
  if (authToken) {
    authenticatedFetch(base + '/logout', {
      method: 'POST'
    })
    .then(() => {
      console.log('Logged out successfully');
    })
    .catch(err => {
      console.error('Logout error:', err);
    });
  }

  // Clear session data
  currentUser = null;
  authToken = null;
  sessionStorage.removeItem('authToken');
  sessionStorage.removeItem('currentUser');
  showNotAuthenticated();
}

// Helper function to make authenticated requests
function authenticatedFetch(url, options = {}) {
  if (!options.headers) {
    options.headers = {};
  }
  options.headers['Authorization'] = `Bearer ${authToken}`;
  return fetch(url, options);
}

function loadUserData() {
  if (!currentUser || !authToken) return;

  loadBusinesses();
  loadStats();
}

function loadBusinesses() {
  authenticatedFetch(base + '/my-businesses')
    .then(res => res.json())
    .then(businesses => {
      allBusinesses = Array.isArray(businesses) ? businesses : [];
      displayBusinesses();
    })
    .catch(err => {
      console.error('Error loading businesses:', err);
      document.getElementById('businesses-list').innerHTML = '<p>Error loading your businesses</p>';
    });
}

function displayBusinesses() {
  const list = document.getElementById('businesses-list');
  list.innerHTML = '';

  if (allBusinesses.length > 0) {
    allBusinesses.forEach(business => {
      const card = document.createElement('div');
      card.className = 'business-card';
      const rating = business.rating ? `${business.rating}★` : 'No rating';
      card.innerHTML = `
        <h3>${business.name}</h3>
        <span class="category">${business.category}</span>
        <div class="rating">${rating}</div>
        <div class="description">${business.description}</div>
        <div class="contact-info">
          <div><strong>Phone:</strong> <a href="tel:${business.phone}">${business.phone}</a></div>
          <div><strong>Email:</strong> <a href="mailto:${business.email}">${business.email}</a></div>
          <div><strong>Address:</strong> ${business.address}</div>
        </div>
        <div class="business-actions">
          <button onclick="toggleEdit(${business.id})" class="secondary">Edit</button>
          <button onclick="deleteBusiness(${business.id})" class="danger">Delete</button>
        </div>
        <div class="edit-form" id="edit-form-${business.id}">
          <h4>Edit Business</h4>
          <label>Business Name:</label>
          <input type="text" id="edit-name-${business.id}" value="${business.name}">

          <label>Category:</label>
          <select id="edit-category-${business.id}">
            <option value="">Select category</option>
            <option value="Restaurant">Restaurant</option>
            <option value="Retail">Retail</option>
            <option value="Services">Services</option>
            <option value="Healthcare">Healthcare</option>
            <option value="Technology">Technology</option>
            <option value="Entertainment">Entertainment</option>
            <option value="Other">Other</option>
          </select>

          <label>Description:</label>
          <textarea id="edit-description-${business.id}">${business.description}</textarea>

          <label>Phone:</label>
          <input type="tel" id="edit-phone-${business.id}" value="${business.phone}">

          <label>Email:</label>
          <input type="email" id="edit-email-${business.id}" value="${business.email}">

          <label>Address:</label>
          <input type="text" id="edit-address-${business.id}" value="${business.address}">

          <label>Rating (0-5):</label>
          <input type="number" id="edit-rating-${business.id}" value="${business.rating || 0}" min="0" max="5" step="0.1">

          <button onclick="saveEdit(${business.id})" class="success">Save Changes</button>
          <button onclick="cancelEdit(${business.id})" class="secondary">Cancel</button>
        </div>
      `;

      // Set selected category
      const categorySelect = card.querySelector(`#edit-category-${business.id}`);
      categorySelect.value = business.category;

      list.appendChild(card);
    });
  } else {
    list.innerHTML = '<div class="no-businesses">No businesses yet. Add your first business below!</div>';
  }
}

function loadStats() {
  // Load business stats from server
  authenticatedFetch(base + '/my-business-stats')
    .then(res => res.json())
    .then(stats => {
      document.getElementById('total-businesses').textContent = stats.business_count || 0;
      document.getElementById('avg-rating').textContent = (stats.average_rating || 0).toFixed(1);
      document.getElementById('total-views').textContent = stats.total_views || 0;
      document.getElementById('uptime').textContent = 'Personal Stats';
    })
    .catch(err => {
      console.error('Error loading stats:', err);
      document.getElementById('total-businesses').textContent = 'N/A';
      document.getElementById('avg-rating').textContent = 'N/A';
      document.getElementById('total-views').textContent = 'N/A';
      document.getElementById('uptime').textContent = 'N/A';
    });
}

function formatUptime(seconds) {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);
  return `${hours}h ${minutes}m ${secs}s`;
}

function toggleEdit(id) {
  const form = document.getElementById(`edit-form-${id}`);
  form.classList.toggle('show');
}

function handleAddBusiness(e) {
  e.preventDefault();

  const name = document.getElementById('business-name').value.trim();
  const category = document.getElementById('business-category').value.trim();
  const description = document.getElementById('business-description').value.trim();
  const phone = document.getElementById('business-phone').value.trim();
  const email = document.getElementById('business-email').value.trim();
  const address = document.getElementById('business-address').value.trim();
  const rating = parseFloat(document.getElementById('business-rating').value) || 0;

  if (!name || !category || !description || !phone || !email || !address) {
    alert('Please fill in all required fields');
    return;
  }

  if (rating < 0 || rating > 5) {
    alert('Rating must be between 0 and 5');
    return;
  }

  authenticatedFetch(base + '/businesses', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, category, description, phone, email, address, rating })
  })
    .then(res => res.json())
    .then(data => {
      // Clear form
      document.getElementById('business-name').value = '';
      document.getElementById('business-category').value = '';
      document.getElementById('business-description').value = '';
      document.getElementById('business-phone').value = '';
      document.getElementById('business-email').value = '';
      document.getElementById('business-address').value = '';
      document.getElementById('business-rating').value = '';

      alert('Business added successfully!');
      loadBusinesses();
    })
    .catch(err => {
      console.error('Error creating business:', err);
      alert('Error adding business: ' + err);
    });
}

function saveEdit(id) {
  const name = document.getElementById(`edit-name-${id}`).value.trim();
  const category = document.getElementById(`edit-category-${id}`).value.trim();
  const description = document.getElementById(`edit-description-${id}`).value.trim();
  const phone = document.getElementById(`edit-phone-${id}`).value.trim();
  const email = document.getElementById(`edit-email-${id}`).value.trim();
  const address = document.getElementById(`edit-address-${id}`).value.trim();
  const rating = parseFloat(document.getElementById(`edit-rating-${id}`).value) || 0;

  if (!name || !category || !description || !phone || !email || !address) {
    alert('Please fill in all required fields');
    return;
  }

  authenticatedFetch(base + '/businesses', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, name, category, description, phone, email, address, rating })
  })
    .then(res => res.json())
    .then(data => {
      alert('Business updated successfully!');
      loadBusinesses();
    })
    .catch(err => {
      console.error('Error updating business:', err);
      alert('Error updating business: ' + err);
    });
}

function cancelEdit(id) {
  const form = document.getElementById(`edit-form-${id}`);
  form.classList.remove('show');
  // Reset form values
  const business = allBusinesses.find(b => b.id === id);
  if (business) {
    document.getElementById(`edit-name-${id}`).value = business.name;
    document.getElementById(`edit-category-${id}`).value = business.category;
    document.getElementById(`edit-description-${id}`).value = business.description;
    document.getElementById(`edit-phone-${id}`).value = business.phone;
    document.getElementById(`edit-email-${id}`).value = business.email;
    document.getElementById(`edit-address-${id}`).value = business.address;
    document.getElementById(`edit-rating-${id}`).value = business.rating || 0;
  }
}

function deleteBusiness(id) {
  if (!confirm('Are you sure you want to delete this business listing?')) {
    return;
  }

  authenticatedFetch(base + '/businesses', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id })
  })
    .then(res => {
      if (res.ok) {
        alert('Business deleted successfully!');
        loadBusinesses();
      } else {
        alert('Error deleting business');
      }
    })
    .catch(err => {
      console.error('Error deleting business:', err);
      alert('Error deleting business: ' + err);
    });
}
