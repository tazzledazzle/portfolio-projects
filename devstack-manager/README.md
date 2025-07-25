# DevStack Manager

A comprehensive UI tool for managing Docker-based and VM-based local development environments with hot-reloading and mock services. Provides developers with an intuitive interface to spin up, manage, and debug local environments mirroring production.

## 🚀 Quick Start

1. **Start the application**:

   ```bash
   cd devstack-manager
   docker compose up --build
   ```

2. **Access the interfaces**:

- Frontend: `http://localhost:5173`
- Backend API: `http://localhost:8000`
- API Documentation: `http://localhost:8000/docs`

## 🛠️ Troubleshooting

### Backend Won't Start

If you see Docker connection errors:

1. **Check Docker is running**:

```bash
docker ps
```

2.**Check Docker socket permissions**  

```bash
ls -la /var/run/docker.sock
```

3.**Test backend independently**:

```bash
cd backend
python test_startup.py
```

### Frontend Can't Connect to Backend

1. **Check backend is running**:

   ```bash
   curl http://localhost:8000/api/health
   ```

2. **Check Docker network**:

   ```bash
   docker network ls
   docker network inspect devstack-manager_devstack-network
   ```

### Services Not Showing

The application will work even if Docker is not available - it will show empty service lists and provide mock functionality.

## 📁 Project Structure

```tree
devstack-manager/
├── backend/
│   ├── app/
│   │   ├── main.py              # FastAPI application
│   │   ├── services/            # Business logic services
│   │   ├── websocket/           # WebSocket handlers
│   │   └── utils/               # Utility functions
│   ├── configs/                 # YAML profile configurations
│   ├── requirements.txt         # Python dependencies
│   └── Dockerfile              # Backend container
├── frontend/
│   ├── src/
│   │   ├── components/          # React components
│   │   └── pages/              # Main application
│   ├── package.json            # Frontend dependencies
│   ├── vite.config.js          # Build configuration
│   └── Dockerfile              # Frontend container
├── common/
│   └── config_schema.json      # Configuration validation
├── docker-compose.yml          # Multi-service orchestration
└── package.json               # Root project configuration
```

## ✨ Features

### Environment Profiles

- YAML-based configuration with services, hooks, and hot-reload support
- Pre-configured development and E2E testing environments
- Custom hook support for pre/post start/stop actions

### Multi-Service Support

- **Docker containers**: Full lifecycle management with stats and logs
- **Virtual machines**: Support for VirtualBox, libvirt, and Vagrant
- **Mock services**: Dynamic API creation and management

### Real-time Monitoring

- Live log streaming via WebSocket
- Service status and resource usage monitoring
- Health checks and uptime tracking

### Mock API Management

- Create custom mock APIs with configurable routes
- Support for all HTTP methods (GET, POST, PUT, DELETE, PATCH)
- Custom status codes, headers, and response bodies

### Modern UI

- React-based interface with Tailwind CSS
- Real-time updates and responsive design
- Terminal interface for container operations

## 🔧 Configuration

### Creating Environment Profiles

Create YAML files in `backend/configs/` directory:

```yaml
profile:
  name: "my-project-dev"
  description: "Development environment"
  version: "1.0"

services:
  - name: "api-server"
    type: "docker"
    image: "node:18-alpine"
    ports: ["8080:8080"]
    volumes: ["./src:/app/src"]
    environment:
      NODE_ENV: "development"
    hot_reload: true
    
  - name: "mock-auth"
    type: "mock"
    ports: ["9000:9000"]
    mock_config:
      routes:
        "/auth/login":
          method: "POST"
          status: 200
          body:
            token: "fake-jwt-token"

hooks:
  pre_start:
    - "echo 'Starting environment...'"
  post_start:
    - "curl -f http://localhost:8080/health"
```

### Mock Service Configuration

Mock services support:

- Custom routes with any HTTP method
- Configurable status codes and headers
- JSON response bodies
- Real-time route management via UI

## 🧪 Development

### Running Tests

```bash
# Test backend startup
cd backend
python test_startup.py

# Test full stack
./test-startup.sh
```

### Development Mode

```bash
# Backend only
cd backend
python start.py

# Frontend only
cd frontend
npm run dev
```

## 🐳 Docker Support

The application requires Docker for container management but will gracefully degrade if Docker is not available:

- **With Docker**: Full container management, log streaming, and profile execution
- **Without Docker**: Mock services, VM management (if available), and UI functionality

## 📝 API Documentation

Visit `http://localhost:8000/docs` for interactive API documentation.

Key endpoints:

- `GET /api/health` - Health check
- `GET /api/profiles` - List environment profiles
- `POST /api/profiles/{name}/start` - Start profile services
- `GET /api/services` - List all services
- `WS /ws/logs/{service}` - Stream service logs

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with `./test-startup.sh`
5. Submit a pull request

## 📄 License

ISC License - see package.json for details.
