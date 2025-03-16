# Expense Manager

A modern web application for managing personal expenses, built with Go and HTMX.

## Features

- 💰 Track expenses with categories and descriptions
- 📊 View monthly summaries and statistics
- 📅 Navigate through expenses by month
- 📱 Responsive design with modern UI
- 🔄 Real-time updates using HTMX
- 📈 Visual reports and analytics
- 🛠️ Admin panel for data management

## Tech Stack

- Backend: Go
- Frontend: HTMX, TailwindCSS
- Database: PostgreSQL
- Development: Air (Live Reload)

## Getting Started

### Prerequisites

- Docker and Docker Compose
- VS Code with Remote - Containers extension

### Development with Dev Container (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/nunomcpereira/expensemanager.git
   cd expensemanager
   ```

2. Open in VS Code:
   ```bash
   code .
   ```

3. When prompted "Reopen in Container" click "Reopen in Container", or:
   - Press F1
   - Type "Dev Containers: Reopen in Container"
   - Press Enter

4. The dev container will:
   - Set up a complete Go development environment
   - Install all necessary tools (Air, golangci-lint)
   - Start PostgreSQL database
   - Configure all environment variables

5. Run the application:
   ```bash
   air
   ```

6. Open http://localhost:8080 in your browser

### Manual Installation (Alternative)

If you prefer not to use the dev container, you'll need:
- Go 1.21 or higher
- PostgreSQL
- Air (for live reload)

1. Clone the repository:
   ```bash
   git clone https://github.com/nunomcpereira/expensemanager.git
   cd expensemanager
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=expensemanager
   export DB_SSLMODE=disable
   ```

4. Run the application:
   ```bash
   air
   ```

## Project Structure

```
.
├── .devcontainer/     # Dev container configuration
├── cmd/
│   ├── migrate/      # Database migration tool
│   └── server/       # Main application
│       ├── main.go
│       ├── static/   # Static assets
│       └── templates/ # HTML templates
├── internal/
│   ├── config/      # Configuration
│   ├── database/    # Database operations
│   ├── handlers/    # HTTP handlers
│   ├── i18n/       # Internationalization
│   ├── middleware/  # HTTP middleware
│   └── models/     # Data models
└── db/             # Database files
```

## Database Migration

When switching from SQLite to PostgreSQL, use the migration tool:

```bash
go run cmd/migrate/main.go
```

The tool will:
- Create the necessary tables in PostgreSQL
- Copy all existing data from SQLite
- Maintain all relationships and data integrity

## Deployment

### Deploy to Render.com (Free Tier)

1. Fork this repository to your GitHub account

2. Create a Render.com account at https://render.com

3. In Render Dashboard:
   - Click "New +"
   - Select "Blueprint"
   - Connect your GitHub account if not already connected
   - Select your forked repository
   - Click "Connect"

4. Render will automatically:
   - Create a PostgreSQL database
   - Deploy the web service
   - Configure environment variables
   - Set up HTTPS

5. Once deployment is complete, click the generated URL to access your application

### Manual Deployment

If you prefer to deploy manually or to another platform:

1. Build the Docker image:
   ```bash
   docker build -t expensemanager .
   ```

2. Set up environment variables:
   ```bash
   export DB_HOST=your-db-host
   export DB_PORT=5432
   export DB_USER=your-db-user
   export DB_PASSWORD=your-db-password
   export DB_NAME=expensemanager
   export DB_SSLMODE=require
   ```

3. Run with docker-compose:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

## License

MIT License 