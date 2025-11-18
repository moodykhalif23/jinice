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

    // Handle multiple images or single image
    let imageHtml = '';
    if (event.images && event.images.length > 0) {
      imageHtml = createCarouselHtml(event.images);
    } else if (event.image_url) {
      imageHtml = `<img class="card-thumb" src="${event.image_url}" alt="${escapeHtml(event.title)} thumbnail" loading="lazy">`;
    } else {
      imageHtml = '<div class="card-thumb" style="background-color: #e9ecef; display: flex; align-items: center; justify-content: center; color: #999;">No image</div>';
    }

    return `
      <a class="event-card" href="event-detail.html?id=${event.id}">
        ${imageHtml}
        <div class="card-body">
          <h3>${escapeHtml(event.title)}</h3>
          <div style="margin-bottom: 8px;">
            <span class="event-date">📅 ${formattedDate}</span>
            ${event.category ? `<span class="event-category">${escapeHtml(event.category)}</span>` : ''}
          </div>
          <div class="event-description">${escapeHtml(event.description || 'No description available')}</div>
          <div class="event-details">
            ${event.location ? `<div><strong>📍</strong> ${escapeHtml(event.location)}</div>` : ''}
          </div>
          ${price}
        </div>
      </a>
    `;
  }).join('');

  // Initialize carousels after rendering
  setTimeout(initializeCarousels, 0);
}

function createCarouselHtml(images) {
  if (!images || images.length === 0) {
    return '<div class="card-thumb" style="background-color: #e9ecef; display: flex; align-items: center; justify-content: center; color: #999;">No image</div>';
  }

  if (images.length === 1) {
    return `<img class="card-thumb" src="${images[0]}" alt="Image" loading="lazy">`;
  }

  const carouselId = 'carousel-' + Math.random().toString(36).substr(2, 9);
  
  return `
    <div class="card-image-carousel" data-carousel-id="${carouselId}">
      <div class="card-carousel-track">
        ${images.map(img => `<img class="card-carousel-image" src="${img}" alt="Image" loading="lazy">`).join('')}
      </div>
      <button class="card-carousel-btn prev" aria-label="Previous image">‹</button>
      <button class="card-carousel-btn next" aria-label="Next image">›</button>
      <div class="card-carousel-indicators">
        ${images.map((_, i) => `<div class="card-carousel-dot ${i === 0 ? 'active' : ''}"></div>`).join('')}
      </div>
    </div>
  `;
}

function initializeCarousels() {
  const carousels = document.querySelectorAll('.card-image-carousel');
  carousels.forEach(carousel => {
    const images = Array.from(carousel.querySelectorAll('.card-carousel-image'));
    if (images.length <= 1) return;

    let currentIndex = 0;
    const track = carousel.querySelector('.card-carousel-track');
    const prevBtn = carousel.querySelector('.card-carousel-btn.prev');
    const nextBtn = carousel.querySelector('.card-carousel-btn.next');
    const dots = carousel.querySelectorAll('.card-carousel-dot');

    const updateCarousel = () => {
      track.style.transform = `translateX(-${currentIndex * 100}%)`;
      dots.forEach((dot, i) => {
        dot.classList.toggle('active', i === currentIndex);
      });
    };

    const goToNext = () => {
      currentIndex = (currentIndex + 1) % images.length;
      updateCarousel();
    };

    const goToPrev = () => {
      currentIndex = (currentIndex - 1 + images.length) % images.length;
      updateCarousel();
    };

    prevBtn.addEventListener('click', (e) => {
      e.preventDefault();
      e.stopPropagation();
      goToPrev();
    });

    nextBtn.addEventListener('click', (e) => {
      e.preventDefault();
      e.stopPropagation();
      goToNext();
    });

    dots.forEach((dot, i) => {
      dot.addEventListener('click', (e) => {
        e.preventDefault();
        e.stopPropagation();
        currentIndex = i;
        updateCarousel();
      });
    });

    // Auto-play
    let autoPlayInterval = setInterval(goToNext, 4000);

    carousel.addEventListener('mouseenter', () => {
      clearInterval(autoPlayInterval);
    });

    carousel.addEventListener('mouseleave', () => {
      autoPlayInterval = setInterval(goToNext, 4000);
    });
  });
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
