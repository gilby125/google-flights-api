# Airfare Search Microservice Architecture Plan

## Overview

This document outlines the architecture and implementation plan for a scalable microservice system that leverages the Google Flights API to search for cheap airfare. The system will include an API service, worker system with queue for Google Flights queries, database integration with PostgreSQL and Neo4j, an admin interface for controlling scraping processes, and a web interface for ad-hoc multi-destination searches.

## System Components

### 1. Core Components

- **API Service**: RESTful API for flight searches and data retrieval
- **Worker System**: Distributed workers that process flight search jobs from a queue
- **Queue System**: Message broker for distributing search jobs to workers
- **Database Layer**: PostgreSQL for relational data and Neo4j for graph-based route analysis
- **Admin Interface**: Web UI for controlling and monitoring the scraping process
- **Search Interface**: Web UI for ad-hoc flight searches across multiple destinations

### 2. Architecture Diagram

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Web UI    │     │ Admin Panel │     │  API Client  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
        │                  │                   │
        └──────────┬──────┴───────────┬───────┘
                   │                  │
           ┌───────▼──────┐   ┌──────▼───────┐
           │  API Gateway │   │ Auth Service │
           └───────┬──────┘   └──────────────┘
                   │
        ┌──────────┴───────────┐
        │                      │
┌───────▼──────┐      ┌───────▼──────┐
│  Flight API   │      │ Admin API    │
└───────┬──────┘      └───────┬──────┘
        │                     │
┌───────▼──────┐      ┌──────▼───────┐
│ Queue Service │      │ Monitoring   │
└───────┬──────┘      └──────────────┘
        │
        │    ┌─────────────────┐
        ├────► Worker Node 1   │
        │    └─────────────────┘
        │    ┌─────────────────┐
        ├────► Worker Node 2   │
        │    └─────────────────┘
        │    ┌─────────────────┐
        └────► Worker Node N   │
             └─────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
┌───────▼──────┐         ┌───────▼──────┐
│  PostgreSQL   │         │    Neo4j     │
└───────────────┘         └──────────────┘
```

## Database Schema

### PostgreSQL Schema

```sql
-- Airports table
CREATE TABLE airports (
    code VARCHAR(3) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    country VARCHAR(255) NOT NULL,
    latitude DECIMAL(10, 6),
    longitude DECIMAL(10, 6)
);

-- Airlines table
CREATE TABLE airlines (
    code VARCHAR(3) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    country VARCHAR(255)
);

-- Search queries table
CREATE TABLE search_queries (
    id SERIAL PRIMARY KEY,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    adults INT DEFAULT 1,
    children INT DEFAULT 0,
    infants_lap INT DEFAULT 0,
    infants_seat INT DEFAULT 0,
    trip_type VARCHAR(20) NOT NULL,
    class VARCHAR(20) NOT NULL,
    stops VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending',
    FOREIGN KEY (origin) REFERENCES airports(code),
    FOREIGN KEY (destination) REFERENCES airports(code)
);

-- Flight offers table
CREATE TABLE flight_offers (
    id SERIAL PRIMARY KEY,
    search_query_id INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    total_duration INT NOT NULL, -- in minutes
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (search_query_id) REFERENCES search_queries(id)
);

-- Flight segments table
CREATE TABLE flight_segments (
    id SERIAL PRIMARY KEY,
    flight_offer_id INT NOT NULL,
    airline_code VARCHAR(3) NOT NULL,
    flight_number VARCHAR(10) NOT NULL,
    departure_airport VARCHAR(3) NOT NULL,
    arrival_airport VARCHAR(3) NOT NULL,
    departure_time TIMESTAMP NOT NULL,
    arrival_time TIMESTAMP NOT NULL,
    duration INT NOT NULL, -- in minutes
    airplane VARCHAR(100),
    legroom VARCHAR(50),
    is_return BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (flight_offer_id) REFERENCES flight_offers(id),
    FOREIGN KEY (airline_code) REFERENCES airlines(code),
    FOREIGN KEY (departure_airport) REFERENCES airports(code),
    FOREIGN KEY (arrival_airport) REFERENCES airports(code)
);

