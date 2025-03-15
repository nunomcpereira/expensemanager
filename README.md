# Expense Manager

A simple and attractive expense manager application built with Go, HTMX, and SQLite.

## Features

- Add expenses with amount, description, category, and date
- View all expenses in a responsive table
- Delete expenses
- Real-time updates using HTMX
- Attractive UI with Tailwind CSS
- Local SQLite database storage

## Prerequisites

- Go 1.16 or later
- SQLite3

## Installation

1. Clone the repository
2. Navigate to the project directory
3. Install dependencies:
   ```bash
   go mod download
   ```

## Running the Application

1. Start the server:
   ```bash
   go run main.go
   ```
2. Open your browser and navigate to `http://localhost:8080`

## Usage

- To add an expense, fill out the form at the top of the page and click "Add Expense"
- To delete an expense, click the "Delete" button next to the expense in the table
- The table updates automatically without page refreshes thanks to HTMX

## Technology Stack

- Backend: Go
- Frontend: HTMX + Tailwind CSS
- Database: SQLite
- Additional: Hyperscript for enhanced interactivity 