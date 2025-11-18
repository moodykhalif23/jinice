const base = 'http://localhost:8080';
let allBusinesses = [];

function loadBusinesses() {
  fetch(base + '/businesses')
    .then(res => res.json())
    .then(businesses => {
      allBusinesses = Array.isArray(businesses) ? businesses : [];
      filterBusinesses();
    })
    .catch(err => {
      console.error('Error loading businesses:', err);
      document.getElementById('businesses-list').innerHTML = '<p>Error loading businesses</p>';
    });
}

function filterBusinesses() {
  const searchTerm = document.getElementById('search-input').value.toLowerCase();
  const filterValue = document.getElementById('category-filter').value;
  const list = document.getElementById('businesses-list');
  list.innerHTML = '';

  const filteredBusinesses = allBusinesses.filter(business => {
    const matchesSearch = !searchTerm || 
      business.name.toLowerCase().includes(searchTerm) ||
      (business.description && business.description.toLowerCase().includes(searchTerm)) ||
      (business.address && business.address.toLowerCase().includes(searchTerm)) ||
      (business.category && business.category.toLowerCase().includes(searchTerm));
    
    const matchesCategory = !filterValue || business.category === filterValue;

    return matchesSearch && matchesCategory;
  });

  if (filteredBusinesses.length > 0) {
    filteredBusinesses.forEach(business => {
      const card = document.createElement('a');
      card.className = 'business-card';
      card.href = `business-detail.html?id=${business.id}`;
      const rating = business.rating ? `${business.rating}★` : 'No rating';
      
      // Handle multiple images or single image
      let imageHtml = '';
      let imagesToUse = [];
      
      // Check for multiple images in different possible formats
      if (business.images && Array.isArray(business.images) && business.images.length > 0) {
        imagesToUse = business.images;
      } else if (business.image_urls && Array.isArray(business.image_urls) && business.image_urls.length > 0) {
        imagesToUse = business.image_urls;
      } else if (business.image_url) {
        imagesToUse = [business.image_url];
      }
      
      if (imagesToUse.length > 0) {
        imageHtml = createCarouselHtml(imagesToUse);
      } else {
        imageHtml = '<div class="card-thumb" style="background-color: #e9ecef; display: flex; align-items: center; justify-content: center; color: #999;">No image</div>';
      }

      card.innerHTML = `
        ${imageHtml}
        <div class="card-body">
          <h3>${business.name}</h3>
          <span class="category">${business.category}</span>
          <div class="rating">${rating}</div>
          <div class="description">${business.description}</div>
          <div class="contact-info">
            <div><strong>📞</strong> ${business.phone || 'N/A'}</div>
            <div><strong>📧</strong> ${business.email || 'N/A'}</div>
            <div><strong>📍</strong> ${business.address}</div>
          </div>
        </div>
      `;
      list.appendChild(card);
    });

    // Initialize carousels after rendering
    setTimeout(initializeCarousels, 0);
  } else {
    const message = filterValue
      ? `No businesses found in "${filterValue}" category.`
      : 'No businesses in directory yet. Be the first to add one!';
    list.innerHTML = `<div class="no-businesses">${message}</div>`;
  }
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
  console.log('Initializing carousels, found:', carousels.length);
  
  carousels.forEach((carousel, idx) => {
    const images = Array.from(carousel.querySelectorAll('.card-carousel-image'));
    console.log(`Carousel ${idx}: ${images.length} images`);
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

function deleteBusiness(id) {
  if (!confirm('Are you sure you want to remove this listing?')) {
    return;
  }
  
  fetch(base + '/businesses', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id })
  })
    .then(res => {
      if (res.ok) loadBusinesses();
    })
    .catch(err => console.error('Error deleting business:', err));
}

document.addEventListener('DOMContentLoaded', () => {
  setTimeout(() => {
    loadBusinesses();
  }, 500);

  // Refresh button
  const refreshBtn = document.getElementById('btn-refresh-businesses');
  if (refreshBtn) {
    refreshBtn.addEventListener('click', loadBusinesses);
  }

  // Search input
  const searchInput = document.getElementById('search-input');
  if (searchInput) {
    searchInput.addEventListener('input', filterBusinesses);
  }

  // Category filter
  const categoryFilter = document.getElementById('category-filter');
  if (categoryFilter) {
    categoryFilter.addEventListener('change', filterBusinesses);
  }

  // Add Business
  const addBusinessBtn = document.getElementById('btn-add-business');
  if (addBusinessBtn) {
    addBusinessBtn.addEventListener('click', () => {
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

      fetch(base + '/businesses', {
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
          
          alert('Business listing added successfully!');
          loadBusinesses();
        })
        .catch(err => {
          console.error('Error creating business:', err);
          alert('Error adding business listing: ' + err);
        });
    });
  }
});
