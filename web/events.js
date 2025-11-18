const API_BASE = 'http://localhost:8080';

let allEvents = [];

async function loadEvents() {
  try {
    const response = await fetch(`${API_BASE}/business-events`);
    if (!response.ok) throw new Error('Failed to load events');
    
    allEvents = await response.json();
    displayEvents(allEvents);
  } catch (error) {
    console.error('Error loading events:', error);
    document.getElementById('events-list').innerHTML = 
      '<div class="no-events">Failed to load events. Please try again later.</div>';
  }
}

function displayEvents(events) {
  const container = document.getElementById('events-list');
  
  if (!events || events.length === 0) {
    container.innerHTML = '<div class="no-events">No upcoming events found.</div>';
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
    
    const price = event.price > 0 
      ? `<span class="event-price">$${event.price.toFixed(2)}</span>`
      : `<span class="event-price free">FREE</span>`;

    const thumbHtml = event.image_url ? `<img class="card-thumb" src="${event.image_url}" alt="${escapeHtml(event.title)} thumbnail" loading="lazy">` : '';

    return `
      <a class="event-card" href="event-detail.html?id=${event.id}">
        ${thumbHtml}
        <h3>${escapeHtml(event.title)}</h3>
        <div>
          <span class="event-date">📅 ${formattedDate}</span>
          ${event.category ? `<span class="event-category">${escapeHtml(event.category)}</span>` : ''}
        </div>
        <div class="event-description">${escapeHtml(event.description || 'No description available')}</div>
        <div class="event-details">
          ${event.location ? `<div><strong>📍 Location:</strong> ${escapeHtml(event.location)}</div>` : ''}
          ${event.business_id ? `<div><strong>🏢 Business ID:</strong> ${event.business_id}</div>` : ''}
        </div>
        ${price}
      </a>
    `;
  }).join('');
}

function filterEvents() {
  const searchTerm = document.getElementById('search-input').value.toLowerCase();
  const category = document.getElementById('category-filter').value;

  const filtered = allEvents.filter(event => {
    const matchesSearch = !searchTerm || 
      event.title.toLowerCase().includes(searchTerm) ||
      (event.description && event.description.toLowerCase().includes(searchTerm)) ||
      (event.location && event.location.toLowerCase().includes(searchTerm));
    
    const matchesCategory = !category || event.category === category;

    return matchesSearch && matchesCategory;
  });

  displayEvents(filtered);
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Event listeners
document.getElementById('search-input').addEventListener('input', filterEvents);
document.getElementById('category-filter').addEventListener('change', filterEvents);

// Load events on page load
loadEvents();
