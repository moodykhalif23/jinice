const base = 'http://localhost:8080';
let eventRefreshInterval;
let autoRefreshEnabled = true;
let lastEventCount = 0;

function formatJson(obj) {
  return JSON.stringify(obj, null, 2);
}

function loadEvents() {
  fetch(base + '/events')
    .then(res => res.json())
    .then(events => {
      const feed = document.getElementById('events-feed');
      const status = document.getElementById('feed-status');
      feed.innerHTML = '';
      
      if (!Array.isArray(events) || events.length === 0) {
        feed.innerHTML = '<p>No events yet. Create todos or users to see activity.</p>';
        status.textContent = 'No events';
        lastEventCount = 0;
        return;
      }
      
      lastEventCount = events.length;
      status.textContent = `${events.length} event${events.length !== 1 ? 's' : ''} • Last updated: ${new Date().toLocaleTimeString()}`;
      
      events.reverse().forEach(event => {
        const item = document.createElement('div');
        item.className = `event-item ${event.type}`;
        const timestamp = new Date(event.timestamp);
        const timeStr = timestamp.toLocaleTimeString();
        const dateStr = timestamp.toLocaleDateString();
        
        const typeLabel = event.type
          .replace(/_/g, ' ')
          .replace(/\b\w/g, l => l.toUpperCase());
        
        item.innerHTML = `
          <div class="event-type">${typeLabel}</div>
          <div><strong>${event.message}</strong></div>
          <div class="event-time">${dateStr} ${timeStr}</div>
        `;
        feed.appendChild(item);
      });
    })
    .catch(err => {
      console.error('Error loading events:', err);
      document.getElementById('events-feed').innerHTML = '<p>Error loading events</p>';
      document.getElementById('feed-status').textContent = 'Error loading';
    });
}

function clearEvents() {
  document.getElementById('events-feed').innerHTML = '<p>Feed cleared. Create new todos or users to see events.</p>';
  document.getElementById('feed-status').textContent = 'Feed cleared';
}

function toggleAutoRefresh() {
  autoRefreshEnabled = !autoRefreshEnabled;
  const btn = document.getElementById('auto-refresh-btn');
  
  if (autoRefreshEnabled) {
    btn.textContent = 'Auto-Refresh: ON';
    btn.style.backgroundColor = '#28a745';
    eventRefreshInterval = setInterval(loadEvents, 2000);
  } else {
    btn.textContent = 'Auto-Refresh: OFF';
    btn.style.backgroundColor = '#dc3545';
    if (eventRefreshInterval) {
      clearInterval(eventRefreshInterval);
    }
  }
}

function callApi(endpoint) {
  const outputId = 'output-' + endpoint;
  const output = document.getElementById(outputId);
  if (!output) return;
  
  output.textContent = `Calling ${endpoint}...`;
  output.classList.remove('output-hidden');

  let url = base + '/' + endpoint;

  switch(endpoint) {
    case 'health':
      fetch(url)
        .then(res => res.text())
        .then(txt => {
          output.textContent = `GET /${endpoint}\nStatus: 200\n\n${txt}`;
        })
        .catch(err => output.textContent = 'Error: ' + err);
      break;
    case 'hello':
      fetch(url)
        .then(res => res.json())
        .then(json => {
          output.textContent = `GET /${endpoint}\nStatus: 200\n\n${formatJson(json)}`;
        })
        .catch(err => output.textContent = 'Error: ' + err);
      break;
    case 'time':
      fetch(url)
        .then(res => res.json())
        .then(json => {
          output.textContent = `GET /${endpoint}\nStatus: 200\n\n${formatJson(json)}`;
        })
        .catch(err => output.textContent = 'Error: ' + err);
      break;
    case 'stats':
      fetch(url)
        .then(res => res.json())
        .then(json => {
          output.textContent = `GET /${endpoint}\nStatus: 200\n\n${formatJson(json)}`;
        })
        .catch(err => output.textContent = 'Error: ' + err);
      break;
    case 'todos':
      fetch(url)
        .then(res => res.json())
        .then(json => {
          output.textContent = `GET /${endpoint}\nStatus: 200\n\n${formatJson(json)}`;
        })
        .catch(err => output.textContent = 'Error: ' + err);
      break;
    case 'users':
      fetch(url)
        .then(res => res.json())
        .then(json => {
          output.textContent = `GET /${endpoint}\nStatus: 200\n\n${formatJson(json)}`;
        })
        .catch(err => output.textContent = 'Error: ' + err);
      break;
    default:
      output.textContent = 'Unknown endpoint';
  }
}

// Load events on page load and auto-refresh
document.addEventListener('DOMContentLoaded', () => {
  loadEvents();
  eventRefreshInterval = setInterval(loadEvents, 2000);
});

// Stop auto-refresh when leaving the page
window.addEventListener('beforeunload', () => {
  if (eventRefreshInterval) {
    clearInterval(eventRefreshInterval);
  }
});
