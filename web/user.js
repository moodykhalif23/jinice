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
  const filterValue = document.getElementById('category-filter').value;
  const list = document.getElementById('businesses-list');
  list.innerHTML = '';

  const filteredBusinesses = filterValue
    ? allBusinesses.filter(business => business.category === filterValue)
    : allBusinesses;

    if (filteredBusinesses.length > 0) {
    filteredBusinesses.forEach(business => {
      const card = document.createElement('a');
      card.className = 'business-card';
      card.href = `business-detail.html?id=${business.id}`;
      const rating = business.rating ? `${business.rating}★` : 'No rating';
      const thumbHtml = business.image_url ? `<img class="card-thumb" src="${business.image_url}" alt="${business.name} thumbnail" loading="lazy">` : '';
      card.innerHTML = `
        ${thumbHtml}
        <h3>${business.name}</h3>
        <span class="category">${business.category}</span>
        <div class="rating">${rating}</div>
        <div class="description">${business.description}</div>
        <div class="contact-info">
          <div><strong>Phone:</strong> <span>${business.phone || 'N/A'}</span></div>
          <div><strong>Email:</strong> <span>${business.email || 'N/A'}</span></div>
          <div><strong>Address:</strong> ${business.address}</div>
        </div>
      `;
      list.appendChild(card);
    });
  } else {
    const message = filterValue
      ? `No businesses found in "${filterValue}" category.`
      : 'No businesses in directory yet. Be the first to add one!';
    list.innerHTML = `<div class="no-businesses">${message}</div>`;
  }
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
