const base = 'http://localhost:8080';

function loadBusinesses() {
  fetch(base + '/businesses')
    .then(res => res.json())
    .then(businesses => {
      const list = document.getElementById('businesses-list');
      list.innerHTML = '';
      if (Array.isArray(businesses) && businesses.length > 0) {
        businesses.forEach(business => {
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
            <div>
              <button onclick="deleteBusiness(${business.id})" class="danger">Remove Listing</button>
            </div>
          `;
          list.appendChild(card);
        });
      } else {
        list.innerHTML = '<div class="no-businesses">No businesses in directory yet. Be the first to add one!</div>';
      }
    })
    .catch(err => {
      console.error('Error loading businesses:', err);
      document.getElementById('businesses-list').innerHTML = '<p>Error loading businesses</p>';
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
