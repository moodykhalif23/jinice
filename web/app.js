const base = 'http://localhost:8080';

function out(elementId, text) {
  const el = document.getElementById(elementId);
  if (el) el.textContent = text;
}

function formatJson(obj) {
  return JSON.stringify(obj, null, 2);
}

function loadTodos() {
  fetch(base + '/todos')
    .then(res => res.json())
    .then(todos => {
      const list = document.getElementById('todos-list');
      list.innerHTML = '';
      if (todos && todos.length > 0) {
        todos.forEach(todo => {
          const item = document.createElement('div');
          item.className = 'item';
          item.innerHTML = `
            <span>${todo.id}. ${todo.title} ${todo.completed ? '✓' : ''}</span>
            <div>
              <button onclick="toggleTodo(${todo.id}, ${!todo.completed})" class="success">Toggle</button>
              <button onclick="deleteTodo(${todo.id})" class="danger">Delete</button>
            </div>
          `;
          list.appendChild(item);
        });
      } else {
        list.innerHTML = '<p>No todos yet. Add one above!</p>';
      }
    })
    .catch(err => console.error(err));
}

function loadUsers() {
  fetch(base + '/users')
    .then(res => res.json())
    .then(users => {
      const list = document.getElementById('users-list');
      list.innerHTML = '';
      if (users && users.length > 0) {
        users.forEach(user => {
          const item = document.createElement('div');
          item.className = 'item';
          item.innerHTML = `
            <span><strong>${user.name}</strong> - ${user.email}</span>
          `;
          list.appendChild(item);
        });
      } else {
        list.innerHTML = '<p>No users yet. Add one above!</p>';
      }
    })
    .catch(err => console.error(err));
}

function toggleTodo(id, completed) {
  fetch(base + '/todos', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, title: '', completed })
  })
    .then(() => loadTodos())
    .catch(err => console.error(err));
}

function deleteTodo(id) {
  fetch(base + '/todos', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id })
  })
    .then(() => loadTodos())
    .catch(err => console.error(err));
}

document.addEventListener('DOMContentLoaded', () => {
  loadTodos();
  loadUsers();

  // Add Todo
  document.getElementById('btn-add-todo').addEventListener('click', () => {
    const input = document.getElementById('todo-input');
    const title = input.value.trim();
    if (!title) return;
    
    fetch(base + '/todos', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title })
    })
      .then(() => {
        input.value = '';
        loadTodos();
      })
      .catch(err => console.error(err));
  });

  // Add User
  document.getElementById('btn-add-user').addEventListener('click', () => {
    const name = document.getElementById('user-name').value.trim();
    const email = document.getElementById('user-email').value.trim();
    if (!name || !email) return;
    
    fetch(base + '/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, email })
    })
      .then(() => {
        document.getElementById('user-name').value = '';
        document.getElementById('user-email').value = '';
        loadUsers();
      })
      .catch(err => console.error(err));
  });

  // Health Check
  document.getElementById('btn-health').addEventListener('click', async () => {
    out('output-dev', 'calling /health...');
    try {
      const res = await fetch(base + '/health');
      const txt = await res.text();
      out('output-dev', `status: ${res.status}\n\n${txt}`);
    } catch (err) {
      out('output-dev', 'error: ' + err);
    }
  });

  // Hello Endpoint
  document.getElementById('btn-hello').addEventListener('click', async () => {
    out('output-dev', 'calling /hello...');
    try {
      const res = await fetch(base + '/hello');
      const json = await res.json();
      out('output-dev', `status: ${res.status}\n\n` + formatJson(json));
    } catch (err) {
      out('output-dev', 'error: ' + err);
    }
  });

  // Time Endpoint
  document.getElementById('btn-time').addEventListener('click', async () => {
    out('output-dev', 'calling /time...');
    try {
      const res = await fetch(base + '/time');
      const json = await res.json();
      out('output-dev', `status: ${res.status}\n\n` + formatJson(json));
    } catch (err) {
      out('output-dev', 'error: ' + err);
    }
  });

  // Stats Endpoint
  document.getElementById('btn-stats').addEventListener('click', async () => {
    out('output-dev', 'calling /stats...');
    try {
      const res = await fetch(base + '/stats');
      const json = await res.json();
      out('output-dev', `status: ${res.status}\n\n` + formatJson(json));
    } catch (err) {
      out('output-dev', 'error: ' + err);
    }
  });
});
