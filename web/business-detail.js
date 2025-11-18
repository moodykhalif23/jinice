const base = 'http://localhost:8080';

function getBusinessIdFromUrl() {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get('id');
}

async function loadBusinessDetails() {
  const businessId = getBusinessIdFromUrl();

  if (!businessId) {
    showError('No business ID provided');
    return;
  }

  // Validate that ID is a number
  const id = parseInt(businessId, 10);
  if (isNaN(id)) {
    showError('Invalid business ID');
    return;
  }

  try {
    const response = await fetch(`${base}/business/${id}`);
    
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error('Business not found');
      }
      throw new Error('Failed to load business details');
    }
    
    const business = await response.json();
    displayBusinessDetails(business);
  } catch (error) {
    console.error('Error loading business:', error);
    showError(error.message || 'Failed to load business details. Please try again later.');
  }
}

function displayBusinessDetails(business) {
  document.getElementById('loading').style.display = 'none';
  document.getElementById('content').style.display = 'block';
  
  const content = document.getElementById('content');
  const formattedDate = new Date(business.created_at).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });

  // Handle images
  let imagesHtml = '';
  let imagesToUse = [];
  
  if (business.images && Array.isArray(business.images) && business.images.length > 0) {
    imagesToUse = business.images;
  } else if (business.image_urls && Array.isArray(business.image_urls) && business.image_urls.length > 0) {
    imagesToUse = business.image_urls;
  } else if (business.image_url) {
    imagesToUse = [business.image_url];
  }
  
  if (imagesToUse.length > 0) {
    imagesHtml = `
      <div class="detail-images">
        ${imagesToUse.map((img, idx) => `
          <img src="${img}" alt="${business.name} - Image ${idx + 1}" class="detail-image">
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

  content.innerHTML = `
    <div class="business-container">
      <div class="detail-layout">
        <div class="detail-left">
          ${imagesHtml}
        </div>
        
        <div class="detail-right">
          <div class="business-header">
            <h1 class="business-title">${business.name}</h1>
            <div class="business-meta">
              <span class="meta-badge meta-category">${business.category}</span>
              ${business.rating ? `<span class="meta-badge meta-rating">⭐ ${business.rating}</span>` : ''}
            </div>
          </div>

          <div class="business-section">
            <h2>📝 Description</h2>
            <div class="business-description">${business.description || 'No description available.'}</div>
          </div>

          <div class="business-section">
            <h2>📞 Contact Information</h2>
            <ul class="contact-info-list">
              ${business.phone ? `
                <li class="contact-item">
                  <span class="contact-icon">📱</span>
                  <div class="contact-content">
                    <div class="contact-label">Phone</div>
                    <div class="contact-value"><a href="tel:${business.phone}">${business.phone}</a></div>
                  </div>
                </li>
              ` : ''}
              ${business.email ? `
                <li class="contact-item">
                  <span class="contact-icon">✉️</span>
                  <div class="contact-content">
                    <div class="contact-label">Email</div>
                    <div class="contact-value"><a href="mailto:${business.email}">${business.email}</a></div>
                  </div>
                </li>
              ` : ''}
              ${business.address ? `
                <li class="contact-item">
                  <span class="contact-icon">📍</span>
                  <div class="contact-content">
                    <div class="contact-label">Address</div>
                    <div class="contact-value">${business.address}</div>
                  </div>
                </li>
              ` : ''}
              <li class="contact-item">
                <span class="contact-icon">📅</span>
                <div class="contact-content">
                  <div class="contact-label">Member Since</div>
                  <div class="contact-value">${formattedDate}</div>
                </div>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  `;
}

function showError(message) {
  document.getElementById('loading').style.display = 'none';
  document.getElementById('error-container').innerHTML = `
    <div class="error">
      <strong>Error:</strong> ${escapeHtml(message)}
      <br><br>
      <a href="user.html">← Back to Businesses</a>
    </div>
  `;
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Load business details on page load
loadBusinessDetails();