-- Scheduled jobs table
CREATE TABLE scheduled_jobs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    cron_expression VARCHAR(100) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    last_run TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Job details table
CREATE TABLE job_details (
    id SERIAL PRIMARY KEY,
    job_id INT NOT NULL,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date_start DATE NOT NULL,
    departure_date_end DATE NOT NULL,
    return_date_start DATE,
    return_date_end DATE,
    trip_length INT,
    adults INT DEFAULT 1,
    children INT DEFAULT 0,
    infants_lap INT DEFAULT 0,
    infants_seat INT DEFAULT 0,
    trip_type VARCHAR(20) NOT NULL,
    class VARCHAR(20) NOT NULL,
    stops VARCHAR(20) NOT NULL,
    FOREIGN KEY (job_id) REFERENCES scheduled_jobs(id),
    FOREIGN KEY (origin) REFERENCES airports(code),
    FOREIGN KEY (destination) REFERENCES airports(code)
);
```

### Neo4j Schema

Neo4j will be used to model flight routes as a graph for more complex route analysis:

```cypher
// Airport nodes
CREATE (a:Airport {code: 'JFK', name: 'John F. Kennedy International Airport', city: 'New York', country: 'USA'})

// Airline nodes
CREATE (a:Airline {code: 'DL', name: 'Delta Air Lines', country: 'USA'})

// Flight route relationships
MATCH (origin:Airport {code: 'JFK'}), (destination:Airport {code: 'LAX'})
CREATE (origin)-[:ROUTE {airline: 'DL', flightNumber: 'DL123', avgPrice: 350.00, avgDuration: 360}]->(destination)

// Price history relationships
MATCH (origin:Airport {code: 'JFK'}), (destination:Airport {code: 'LAX'})
CREATE (origin)-[:PRICE_POINT {date: '2023-06-01', price: 299.99, airline: 'DL'}]->(destination)
```

## API Endpoints

### Flight Search API

```
GET /api/v1/airports - List all airports
GET /api/v1/airlines - List all airlines

POST /api/v1/search - Create a new search
GET /api/v1/search/{id} - Get search results by ID
GET /api/v1/search - List recent searches

POST /api/v1/bulk-search - Create a bulk search across multiple destinations/dates
GET /api/v1/bulk-search/{id} - Get bulk search results

GET /api/v1/price-history/{origin}/{destination} - Get price history for a route
```

### Admin API

```
GET /api/v1/admin/jobs - List all scheduled jobs
POST /api/v1/admin/jobs - Create a new scheduled job
GET /api/v1/admin/jobs/{id} - Get job details
PUT /api/v1/admin/jobs/{id} - Update a job
DELETE /api/v1/admin/jobs/{id} - Delete a job

POST /api/v1/admin/jobs/{id}/run - Manually trigger a job
POST /api/v1/admin/jobs/{id}/enable - Enable a job
POST /api/v1/admin/jobs/{id}/disable - Disable a job

GET /api/v1/admin/workers - Get worker status
GET /api/v1/admin/queue - Get queue status
```

## Implementation Plan

### Phase 1: Core Infrastructure

1. Set up project structure and Docker environment
2. Implement database schemas (PostgreSQL and Neo4j)
3. Create basic API server with health endpoints
4. Implement message queue system (Redis or RabbitMQ)
5. Create worker framework

### Phase 2: Flight Search Implementation

1. Integrate Google Flights API client
2. Implement flight search endpoints
3. Create worker implementation for processing flight searches
4. Implement data storage and retrieval

### Phase 3: Admin and Monitoring

1. Create admin API endpoints
2. Implement scheduled job system
3. Build admin web interface
4. Add monitoring and logging

### Phase 4: Search Interface

1. Design and implement web UI for flight searches
2. Add multi-destination search capabilities
3. Implement date range visualization
4. Add price alerts and notifications

## Technology Stack

- **Backend**: Go (with Gin or Echo framework)
- **Queue**: Redis or RabbitMQ
- **Databases**: PostgreSQL and Neo4j
- **Frontend**: React with TypeScript
- **Containerization**: Docker and Docker Compose
- **API Gateway**: Traefik
- **Monitoring**: Prometheus and Grafana

## Deployment Architecture

The system will be deployed using Docker Compose for development and potentially Kubernetes for production. Each component (API, workers, databases, etc.) will be containerized separately to allow for independent scaling.

## Scaling Considerations

- Worker nodes can be horizontally scaled based on queue load
- Read-heavy operations can utilize database replicas
- API layer can be scaled horizontally behind a load balancer
- Implement caching for frequently accessed data
- Use connection pooling for database connections

## Security Considerations

- Implement API authentication and authorization
- Secure database connections
- Use HTTPS for all external communications
- Implement rate limiting to prevent abuse
- Regular security audits and dependency updates