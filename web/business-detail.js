const base = 'http://localhost:8080';

function getBusinessIdFromUrl() {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get('id');
}

function loadBusinessDetails() {
  const businessId = getBusinessIdFromUrl();

  if (!businessId) {
    showError('Business ID not found in URL');
    return;
  }

  // Validate that ID is a number
  const id = parseInt(businessId, 10);
  if (isNaN(id)) {
    showError('Invalid business ID');
    return;
  }

  fetch(`${base}/business/${id}`)
    .then(res => {
      if (!res.ok) {
        if (res.status === 404) {
          throw new Error('Business not found');
        }
        throw new Error(`HTTP error! status: ${res.status}`);
      }
      return res.json();
    })
    .then(business => {
      displayBusinessDetails(business);
    })
    .catch(err => {
      console.error('Error loading business details:', err);
      showError(err.message || 'Error loading business details');
    });
}

function displayBusinessDetails(business) {
  const content = document.getElementById('content');
  const formattedDate = new Date(business.created_at).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });

  content.innerHTML = `
    <div class="business-details">
      <div class="business-header">
        <div class="business-title">
          <h2>${business.name}</h2>
          <span class="category-badge">${business.category}</span>
        </div>
        <div class="business-rating">
          ${business.rating || 'No rating'}★
        </div>
      </div>

      <div class="business-description">
        ${business.description || 'No description available.'}
      </div>

      <div class="contact-section">
        <h3>Contact Information</h3>
        <div class="contact-item">
          <span class="contact-label">Phone:</span>
          <span class="contact-value">
            ${business.phone ? `<a href="tel:${business.phone}">${business.phone}</a>` : 'Not provided'}
          </span>
        </div>
        <div class="contact-item">
          <span class="contact-label">Email:</span>
          <span class="contact-value">
            ${business.email ? `<a href="mailto:${business.email}">${business.email}</a>` : 'Not provided'}
          </span>
        </div>
        <div class="contact-item">
          <span class="contact-label">Address:</span>
          <span class="contact-value">${business.address || 'Not provided'}</span>
        </div>
        <div class="contact-item">
          <span class="contact-label">Added:</span>
          <span class="contact-value">${formattedDate}</span>
        </div>
      </div>
    </div>
  `;
}

function showError(message) {
  const content = document.getElementById('content');
  content.innerHTML = `
    <div class="error">
      <h3>⚠️ Error</h3>
      <p>${message}</p>
      <p><a href="user.html">← Back to Directory</a></p>
    </div>
  `;
}

document.addEventListener('DOMContentLoaded', () => {
  loadBusinessDetails();
});
