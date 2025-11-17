// Image Manager Component for Business Directory
// Handles image upload, display, and management for businesses and events

class ImageManager {
    constructor(entityType, entityId, containerId, options = {}) {
        this.entityType = entityType; // 'business' or 'event'
        this.entityId = entityId;
        this.container = document.getElementById(containerId);
        this.options = {
            allowUpload: options.allowUpload !== false,
            allowDelete: options.allowDelete !== false,
            allowReorder: options.allowReorder !== false,
            maxImages: options.maxImages || 10,
            ...options
        };
        this.images = [];
        this.init();
    }

    async init() {
        if (!this.container) {
            console.error('Image manager container not found');
            return;
        }
        await this.loadImages();
        this.render();
    }

    async loadImages() {
        try {
            const response = await fetch(`/images?entity_type=${this.entityType}&entity_id=${this.entityId}`);
            if (response.ok) {
                this.images = await response.json();
            }
        } catch (error) {
            console.error('Error loading images:', error);
        }
    }

    render() {
        this.container.innerHTML = `
            <div class="image-manager">
                <div class="image-gallery">
                    ${this.renderImages()}
                </div>
                ${this.options.allowUpload ? this.renderUploadSection() : ''}
            </div>
        `;
        this.attachEventListeners();
    }

    renderImages() {
        if (!this.images || this.images.length === 0) {
            return '<p class="no-images">No images yet</p>';
        }

        return this.images.map((img, index) => `
            <div class="image-item ${img.is_primary ? 'primary' : ''}" data-image-id="${img.id}">
                <img src="${img.image_url}" alt="${img.caption || 'Image'}" />
                <div class="image-overlay">
                    ${img.is_primary ? '<span class="primary-badge">Primary</span>' : ''}
                    <div class="image-actions">
                        ${!img.is_primary && this.options.allowReorder ? 
                            `<button class="btn-set-primary" data-id="${img.id}">Set Primary</button>` : ''}
                        ${this.options.allowDelete ? 
                            `<button class="btn-delete" data-id="${img.id}">Delete</button>` : ''}
                    </div>
                </div>
                ${img.caption ? `<p class="image-caption">${img.caption}</p>` : ''}
            </div>
        `).join('');
    }

    renderUploadSection() {
        if (this.images.length >= this.options.maxImages) {
            return '<p class="max-images-reached">Maximum number of images reached</p>';
        }

        return `
            <div class="upload-section">
                <h3>Add Images</h3>
                <div class="upload-tabs">
                    <button class="tab-btn active" data-tab="file">Upload File</button>
                    <button class="tab-btn" data-tab="url">Add URL</button>
                </div>
                <div class="upload-content">
                    <div class="tab-content active" id="upload-file">
                        <form id="upload-form">
                            <input type="file" id="image-file" accept="image/*" required />
                            <input type="text" id="image-caption" placeholder="Caption (optional)" />
                            <label>
                                <input type="checkbox" id="is-primary" />
                                Set as primary image
                            </label>
                            <button type="submit" class="btn-upload">Upload</button>
                        </form>
                    </div>
                    <div class="tab-content" id="upload-url">
                        <form id="url-form">
                            <input type="url" id="image-url" placeholder="Image URL" required />
                            <input type="text" id="url-caption" placeholder="Caption (optional)" />
                            <label>
                                <input type="checkbox" id="url-is-primary" />
                                Set as primary image
                            </label>
                            <button type="submit" class="btn-add-url">Add Image</button>
                        </form>
                    </div>
                </div>
            </div>
        `;
    }

