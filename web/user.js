const base = 'http://localhost:8080';

function loadTodos() {
  fetch(base + '/todos')
    .then(res => res.json())
    .then(todos => {
      const list = document.getElementById('todos-list');
      list.innerHTML = '';
      if (Array.isArray(todos) && todos.length > 0) {
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
    .catch(err => {
      console.error('Error loading todos:', err);
      document.getElementById('todos-list').innerHTML = '<p>Error loading todos</p>';
    });
}

function loadUsers() {
  fetch(base + '/users')
    .then(res => res.json())
    .then(users => {
      const list = document.getElementById('users-list');
      list.innerHTML = '';
      if (Array.isArray(users) && users.length > 0) {
        users.forEach(user => {
          const item = document.createElement('div');
          item.className = 'item';
          item.innerHTML = `
            <span><strong>${user.name}</strong> - ${user.email}</span>
          `;
          list.appendChild(item);
        });
      } else {
        list.innerHTML = '<p>No users registered yet.</p>';
      }
    })
    .catch(err => {
      console.error('Error loading users:', err);
      document.getElementById('users-list').innerHTML = '<p>Error loading users</p>';
    });
}

function toggleTodo(id, completed) {
  fetch(base + '/todos', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, title: '', completed })
  })
    .then(res => {
      if (res.ok) loadTodos();
    })
    .catch(err => console.error('Error toggling todo:', err));
}

function deleteTodo(id) {
  fetch(base + '/todos', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id })
  })
    .then(res => {
      if (res.ok) loadTodos();
    })
    .catch(err => console.error('Error deleting todo:', err));
}

document.addEventListener('DOMContentLoaded', () => {
  setTimeout(() => {
    loadTodos();
    loadUsers();
  }, 500);

  // Add Todo
  const addTodoBtn = document.getElementById('btn-add-todo');
  if (addTodoBtn) {
    addTodoBtn.addEventListener('click', () => {
      const input = document.getElementById('todo-input');
      const title = input.value.trim();
      if (!title) {
        alert('Please enter a todo');
        return;
      }
      
      fetch(base + '/todos', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title })
      })
        .then(res => res.json())
        .then(data => {
          input.value = '';
          loadTodos();
        })
        .catch(err => {
          console.error('Error creating todo:', err);
          alert('Error creating todo: ' + err);
        });
    });
  }

  // Add User
  const addUserBtn = document.getElementById('btn-add-user');
  if (addUserBtn) {
    addUserBtn.addEventListener('click', () => {
      const name = document.getElementById('user-name').value.trim();
      const email = document.getElementById('user-email').value.trim();
      if (!name || !email) {
        alert('Please enter name and email');
        return;
      }
      
      fetch(base + '/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, email })
      })
        .then(res => res.json())
        .then(data => {
          document.getElementById('user-name').value = '';
          document.getElementById('user-email').value = '';
          loadUsers();
        })
        .catch(err => {
          console.error('Error creating user:', err);
          alert('Error creating user: ' + err);
        });
    });
  }
});
