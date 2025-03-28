# Go Todo API
A simple RESTful API for managing todos with user authentication, built with Go.

## Overview
This is a Todo API that allows users to:
- Register and login with JWT authentication
- Create, read, update, and delete todos
- Filter todos by due date and tags
- Paginate todo lists

The API uses PostgreSQL as the database and follows a clean architecture pattern.

## Features

### Authentication
- User registration with password hashing
- JWT-based login with access and refresh tokens
- Token refresh endpoint
- Refresh token management in database
- Authentication middleware

### Todo Management
- Create new todos
- List todos with pagination and filtering
- Get todo by ID
- Update existing todos 
- Delete todos

## API Endpoints
Base path: `bash /api/v2`

### Auth Routes
- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login and get tokens
- `POST /auth/refresh` - Refresh access token

### Todo Routes (Protected)
- `POST /todos` - Create a new todo
- `GET /todos` - List todos with pagination and filters
- `GET /todos/{id}` - Get a specific todo 
- `PUT /todos` - Update a todo 
- `DELETE /todos/{id}` - Delete a todo

For detailed API documentation:
- See [swagger.yaml](docs/swagger.yaml) for the static Swagger specification.
- Access the interactive Swagger UI at `http://localhost:8888/swagger/index.html` when the server is running (e.g., in local environment).


## Running the Application
Build and run:
```bash
make build
./cmd/app
```
Or run directly:
```bash
go run cmd/main.go
```