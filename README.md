# User Management Microservice

This microservice is built with Go and designed to handle user management functionalities.

## Prerequisites

Before running the microservice, make sure you have the following prerequisites installed:

- Go (version 1.20.4)


## Getting Started

Follow these steps to get the microservice up and running:

1. Clone the repository:

```shell
git clone https://github.com/FranRossi/ASP-MicroService-GO.git
```

2. Install dependencies:

```shell
go mod download
```

3. Configure environment variables:

```shell
cp .env.example .env
```

4. Start container:
    
```shell
docker compose up
```

The microservice should now be running on http://localhost:6000


## API Documentation

The following endpoints are available:

| Method | Endpoint | Description |
| --- | --- | --- |
| GET | /users | Get all users or filter users by email or company |
| GET | /users/:id | Get user by id |
| POST | /users | Create new user |

### GET /users

Returns a list of all users. You can optionally filter the users based on the following query parameters:

- `email`: Search for a specific user by email.
- `company`: Get all users associated with a particular company.

Examples:

- To get a user by email: `GET /users?email=johndoe@example.com`
- To get all users in a specific company: `GET /users?company=648e278ab68985665b4fc6e8`

Note: If both `email` and `company` query parameters are provided, the microservice will prioritize the `email` parameter.


## Testing

To run the tests, run the following command:

```shell
go test -cover ./cmd/controllers/...
```