# Chirpy Web Server Documentation

## Overview

The **Chirpy Web Server** is a Go-based API server that provides functionality for managing users and chirps (posts). It includes endpoints for user creation, chirp management, authentication, and webhook handling for upgrading users.

The server serves static assets, tracks file server hits, and provides metrics for administrative purposes.

## Environment Variables

The server requires several environment variables for proper configuration:

| Variable   | Description                                     |
|------------|-------------------------------------------------|
| `DB_URL`   | The database URL connection string (PostgreSQL) |
| `PLATFORM` | The platform the server is running on           |
| `SECRET`   | Secret used for token signing (JWT)             |
| `POLKA_KEY`| API Key required for Polka webhook validation   |

Ensure all the environment variables are set before running the server.

## Endpoints

### Health Check

- **GET `/api/healthz`**
  - **Description**: Check the health status of the server.
  - **Response**: HTTP 200 if the server is healthy.

### Metrics

- **GET `/admin/metrics`**
  - **Description**: Retrieve server metrics, including file server hit counts.

### Reset Metrics

- **POST `/admin/reset`**
  - **Description**: Delete all the data from the DB (can be used only if PLATFORM = "dev")

### Users

- **POST `/api/users`**
  - **Description**: Create a new user.
  - **Request Body**:
    ```json
    {
      "email": "example@example.com",
      "password": "password123"
    }
    ```
  - **Response**: Returns the newly created user's information.

- **PUT `/api/users`**
  - **Description**: Update user's password and email.
  - **Request Body**:
    ```json
    {
      "email": "newemail@example.com",
      "password": "newpassword123"
    }
    ```

### Authentication

- **POST `/api/login`**
  - **Description**: Log in a user and retrieve an access token.
  - **Request Body**:
    ```json
    {
      "email": "example@example.com",
      "password": "password123"
    }
    ```
  - **Response**: A JSON Web Token (JWT) for authenticated access.

- **POST `/api/refresh`**
  - **Description**: Refresh the access token using a refresh token.
  - **Headers**: `Authorization: Bearer <refresh_token>`
  - **Response**: A new access token.

- **POST `/api/revoke`**
  - **Description**: Revoke a refresh token, preventing further use.

### Chirps

- **POST `/api/chirps`**
  - **Description**: Create a new chirp.
  - **Request Body**:
    ```json
    {
      "body": "This is a chirp."
    }
    ```
  - **Response**: The created chirp's information.

- **GET `/api/chirps`**
  - **Description**: Retrieve all chirps. Supports filtering by author and sorting.
  - **Query Parameters**:
    - `author_id`: Filter chirps by a specific user.
    - `sort`: Sort chirps by creation date (`asc` or `desc`).

- **GET `/api/chirps/{id}`**
  - **Description**: Retrieve a specific chirp by its ID.

- **DELETE `/api/chirps/{chirpID}`**
  - **Description**: Delete a chirp by its ID.

### Webhooks

- **POST `/api/polka/webhooks`**
  - **Description**: Handle Polka webhook events, such as user upgrades.
  - **Request Body**:
    ```json
    {
      "event": "user.upgraded",
      "data": {
        "user_id": "3311741c-680c-4546-99f3-fc9efac2036c"
      }
    }
    ```

## Running the Server

Ensure that all required environment variables are set:

```bash
export DB_URL="your-database-url"
export PLATFORM="your-platform"
export SECRET="your-secret"
export POLKA_KEY="your-polka-key"
```

## Error handling
For failed requests, the server responds with appropriate HTTP status codes such as:
- **400 Bad Request: `Invalid request data.`**
- **401 Unauthorized: `Authentication failure or missing/invalid token..`**
- **403 Forbidden: `Access to the resource is denied.`**
- **400 Not Found: `Resource not found.`**