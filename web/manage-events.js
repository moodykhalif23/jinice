const API_BASE = 'http://localhost:8080';

let authToken = sessionStorage.getItem('authToken');
let currentUser = null;
let myBusinesses = [];
let myEvents = [];

// Check authentication on page load
async function checkAuth() {
  if (!authToken) {
    showAuthRequired();
    return false;
  }

  try {
    const response = await fetch(`${API_BASE}/my-events`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    if (!response.ok) {
      if (response.status === 401 || response.status === 403) {
        sessionStorage.removeItem('authToken');
        authToken = null;
        showAuthRequired();
        return false;
      }
      throw new Error('Failed to verify authentication');
    }

    showDashboard();
    return true;
  } catch (error) {
    console.error('Auth check failed:', error);
    showAuthRequired();
    return false;
  }
}

function showAuthRequired() {
  document.getElementById('auth-required').style.display = 'block';
  document.getElementById('dashboard').style.display = 'none';
  document.getElementById('logout-nav').style.display = 'none';
}

function showDashboard() {
  document.getElementById('auth-required').style.display = 'none';
  document.getElementById('dashboard').style.display = 'block';
  document.getElementById('logout-nav').style.display = 'inline-block';
  
  loadMyBusinesses();
  loadMyEvents();
  hidePortalLinkIfPortalUser();
}

async function loadMyBusinesses() {
  try {
    const response = await fetch(`${API_BASE}/my-businesses`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    if (!response.ok) {
      // Event owners may not have businesses - that's okay
      if (response.status === 403) {
        myBusinesses = [];
        populateBusinessSelect();
        return;
      }
      throw new Error('Failed to load businesses');
    }
    
    myBusinesses = await response.json();
    populateBusinessSelect();
  } catch (error) {
    console.error('Error loading businesses:', error);
    myBusinesses = [];
    populateBusinessSelect();
  }
}

function populateBusinessSelect() {
  const select = document.getElementById('event-business');
  
  if (!myBusinesses || myBusinesses.length === 0) {
    select.innerHTML = '<option value="">No business (standalone event)</option>';
    return;
  }

  select.innerHTML = '<option value="">No business (standalone event)</option>' +
    myBusinesses.map(b => `<option value="${b.id}">${escapeHtml(b.name)}</option>`).join('');
}

async function loadMyEvents() {
  try {
    const response = await fetch(`${API_BASE}/my-events`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    if (!response.ok) throw new Error('Failed to load events');
    
    myEvents = await response.json();
    displayEvents(myEvents);
  } catch (error) {
    console.error('Error loading events:', error);
    document.getElementById('events-list').innerHTML = 
      '<div class="no-events">Failed to load events.</div>';
  }
}

function displayEvents(events) {
  const container = document.getElementById('events-list');
  
  if (!events || events.length === 0) {
    container.innerHTML = '<div class="no-events">No events yet. Create your first event below!</div>';
    return;
  }

  container.innerHTML = events.map(event => {
    const eventDate = new Date(event.event_date);
    const formattedDate = eventDate.toLocaleDateString('en-US', {
      weekday: 'short',
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
    
    const isPast = eventDate < new Date();
    const business = event.business_id ? myBusinesses.find(b => b.id === event.business_id) : null;
    const businessName = business ? business.name : (event.business_id ? `Business #${event.business_id}` : 'Standalone Event');

    return `
      <div class="event-card" style="${isPast ? 'opacity: 0.6;' : ''}">
        <h3>${escapeHtml(event.title)} ${isPast ? '(Past)' : ''}</h3>
        <div>
          <span class="event-date">📅 ${formattedDate}</span>
          ${event.category ? `<span class="event-category">${escapeHtml(event.category)}</span>` : ''}
        </div>
        <div class="event-description">${escapeHtml(event.description || 'No description')}</div>
        <div class="event-details">
          <div><strong>🏢 ${event.business_id ? 'Business' : 'Type'}:</strong> ${escapeHtml(businessName)}</div>
          ${event.location ? `<div><strong>📍 Location:</strong> ${escapeHtml(event.location)}</div>` : ''}
          <div><strong>💰 Price:</strong> ${event.price > 0 ? `$${event.price.toFixed(2)}` : 'FREE'}</div>
        </div>
        <div class="event-actions">
          <button onclick="editEvent(${event.id})" class="secondary">Edit</button>
          <button onclick="deleteEvent(${event.id}, '${escapeHtml(event.title)}')" class="danger">Delete</button>
        </div>
        <div id="edit-form-${event.id}" class="edit-form"></div>
      </div>
    `;
  }).join('');
}

async function createEvent() {
  const businessIdStr = document.getElementById('event-business').value;
  const businessId = businessIdStr ? parseInt(businessIdStr) : null;
  const title = document.getElementById('event-title').value.trim();
  const category = document.getElementById('event-category').value;
  const description = document.getElementById('event-description').value.trim();
  const eventDate = document.getElementById('event-date').value;
  const location = document.getElementById('event-location').value.trim();
  const price = parseFloat(document.getElementById('event-price').value) || 0;

  if (!title || !eventDate) {
    alert('Please fill in all required fields (Title, Date & Time)');
    return;
  }

  try {
    const response = await fetch(`${API_BASE}/business-events`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authToken}`
      },
      body: JSON.stringify({
        business_id: businessId,
        title,
        category,
        description,
        event_date: eventDate,
        location,
        price
      })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to create event');
    }

    const created = await response.json();
    const eventId = created.id;

    // If images were selected, upload them
    const imageInput = document.getElementById('event-images');
    if (imageInput && imageInput.files && imageInput.files.length > 0) {
      await uploadFiles(imageInput.files, 'event', eventId);
    }

    alert('Event created successfully!');
    
    // Clear form
    document.getElementById('event-title').value = '';
    document.getElementById('event-category').value = '';
    document.getElementById('event-description').value = '';
    document.getElementById('event-date').value = '';
    document.getElementById('event-location').value = '';
    document.getElementById('event-price').value = '';
    document.getElementById('event-business').value = '';

    loadMyEvents();
  } catch (error) {
    console.error('Error creating event:', error);
    alert('Failed to create event: ' + error.message);
  }
}

async function uploadFiles(fileList, entityType, entityId) {
  const files = Array.from(fileList);
  for (const file of files) {
    try {
      const formData = new FormData();
      formData.append('image', file);
      formData.append('entity_type', entityType);
      formData.append('entity_id', String(entityId));

      const res = await fetch(`${API_BASE}/images/upload`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${authToken}`
        },
        body: formData
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        console.error('Image upload failed:', err);
      }
    } catch (err) {
      console.error('Upload error:', err);
    }
  }
}

function hidePortalLinkIfPortalUser() {
  try {
    const raw = sessionStorage.getItem('currentUser');
    if (!raw) return;
    const user = JSON.parse(raw);
    if (user && (user.type === 'business_owner' || user.type === 'event_owner')) {
      document.querySelectorAll('.nav a[href="auth.html"]').forEach(a => a.remove());
    }
  } catch (e) {
    // ignore
  }
}

function restorePortalLink() {
  const nav = document.querySelector('.nav');
  if (!nav) return;
  
  // Check if Portal Access link already exists
  const existing = nav.querySelector('a[href="auth.html"]');
  if (existing) return;
  
  // Create and insert Portal Access link
  const link = document.createElement('a');
  link.href = 'auth.html';
  link.textContent = 'Portal Access';
  nav.appendChild(link);
}

function editEvent(eventId) {
  const event = myEvents.find(e => e.id === eventId);
  if (!event) return;

  const formContainer = document.getElementById(`edit-form-${eventId}`);
  const isVisible = formContainer.classList.contains('show');

  // Hide all edit forms
  document.querySelectorAll('.edit-form').forEach(f => f.classList.remove('show'));

  if (isVisible) return;

  const eventDate = new Date(event.event_date);
  const formattedDate = eventDate.toISOString().slice(0, 16);

  formContainer.innerHTML = `
    <h4>Edit Event</h4>
    <label>Title:</label>
    <input type="text" id="edit-title-${eventId}" value="${escapeHtml(event.title)}">
    
    <label>Category:</label>
    <select id="edit-category-${eventId}">
      <option value="">Select a category</option>
      <option value="Workshop" ${event.category === 'Workshop' ? 'selected' : ''}>Workshop</option>
      <option value="Concert" ${event.category === 'Concert' ? 'selected' : ''}>Concert</option>
      <option value="Sale" ${event.category === 'Sale' ? 'selected' : ''}>Sale</option>
      <option value="Class" ${event.category === 'Class' ? 'selected' : ''}>Class</option>
      <option value="Meetup" ${event.category === 'Meetup' ? 'selected' : ''}>Meetup</option>
      <option value="Other" ${event.category === 'Other' ? 'selected' : ''}>Other</option>
    </select>
    
    <label>Description:</label>
    <textarea id="edit-description-${eventId}">${escapeHtml(event.description || '')}</textarea>
    
    <label>Date & Time:</label>
    <input type="datetime-local" id="edit-date-${eventId}" value="${formattedDate}">
    
    <label>Location:</label>
    <input type="text" id="edit-location-${eventId}" value="${escapeHtml(event.location || '')}">
    
    <label>Price:</label>
    <input type="number" id="edit-price-${eventId}" value="${event.price}" min="0" step="0.01">
    
    <button onclick="saveEvent(${eventId})" class="success">Save Changes</button>
    <button onclick="cancelEdit(${eventId})" class="secondary">Cancel</button>
  `;
  
  formContainer.classList.add('show');
}

function cancelEdit(eventId) {
  document.getElementById(`edit-form-${eventId}`).classList.remove('show');
}

async function saveEvent(eventId) {
  const title = document.getElementById(`edit-title-${eventId}`).value.trim();
  const category = document.getElementById(`edit-category-${eventId}`).value;
  const description = document.getElementById(`edit-description-${eventId}`).value.trim();
  const eventDate = document.getElementById(`edit-date-${eventId}`).value;
  const location = document.getElementById(`edit-location-${eventId}`).value.trim();
  const price = parseFloat(document.getElementById(`edit-price-${eventId}`).value) || 0;

  if (!title || !eventDate) {
    alert('Title and Date are required');
    return;
  }

  try {
    const response = await fetch(`${API_BASE}/business-events`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authToken}`
      },
      body: JSON.stringify({
        id: eventId,
        title,
        category,
        description,
        event_date: eventDate,
        location,
        price
      })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to update event');
    }

    alert('Event updated successfully!');
    cancelEdit(eventId);
    loadMyEvents();
  } catch (error) {
    console.error('Error updating event:', error);
    alert('Failed to update event: ' + error.message);
  }
}

async function deleteEvent(eventId, eventTitle) {
  if (!confirm(`Are you sure you want to delete "${eventTitle}"?`)) {
    return;
  }

  try {
    const response = await fetch(`${API_BASE}/business-events`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authToken}`
      },
      body: JSON.stringify({ id: eventId })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to delete event');
    }

    alert('Event deleted successfully!');
    loadMyEvents();
  } catch (error) {
    console.error('Error deleting event:', error);
    alert('Failed to delete event: ' + error.message);
  }
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Event listeners
document.getElementById('btn-create-event').addEventListener('click', createEvent);
document.getElementById('btn-refresh-events').addEventListener('click', loadMyEvents);
document.getElementById('logout-nav').addEventListener('click', () => {
  // Call logout endpoint
  const token = sessionStorage.getItem('authToken');
  if (token) {
    fetch(`${API_BASE}/logout`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    .then(() => {
      console.log('Logged out successfully');
    })
    .catch(err => {
      console.error('Logout error:', err);
    });
  }
  
  sessionStorage.removeItem('authToken');
  sessionStorage.removeItem('currentUser');
  
  // Restore Portal Access link
  restorePortalLink();
  
  // Redirect
  window.location.href = 'manage-events.html';
});

// Initialize
checkAuth();
