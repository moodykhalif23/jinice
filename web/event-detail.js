const API_BASE = 'http://localhost:8080';

let currentEvent = null;

// Get event ID from URL
function getEventIdFromURL() {
  const params = new URLSearchParams(window.location.search);
  return params.get('id');
}

async function loadEventDetails() {
  const eventId = getEventIdFromURL();
  
  if (!eventId) {
    showError('No event ID provided');
    return;
  }

  try {
    const response = await fetch(`${API_BASE}/event/${eventId}`);
    
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error('Event not found');
      }
      throw new Error('Failed to load event details');
    }
    
    currentEvent = await response.json();
    displayEvent(currentEvent);
  } catch (error) {
    console.error('Error loading event:', error);
    showError(error.message || 'Failed to load event details. Please try again later.');
  }
}

function displayEvent(event) {
  document.getElementById('loading').style.display = 'none';
  document.getElementById('event-container').style.display = 'block';
  
  const eventDate = new Date(event.event_date);
  const formattedDate = eventDate.toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
  const formattedTime = eventDate.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit'
  });
  
  const isPast = eventDate < new Date();
  const priceDisplay = event.price > 0 
    ? `$${event.price.toFixed(2)}`
    : 'FREE';
  const priceClass = event.price > 0 ? 'meta-price' : 'meta-price free';

  // Handle images
  let imagesHtml = '';
  let imagesToUse = [];
  
  if (event.images && Array.isArray(event.images) && event.images.length > 0) {
    imagesToUse = event.images;
  } else if (event.image_urls && Array.isArray(event.image_urls) && event.image_urls.length > 0) {
    imagesToUse = event.image_urls;
  } else if (event.image_url) {
    imagesToUse = [event.image_url];
  }
  
  if (imagesToUse.length > 0) {
    imagesHtml = `
      <div class="detail-images">
        ${imagesToUse.map((img, idx) => `
          <img src="${img}" alt="${escapeHtml(event.title)} - Image ${idx + 1}" class="detail-image">
        `).join('')}
      </div>
    `;
  } else {
    imagesHtml = `
      <div class="detail-images">
        <div class="no-image-placeholder">No images available</div>
      </div>
    `;
  }

  let businessSection = '';
  if (event.business_id) {
    businessSection = `
      <div class="business-link">
        <strong>🏢 Hosted by Business:</strong> 
        <a href="business-detail.html?id=${event.business_id}">View Business Details</a>
      </div>
    `;
  }

  const container = document.getElementById('event-container');
  container.innerHTML = `
    <div class="event-container">
      <div class="detail-layout">
        <div class="detail-left">
          ${imagesHtml}
        </div>
        
        <div class="detail-right">
          <div class="event-header">
            <h1 class="event-title">${escapeHtml(event.title)}</h1>
            <div class="event-meta">
              <span class="meta-badge meta-date">📅 ${formattedDate} at ${formattedTime}</span>
              ${event.category ? `<span class="meta-badge meta-category">${escapeHtml(event.category)}</span>` : ''}
              <span class="meta-badge ${priceClass}">💰 ${priceDisplay}</span>
            </div>
            ${isPast ? '<p style="color: #dc3545; font-weight: bold;">⚠️ This event has already passed</p>' : ''}
          </div>

          ${businessSection}

          <div class="event-section">
            <h2>📝 Description</h2>
            <div class="event-description">${escapeHtml(event.description || 'No description available')}</div>
          </div>

          <div class="event-section">
            <h2>ℹ️ Event Information</h2>
            <ul class="event-info-list">
              ${event.location ? `
                <li class="info-item">
                  <span class="info-icon">📍</span>
                  <div class="info-content">
                    <div class="info-label">Location</div>
                    <div class="info-value">${escapeHtml(event.location)}</div>
                  </div>
                </li>
              ` : ''}
              <li class="info-item">
                <span class="info-icon">📅</span>
                <div class="info-content">
                  <div class="info-label">Date</div>
                  <div class="info-value">${formattedDate}</div>
                </div>
              </li>
              <li class="info-item">
                <span class="info-icon">🕐</span>
                <div class="info-content">
                  <div class="info-label">Time</div>
                  <div class="info-value">${formattedTime}</div>
                </div>
              </li>
              <li class="info-item">
                <span class="info-icon">💵</span>
                <div class="info-content">
                  <div class="info-label">Price</div>
                  <div class="info-value">${priceDisplay}</div>
                </div>
              </li>
            </ul>
          </div>

          ${!isPast ? `
            <div class="booking-section">
              <h2>🎟️ Book Your Spot</h2>
              <p>Interested in attending this event? Fill out the form below to reserve your spot!</p>
              
              <div id="booking-message"></div>
              
              <form id="booking-form" class="booking-form">
                <div class="form-group">
                  <label for="booking-name">Full Name *</label>
                  <input type="text" id="booking-name" required placeholder="Enter your full name">
                </div>

                <div class="form-group">
                  <label for="booking-email">Email Address *</label>
                  <input type="email" id="booking-email" required placeholder="your.email@example.com">
                </div>

                <div class="form-group">
                  <label for="booking-phone">Phone Number</label>
                  <input type="tel" id="booking-phone" placeholder="+1 (555) 123-4567">
                </div>

                <div class="form-group">
                  <label for="booking-tickets">Number of Tickets *</label>
                  <input type="number" id="booking-tickets" min="1" max="10" value="1" required>
                </div>

                <div class="form-group">
                  <label for="booking-notes">Additional Notes</label>
                  <textarea id="booking-notes" placeholder="Any special requirements or questions?"></textarea>
                </div>

                <button type="submit" class="success">Submit Booking Request</button>
              </form>
            </div>
          ` : ''}
        </div>
      </div>
    </div>
  `;

  // Add form submit handler if event is not past
  if (!isPast) {
    document.getElementById('booking-form').addEventListener('submit', handleBooking);
  }
}

