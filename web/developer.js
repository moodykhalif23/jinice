const base = 'http://localhost:8080';

function formatJson(obj) {
  return JSON.stringify(obj, null, 2);
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
