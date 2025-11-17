# JiNice

A full-stack business directory and events platform with user authentication, business management, event management, and analytics.

## Features

- **Secure Authentication**: JWT-based authentication with database-stored sessions
- **User Management**: Business owner registration and login
- **Business Management**: Create, read, update, and delete business listings
- **Event Management**: Business owners can create and manage events for their businesses
- **Public Event Browsing**: Users can discover upcoming events from local businesses
- **Analytics**: Track business views and statistics
- **Responsive UI**: Clean, modern interface for both public and business owner portals

## Tech Stack

- **Backend**: Go (Golang) with MySQL database
- **Frontend**: Vanilla JavaScript, HTML, CSS
- **Infrastructure**: Docker & Docker Compose
- **Authentication**: JWT tokens with bcrypt password hashing

## Getting Started

### Prerequisites

- Docker and Docker Compose installed
- Go 1.23+ (for local development)

### Running with Docker

1. **Start the application**:
   ```bash
   docker-compose up -d
   ```

### Local Development (without Docker)

1. **Start MySQL**:
   ```bash
   docker-compose up -d mysql
   ```

2. **Update .env** to use localhost:
   ```
   DB_HOST=localhost
   ```

3. **Run the application**:
   ```bash
   go run ./cmd/app
   ```

### Building

```bash
go build -o main.exe ./cmd/app
```

## License

Copyright © 2026 Jinice

## Contributing

Feel free to submit issues and enhancement requests!
