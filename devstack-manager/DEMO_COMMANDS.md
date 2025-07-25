# DevStack Manager Terminal Demo

## 🎬 **Terminal Demo Commands**

Here's a sequence of commands to demonstrate the enhanced terminal features:

### **Basic Commands Demo**
```bash
# Show welcome and help
help

# Check system status
status

# List available profiles
profiles

# Show current directory and user
pwd
whoami

# List project files
ls
```

### **Docker Integration Demo**
```bash
# List running containers
docker ps

# View container logs
docker logs web-server
docker logs database

# Try invalid container
docker logs nonexistent
```

### **Terminal Features Demo**
```bash
# Clear terminal (or use Ctrl+L)
clear

# Test command not found
invalid-command

# Test directory navigation (simulated)
cd /some/path

# Show help again with colors
help
```

### **Interactive Features**
- **Ctrl+L**: Quick clear shortcut
- **Ctrl+C**: Interrupt current command
- **Backspace**: Edit current input
- **Enter**: Execute command
- **Connect Live**: Switch to WebSocket mode
- **Disconnect**: Return to simulation mode

### **Color Coding Examples**
- 🟢 **Green**: Success messages, running status
- 🔴 **Red**: Error messages, failed operations  
- 🟡 **Yellow**: Warnings, usage information
- 🔵 **Blue**: Directory names, paths
- 🟣 **Magenta**: Container IDs, special identifiers
- 🟦 **Cyan**: File names, commands

### **WebSocket Mode Testing**
1. Click "Connect Live" button
2. Wait for connection confirmation
3. Try basic shell commands:
   ```bash
   ls -la
   pwd
   echo "Hello from live terminal!"
   ps aux
   ```
4. Click "Disconnect" to return to simulation

### **Expected Outputs**

#### Status Command
```
DevStack Manager Status:
  Backend:  ● Running on port 8000
  Frontend: ● Running on port 5173
  Docker:   ● Limited access
  Profiles: ● 2 profiles loaded
```

#### Docker PS Command
```
CONTAINER ID   IMAGE          STATUS         PORTS
abc123def456   nginx:latest   Up 2 hours     0.0.0.0:8080->80/tcp
def456ghi789   postgres:13    Up 2 hours     0.0.0.0:5432->5432/tcp
ghi789jkl012   redis:7-alpine Up 2 hours     0.0.0.0:6379->6379/tcp
```

#### Profiles Command
```
Available Profiles:
  ● my-project-dev - Development environment
  ● my-project-e2e - End-to-end testing environment
```

### **Testing Checklist**
- [ ] Terminal loads with welcome message
- [ ] All commands show colored output
- [ ] Help command displays full command list
- [ ] Docker commands show formatted container info
- [ ] Status shows service health indicators
- [ ] Ctrl+L clears terminal properly
- [ ] WebSocket connection toggles work
- [ ] Error messages display in red
- [ ] Command history works with up/down arrows (in live mode)
- [ ] Terminal resizes properly with window

This demo showcases the full range of terminal capabilities in both simulation and live modes!