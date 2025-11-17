const base = 'http://localhost:8080';

function out(elementId, text) {
  const el = document.getElementById(elementId);
  if (el) el.textContent = text;
}

function formatJson(obj) {
  return JSON.stringify(obj, null, 2);
}

document.addEventListener('DOMContentLoaded', () => {
  // Health Check
  document.getElementById('btn-health').addEventListener('click', async () => {
    out('output-basic', 'calling /health...');
    try {
      const res = await fetch(base + '/health');
      const txt = await res.text();
      out('output-basic', `status: ${res.status}\n\n${txt}`);
    } catch (err) {
      out('output-basic', 'error: ' + err);
    }
  });

  // Hello Endpoint
  document.getElementById('btn-hello').addEventListener('click', async () => {
    out('output-basic', 'calling /hello...');
    try {
      const res = await fetch(base + '/hello');
      const json = await res.json();
      out('output-basic', `status: ${res.status}\n\n` + formatJson(json));
    } catch (err) {
      out('output-basic', 'error: ' + err);
    }
  });

  // Time Endpoint
  document.getElementById('btn-time').addEventListener('click', async () => {
    out('output-time', 'calling /time...');
    try {
      const res = await fetch(base + '/time');
      const json = await res.json();
      out('output-time', `status: ${res.status}\n\n` + formatJson(json));
    } catch (err) {
      out('output-time', 'error: ' + err);
    }
  });

  // Echo Endpoint
  document.getElementById('btn-echo').addEventListener('click', async () => {
    const input = document.getElementById('echo-input').value;
    out('output-echo', 'parsing and sending...');
    try {
      const data = JSON.parse(input);
      const res = await fetch(base + '/echo', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      });
      const json = await res.json();
      out('output-echo', `status: ${res.status}\n\n` + formatJson(json));
    } catch (err) {
      out('output-echo', 'error: ' + err);
    }
  });

  // Stats Endpoint
  document.getElementById('btn-stats').addEventListener('click', async () => {
    out('output-stats', 'calling /stats...');
    try {
      const res = await fetch(base + '/stats');
      const json = await res.json();
      
      // Update stats grid
      document.getElementById('stat-requests').textContent = json.total_requests;
      document.getElementById('stat-uptime').textContent = json.uptime_seconds.toFixed(2);
      document.getElementById('stat-start').textContent = json.start_time;
      document.getElementById('stats-grid').style.display = 'grid';
      
      out('output-stats', `status: ${res.status}\n\n` + formatJson(json));
    } catch (err) {
      out('output-stats', 'error: ' + err);
    }
  });
});