async function handleBooking(e) {
  e.preventDefault();
  
  const name = document.getElementById('booking-name').value.trim();
  const email = document.getElementById('booking-email').value.trim();
  const phone = document.getElementById('booking-phone').value.trim();
  const tickets = parseInt(document.getElementById('booking-tickets').value);
  const notes = document.getElementById('booking-notes').value.trim();

  if (!name || !email || !tickets) {
    showBookingMessage('Please fill in all required fields', 'error');
    return;
  }

  // Disable submit button
  const submitBtn = document.querySelector('#booking-form button[type="submit"]');
  submitBtn.disabled = true;
  submitBtn.textContent = 'Submitting...';

  try {
    const response = await fetch(`${API_BASE}/bookings`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        event_id: currentEvent.id,
        name,
        email,
        phone,
        tickets,
        notes
      })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to submit booking');
    }

    const result = await response.json();
    
    // Show success message
    showBookingMessage(
      'Booking request submitted successfully! You will receive a confirmation email shortly.',
      'success'
    );
    
    // Clear form
    document.getElementById('booking-form').reset();
    
  } catch (error) {
    console.error('Booking error:', error);
    showBookingMessage(
      'Failed to submit booking: ' + error.message,
      'error'
    );
  } finally {
    // Re-enable submit button
    submitBtn.disabled = false;
    submitBtn.textContent = 'Submit Booking Request';
  }
}

function showBookingMessage(message, type) {
  const container = document.getElementById('booking-message');
  const className = type === 'success' ? 'success-message' : 'error';
  container.innerHTML = `<div class="${className}">${escapeHtml(message)}</div>`;
  
  // Scroll to message
  container.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  
  // Auto-hide success messages after 5 seconds
  if (type === 'success') {
    setTimeout(() => {
      container.innerHTML = '';
    }, 5000);
  }
}

function showError(message) {
  document.getElementById('loading').style.display = 'none';
  document.getElementById('error-container').innerHTML = `
    <div class="error">
      <strong>Error:</strong> ${escapeHtml(message)}
      <br><br>
      <a href="events.html">← Back to Events</a>
    </div>
  `;
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Load event details on page load
loadEventDetails();
