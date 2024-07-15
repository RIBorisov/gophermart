# gophermart

# Service Description
This service provides a RESTful API for managing user accounts, orders and accrual withdrawals. It uses the Chi router and includes several middleware functions for logging, error recovery, and authentication.

# Endpoints
### User Management
   - **POST** /api/user/register: Registers a new user.
   - **POST** /api/user/login: Logs in an existing user.
   
### Protected Endpoints (Require Authentication)
   - **POST** /api/user/orders: Creates a new order for the authenticated user.
   - **GET** /api/user/orders: Retrieves a list of orders for the authenticated user.
   - **GET** /api/user/balance: Retrieves the current balance of the authenticated user.
   - **POST** /api/user/balance/withdraw: Initiates a withdrawal from the authenticated user's balance.
   - **GET** /api/user/withdrawals: Retrieves a list of withdrawals for the authenticated user.
   
# Middleware
   - Logger: Logs requests and responses.
   - Recoverer: Recovers from panics and returns a 500 error.
   - CheckAuth: Checks if the user is authenticated before allowing access to protected endpoints.

# Local launch

1. Clone the repository to any suitable directory on your computer.
2. Run the command `go mod tidy` in the repository root to pull dependencies.
3. Start the database using docker-compose (download from the official Docker [website](https://www.docker.com/products/docker-desktop/)). 
   - Run the command `docker-compose up -d db` to start the database container.
   - Check if the database container is running with `docker-compose ps`.
   - Stop the database container with `docker-compose down`.
4. Launch the application in one of the following ways:
   - From the `/cmd/gophermart` directory, run `go run . -d <DATABASE_URI>`, where `DATABASE_URI` is the database connection string, for example, `postgresql://odmen:odmenpass@localhost:5432/gophermart?sslmode=disable`.
   - Using your IDE, where you need to set the `DATABASE_URI` environment variable beforehand, for example, `export DATABASE_URI=postgresql://odmen:odmenpass@localhost:5432/gophermart?sslmode=disable`.
5. Launch Accrual application from the directory `/cmd/accrual` on your device 
   - `./accrual_darwin_amd64` - for Intel Macbook OS
   - `./accrual_darwin_ard64` - for Silicon Macbook (M series chip) OS 
   - `./accrual_linux_amd64` - for Linux OS
   - `./accrual_darwin_amd64` - for Windows OS 