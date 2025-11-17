const API_BASE = 'http://localhost:8080';

document.getElementById('register-form').addEventListener('submit', async (e) => {
  e.preventDefault();

  const name = document.getElementById('register-name').value.trim();
  const email = document.getElementById('register-email').value.trim();
  const password = document.getElementById('register-password').value;
  const organization = document.getElementById('register-organization').value.trim();
  const phone = document.getElementById('register-phone').value.trim();

  if (!name || !email || !password) {
    alert('Please fill in all required fields');
    return;
  }

  try {
    const response = await fetch(`${API_BASE}/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        name,
        email,
        password,
        type: 'event_owner',
        company: organization, // Using company field for organization
        phone
      })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Registration failed');
    }

    const data = await response.json();
    
    // Store auth token
    localStorage.setItem('authToken', data.token);
    
    alert('Registration successful! Redirecting to event management...');
    window.location.href = 'manage-events.html';
  } catch (error) {
    console.error('Registration error:', error);
    alert('Registration failed: ' + error.message);
  }
});
