// Image Carousel functionality for cards
class CardCarousel {
  constructor(container, images) {
    this.container = container;
    this.images = images || [];
    this.currentIndex = 0;
    this.autoPlayInterval = null;
    this.autoPlayDelay = 4000; // 4 seconds
  }

  render() {
    if (!this.images || this.images.length === 0) {
      return '<div class="card-thumb" style="background-color: #e9ecef; display: flex; align-items: center; justify-content: center; color: #999;">No image</div>';
    }

    if (this.images.length === 1) {
      return `<img class="card-thumb" src="${this.images[0]}" alt="Image" loading="lazy">`;
    }

    const carouselId = 'carousel-' + Math.random().toString(36).substr(2, 9);
    
    const html = `
      <div class="card-image-carousel" data-carousel-id="${carouselId}">
        <div class="card-carousel-track">
          ${this.images.map(img => `<img class="card-carousel-image" src="${img}" alt="Image" loading="lazy">`).join('')}
        </div>
        <button class="card-carousel-btn prev" aria-label="Previous image">‹</button>
        <button class="card-carousel-btn next" aria-label="Next image">›</button>
        <div class="card-carousel-indicators">
          ${this.images.map((_, i) => `<div class="card-carousel-dot ${i === 0 ? 'active' : ''}"></div>`).join('')}
        </div>
      </div>
    `;

    return html;
  }

  init(carouselElement) {
    if (!carouselElement) return;

    const track = carouselElement.querySelector('.card-carousel-track');
    const prevBtn = carouselElement.querySelector('.card-carousel-btn.prev');
    const nextBtn = carouselElement.querySelector('.card-carousel-btn.next');
    const dots = carouselElement.querySelectorAll('.card-carousel-dot');

    if (!track || !prevBtn || !nextBtn) return;

    const updateCarousel = () => {
      track.style.transform = `translateX(-${this.currentIndex * 100}%)`;
      dots.forEach((dot, i) => {
        dot.classList.toggle('active', i === this.currentIndex);
      });
    };

    const goToNext = () => {
      this.currentIndex = (this.currentIndex + 1) % this.images.length;
      updateCarousel();
    };

    const goToPrev = () => {
      this.currentIndex = (this.currentIndex - 1 + this.images.length) % this.images.length;
      updateCarousel();
    };

    // Prevent card navigation when clicking carousel buttons
    prevBtn.addEventListener('click', (e) => {
      e.preventDefault();
      e.stopPropagation();
      goToPrev();
      this.resetAutoPlay();
    });

    nextBtn.addEventListener('click', (e) => {
      e.preventDefault();
      e.stopPropagation();
      goToNext();
      this.resetAutoPlay();
    });

    // Dot indicators
    dots.forEach((dot, i) => {
      dot.addEventListener('click', (e) => {
        e.preventDefault();
        e.stopPropagation();
        this.currentIndex = i;
        updateCarousel();
        this.resetAutoPlay();
      });
    });

    // Auto-play
    this.startAutoPlay(goToNext);

    // Pause on hover
    carouselElement.addEventListener('mouseenter', () => {
      this.stopAutoPlay();
    });

    carouselElement.addEventListener('mouseleave', () => {
      this.startAutoPlay(goToNext);
    });
  }

  startAutoPlay(callback) {
    this.stopAutoPlay();
    if (this.images.length > 1) {
      this.autoPlayInterval = setInterval(callback, this.autoPlayDelay);
    }
  }

  stopAutoPlay() {
    if (this.autoPlayInterval) {
      clearInterval(this.autoPlayInterval);
      this.autoPlayInterval = null;
    }
  }

  resetAutoPlay() {
    const callback = () => {
      this.currentIndex = (this.currentIndex + 1) % this.images.length;
      const carouselElement = document.querySelector(`[data-carousel-id]`);
      if (carouselElement) {
        const track = carouselElement.querySelector('.card-carousel-track');
        const dots = carouselElement.querySelectorAll('.card-carousel-dot');
        if (track) {
          track.style.transform = `translateX(-${this.currentIndex * 100}%)`;
          dots.forEach((dot, i) => {
            dot.classList.toggle('active', i === this.currentIndex);
          });
        }
      }
    };
    this.startAutoPlay(callback);
  }
}

// Initialize all carousels on the page
function initializeCarousels() {
  const carousels = document.querySelectorAll('.card-image-carousel');
  carousels.forEach(carousel => {
    const images = Array.from(carousel.querySelectorAll('.card-carousel-image')).map(img => img.src);
    const carouselInstance = new CardCarousel(carousel, images);
    carouselInstance.init(carousel);
  });
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { CardCarousel, initializeCarousels };
}
