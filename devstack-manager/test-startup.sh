#!/bin/bash

echo "Testing DevStack Manager startup..."

# Stop any existing containers
docker-compose down

# Build and start the services
echo "Building and starting services..."
docker-compose up --build -d

# Wait for services to start
echo "Waiting for services to start..."
sleep 10

# Test backend health
echo "Testing backend health..."
curl -f http://localhost:8000/api/health || {
    echo "Backend health check failed"
    docker-compose logs backend
    exit 1
}

# Test frontend
echo "Testing frontend..."
curl -f http://localhost:5173 || {
    echo "Frontend check failed"
    docker-compose logs frontend
    exit 1
}

echo "✓ All services started successfully!"

# Show logs
echo "Backend logs:"
docker-compose logs backend | tail -10

echo "Frontend logs:"
docker-compose logs frontend | tail -10

echo "Services are running at:"
echo "  Frontend: http://localhost:5173"
echo "  Backend:  http://localhost:8000"
echo "  API Docs: http://localhost:8000/docs"