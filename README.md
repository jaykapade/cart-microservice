# Cart Microservice

A microservice-based shopping cart system built with GraphQL and gRPC.

## Architecture

The system consists of multiple microservices:

- Account Service
- Catalog Service
- Order Service
- GraphQL API Gateway

## Technologies

- GraphQL (gqlgen)
- gRPC
- Protocol Buffers
- Go
- Docker

## Features

- Product catalog management
- Account management
- Order processing
- Pagination support
- GraphQL API for frontend integration

## API Endpoints

- The GraphQL playground is available at: `localhost:8000/playground`
- The GraphQL API is available at: `localhost:8000/graphql`

### GraphQL Queries

- `products` - Get products with pagination, search and filtering
- `accounts` - Get user accounts with pagination and filtering
- `orders` - Get order history for accounts

### GraphQL Mutations

- `createAccount` - Create a new user account
- `createProduct` - Create a new product
- `createOrder` - Create a new order

### GraphQL Types

- `Account` - User account information
- `Product` - Product details
- `Order` - Order information with ordered products
- `OrderedProduct` - Products within an order

## Getting Started

1. Run `docker compose up -d --build` to build and start all services
2. Access the GraphQL playground at `localhost:8000/playground`

## Development

The project uses:

- Generated GraphQL code via gqlgen
- Protocol buffer definitions for service communication
- Middleware support for error handling and recovery
- Type-safe resolvers

## Schema

The GraphQL schema defines the complete API surface including:

- Queries
- Types
- Input types
- Pagination
- Search/filtering

## Contributing

1. Fork the repository
2. Create a feature branch
3. Submit a pull request
