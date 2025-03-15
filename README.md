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
- Database: SQLite
- Development: Air (Live Reload)

## Getting Started

### Prerequisites

- Go 1.21 or higher
- SQLite

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/nunomcpereira/expensemanager.git
   cd expensemanager
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the application:
   ```bash
   ./bin/air
   ```

4. Open http://localhost:8080 in your browser

## Development

The application uses Air for live reloading during development. Any changes to Go files, templates, or static assets will trigger an automatic rebuild and reload.

## Project Structure

```
.
├── cmd/server/          # Main application entry
│   ├── main.go
│   ├── static/         # Static assets
│   └── templates/      # HTML templates
├── internal/
│   ├── database/      # Database operations
│   ├── handlers/      # HTTP handlers
│   ├── middleware/    # HTTP middleware
│   └── models/        # Data models
└── db/                # Database files
```

## License

MIT License 