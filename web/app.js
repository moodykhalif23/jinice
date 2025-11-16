const base = 'http://localhost:8080';

function out(text) {
  document.getElementById('output').textContent = text;
}

document.addEventListener('DOMContentLoaded', () => {
  document.getElementById('btn-health').addEventListener('click', async () => {
    out('calling /health...');
    try {
      const res = await fetch(base + '/health');
      const txt = await res.text();
      out(`status: ${res.status}\n\n${txt}`);
    } catch (err) {
      out('error: ' + err);
    }
  });

  document.getElementById('btn-hello').addEventListener('click', async () => {
    out('calling /hello...');
    try {
      const res = await fetch(base + '/hello');
      const json = await res.json();
      out(`status: ${res.status}\n\n` + JSON.stringify(json, null, 2));
    } catch (err) {
      out('error: ' + err);
    }
  });
});