    attachEventListeners() {
        // Tab switching
        const tabBtns = this.container.querySelectorAll('.tab-btn');
        tabBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tab = e.target.dataset.tab;
                this.switchTab(tab);
            });
        });

        // File upload
        const uploadForm = this.container.querySelector('#upload-form');
        if (uploadForm) {
            uploadForm.addEventListener('submit', (e) => this.handleFileUpload(e));
        }

        // URL upload
        const urlForm = this.container.querySelector('#url-form');
        if (urlForm) {
            urlForm.addEventListener('submit', (e) => this.handleURLUpload(e));
        }

        // Set primary
        const primaryBtns = this.container.querySelectorAll('.btn-set-primary');
        primaryBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const imageId = parseInt(e.target.dataset.id);
                this.setPrimary(imageId);
            });
        });

        // Delete
        const deleteBtns = this.container.querySelectorAll('.btn-delete');
        deleteBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const imageId = parseInt(e.target.dataset.id);
                this.deleteImage(imageId);
            });
        });
    }

    switchTab(tab) {
        const tabBtns = this.container.querySelectorAll('.tab-btn');
        const tabContents = this.container.querySelectorAll('.tab-content');
        
        tabBtns.forEach(btn => btn.classList.remove('active'));
        tabContents.forEach(content => content.classList.remove('active'));
        
        this.container.querySelector(`[data-tab="${tab}"]`).classList.add('active');
        this.container.querySelector(`#upload-${tab}`).classList.add('active');
    }

    async handleFileUpload(e) {
        e.preventDefault();
        
        const fileInput = this.container.querySelector('#image-file');
        const captionInput = this.container.querySelector('#image-caption');
        const isPrimaryInput = this.container.querySelector('#is-primary');
        
        const file = fileInput.files[0];
        if (!file) return;

        const formData = new FormData();
        formData.append('image', file);
        formData.append('entity_type', this.entityType);
        formData.append('entity_id', this.entityId);
        formData.append('caption', captionInput.value);
        formData.append('is_primary', isPrimaryInput.checked);

        try {
            const token = localStorage.getItem('token');
            const response = await fetch('/images/upload', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData
            });

            if (response.ok) {
                alert('Image uploaded successfully!');
                fileInput.value = '';
                captionInput.value = '';
                isPrimaryInput.checked = false;
                await this.loadImages();
                this.render();
            } else {
                const error = await response.json();
                alert('Upload failed: ' + (error.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Upload error:', error);
            alert('Upload failed: ' + error.message);
        }
    }

    async handleURLUpload(e) {
        e.preventDefault();
        
        const urlInput = this.container.querySelector('#image-url');
        const captionInput = this.container.querySelector('#url-caption');
        const isPrimaryInput = this.container.querySelector('#url-is-primary');

        try {
            const token = localStorage.getItem('token');
            const response = await fetch('/images/add-url', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    entity_type: this.entityType,
                    entity_id: this.entityId,
                    image_url: urlInput.value,
                    caption: captionInput.value,
                    is_primary: isPrimaryInput.checked
                })
            });

            if (response.ok) {
                alert('Image added successfully!');
                urlInput.value = '';
                captionInput.value = '';
                isPrimaryInput.checked = false;
                await this.loadImages();
                this.render();
            } else {
                const error = await response.json();
                alert('Failed to add image: ' + (error.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Add URL error:', error);
            alert('Failed to add image: ' + error.message);
        }
    }

    async setPrimary(imageId) {
        try {
            const token = localStorage.getItem('token');
            const image = this.images.find(img => img.id === imageId);
            
            const response = await fetch('/images/update', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    id: imageId,
                    caption: image.caption,
                    display_order: image.display_order,
                    is_primary: true
                })
            });

            if (response.ok) {
                await this.loadImages();
                this.render();
            } else {
                const error = await response.json();
                alert('Failed to set primary: ' + (error.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Set primary error:', error);
            alert('Failed to set primary: ' + error.message);
        }
    }

    async deleteImage(imageId) {
        if (!confirm('Are you sure you want to delete this image?')) {
            return;
        }

        try {
            const token = localStorage.getItem('token');
            const response = await fetch('/images/delete', {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ id: imageId })
            });

            if (response.ok) {
                await this.loadImages();
                this.render();
            } else {
                const error = await response.json();
                alert('Failed to delete image: ' + (error.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Delete error:', error);
            alert('Failed to delete image: ' + error.message);
        }
    }

    // Public method to get primary image
    getPrimaryImage() {
        return this.images.find(img => img.is_primary);
    }

    // Public method to get all images
    getAllImages() {
        return this.images;
    }
}

// Simple image gallery for public viewing (no management features)
class ImageGallery {
    constructor(entityType, entityId, containerId) {
        this.entityType = entityType;
        this.entityId = entityId;
        this.container = document.getElementById(containerId);
        this.images = [];
        this.currentIndex = 0;
        this.init();
    }

    async init() {
        if (!this.container) return;
        await this.loadImages();
        this.render();
    }

    async loadImages() {
        try {
            const response = await fetch(`/images?entity_type=${this.entityType}&entity_id=${this.entityId}`);
            if (response.ok) {
                this.images = await response.json();
            }
        } catch (error) {
            console.error('Error loading images:', error);
        }
    }

    render() {
        if (!this.images || this.images.length === 0) {
            this.container.innerHTML = '<p class="no-images">No images available</p>';
            return;
        }

        const primaryImage = this.images.find(img => img.is_primary) || this.images[0];
        
        this.container.innerHTML = `
            <div class="image-gallery-viewer">
                <div class="main-image">
                    <img src="${primaryImage.image_url}" alt="${primaryImage.caption || 'Image'}" id="main-img" />
                    ${primaryImage.caption ? `<p class="caption">${primaryImage.caption}</p>` : ''}
                </div>
                ${this.images.length > 1 ? `
                    <div class="thumbnail-strip">
                        ${this.images.map((img, index) => `
                            <img src="${img.image_url}" 
                                 alt="${img.caption || 'Thumbnail'}" 
                                 class="thumbnail ${index === 0 ? 'active' : ''}"
                                 data-index="${index}" />
                        `).join('')}
                    </div>
                ` : ''}
            </div>
        `;

        this.attachGalleryListeners();
    }

    attachGalleryListeners() {
        const thumbnails = this.container.querySelectorAll('.thumbnail');
        thumbnails.forEach(thumb => {
            thumb.addEventListener('click', (e) => {
                const index = parseInt(e.target.dataset.index);
                this.showImage(index);
            });
        });
    }

    showImage(index) {
        this.currentIndex = index;
        const image = this.images[index];
        const mainImg = this.container.querySelector('#main-img');
        const caption = this.container.querySelector('.caption');
        
        mainImg.src = image.image_url;
        mainImg.alt = image.caption || 'Image';
        
        if (caption) {
            caption.textContent = image.caption || '';
        }

        // Update active thumbnail
        const thumbnails = this.container.querySelectorAll('.thumbnail');
        thumbnails.forEach((thumb, i) => {
            thumb.classList.toggle('active', i === index);
        });
    }
}
