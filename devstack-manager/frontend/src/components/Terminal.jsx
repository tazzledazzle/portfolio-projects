import React, { useEffect, useRef, useState } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

export default function Terminal() {
  const terminalRef = useRef(null)
  const xtermRef = useRef(null)
  const fitAddonRef = useRef(null)
  const wsRef = useRef(null)
  const [isConnected, setIsConnected] = useState(false)
  const [useWebSocket, setUseWebSocket] = useState(false)

  useEffect(() => {
    if (terminalRef.current && !xtermRef.current) {
      // Initialize xterm with proper configuration
      xtermRef.current = new XTerm({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
        rows: 24,
        cols: 80,
        theme: {
          background: '#1e1e1e',
          foreground: '#ffffff',
          cursor: '#ffffff',
          selection: '#3a3a3a',
          black: '#000000',
          red: '#ff5555',
          green: '#50fa7b',
          yellow: '#f1fa8c',
          blue: '#bd93f9',
          magenta: '#ff79c6',
          cyan: '#8be9fd',
          white: '#f8f8f2'
        },
        allowTransparency: false,
        convertEol: true
      })

      // Initialize fit addon
      fitAddonRef.current = new FitAddon()
      xtermRef.current.loadAddon(fitAddonRef.current)

      // Open terminal
      xtermRef.current.open(terminalRef.current)
      
      // Fit terminal to container
      setTimeout(() => {
        if (fitAddonRef.current) {
          fitAddonRef.current.fit()
        }
      }, 100)

      // Set connected status
      setIsConnected(true)

      // Welcome message
      xtermRef.current.writeln('\x1b[1;32m╭─────────────────────────────────────────────╮\x1b[0m')
      xtermRef.current.writeln('\x1b[1;32m│     Welcome to DevStack Manager Terminal    │\x1b[0m')
      xtermRef.current.writeln('\x1b[1;32m╰─────────────────────────────────────────────╯\x1b[0m')
      xtermRef.current.writeln('')
      xtermRef.current.writeln('\x1b[1;36mThis is a simulated terminal for demonstration purposes.\x1b[0m')
      xtermRef.current.writeln('\x1b[1;33mType "help" to see available commands.\x1b[0m')
      xtermRef.current.writeln('')
      xtermRef.current.write('\x1b[1;32mdevstack@manager\x1b[0m:\x1b[1;34m/workspace\x1b[0m$ ')

      // Handle keyboard shortcuts
      xtermRef.current.attachCustomKeyEventHandler((event) => {
        // Ctrl+L to clear terminal
        if (event.ctrlKey && event.key === 'l') {
          event.preventDefault()
          if (useWebSocket && wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify({
              type: 'input',
              data: 'clear\r'
            }))
          } else {
            xtermRef.current.clear()
            xtermRef.current.writeln('\x1b[1;32m╭─────────────────────────────────────────────╮\x1b[0m')
            xtermRef.current.writeln('\x1b[1;32m│     Welcome to DevStack Manager Terminal    │\x1b[0m')
            xtermRef.current.writeln('\x1b[1;32m╰─────────────────────────────────────────────╯\x1b[0m')
            xtermRef.current.writeln('')
            xtermRef.current.write('\x1b[1;32mdevstack@manager\x1b[0m:\x1b[1;34m/workspace\x1b[0m$ ')
          }
          return false
        }
        return true
      })

      // Handle input based on connection type
      let currentInput = ''
      xtermRef.current.onData((data) => {
        if (useWebSocket && wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
          // Send data to WebSocket terminal
          wsRef.current.send(JSON.stringify({
            type: 'input',
            data: data
          }))
        } else {
          // Handle local simulation
          const code = data.charCodeAt(0)
          if (code === 13) { // Enter
            xtermRef.current.writeln('')
            handleCommand(currentInput.trim())
            currentInput = ''
            xtermRef.current.write('\x1b[1;32mdevstack@manager\x1b[0m:\x1b[1;34m/workspace\x1b[0m$ ')
          } else if (code === 127) { // Backspace
            if (currentInput.length > 0) {
              currentInput = currentInput.slice(0, -1)
              xtermRef.current.write('\b \b')
            }
          } else if (code === 3) { // Ctrl+C
            xtermRef.current.writeln('^C')
            currentInput = ''
            xtermRef.current.write('\x1b[1;32mdevstack@manager\x1b[0m:\x1b[1;34m/workspace\x1b[0m$ ')
          } else if (code >= 32) { // Printable characters
            currentInput += data
            xtermRef.current.write(data)
          }
        }
      })

      // Handle resize
      const handleResize = () => {
        if (fitAddonRef.current && xtermRef.current) {
          setTimeout(() => {
            fitAddonRef.current.fit()
          }, 100)
        }
      }

      window.addEventListener('resize', handleResize)

      return () => {
        window.removeEventListener('resize', handleResize)
        if (wsRef.current) {
          wsRef.current.close()
        }
        if (xtermRef.current) {
          xtermRef.current.dispose()
          xtermRef.current = null
        }
      }
    }
  }, [])

  const connectWebSocket = () => {
    if (wsRef.current) {
      wsRef.current.close()
    }
    
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.hostname}:8000/ws/terminal`
    
    wsRef.current = new WebSocket(wsUrl)
    
    wsRef.current.onopen = () => {
      setIsConnected(true)
      setUseWebSocket(true)
      if (xtermRef.current) {
        xtermRef.current.clear()
        xtermRef.current.writeln('\x1b[1;32mConnected to live terminal session...\x1b[0m')
      }
    }
    
    wsRef.current.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        if (message.type === 'output' && xtermRef.current) {
          xtermRef.current.write(message.data)
        } else if (message.type === 'error' && xtermRef.current) {
          xtermRef.current.writeln(`\x1b[1;31mError: ${message.data}\x1b[0m`)
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error)
      }
    }
    
    wsRef.current.onclose = () => {
      setIsConnected(false)
      setUseWebSocket(false)
      if (xtermRef.current) {
        xtermRef.current.writeln('\x1b[1;31mWebSocket connection closed. Switched to simulation mode.\x1b[0m')
        xtermRef.current.write('\x1b[1;32mdevstack@manager\x1b[0m:\x1b[1;34m/workspace\x1b[0m$ ')
      }
    }
    
    wsRef.current.onerror = (error) => {
      console.error('WebSocket error:', error)
      setIsConnected(false)
      setUseWebSocket(false)
    }
  }

  const disconnectWebSocket = () => {
    if (wsRef.current) {
      wsRef.current.close()
    }
    setUseWebSocket(false)
    setIsConnected(false)
  }

  const handleCommand = (command) => {
    const term = xtermRef.current
    switch (command.toLowerCase()) {
      case 'help':
        term.writeln('\x1b[1;36mAvailable commands:\x1b[0m')
        term.writeln('  \x1b[1;33mhelp\x1b[0m          - Show this help message')
        term.writeln('  \x1b[1;33mclear\x1b[0m         - Clear the terminal')
        term.writeln('  \x1b[1;33mdocker ps\x1b[0m     - List Docker containers')
        term.writeln('  \x1b[1;33mdocker logs\x1b[0m   - Show container logs')
        term.writeln('  \x1b[1;33mls\x1b[0m            - List files (simulated)')
        term.writeln('  \x1b[1;33mpwd\x1b[0m           - Show current directory')
        term.writeln('  \x1b[1;33mwhoami\x1b[0m        - Show current user')
        term.writeln('  \x1b[1;33mstatus\x1b[0m        - Show DevStack Manager status')
        term.writeln('  \x1b[1;33mprofiles\x1b[0m      - List available profiles')
        break
      case 'clear':
        term.clear()
        term.writeln('\x1b[1;32m╭─────────────────────────────────────────────╮\x1b[0m')
        term.writeln('\x1b[1;32m│     Welcome to DevStack Manager Terminal    │\x1b[0m')
        term.writeln('\x1b[1;32m╰─────────────────────────────────────────────╯\x1b[0m')
        term.writeln('')
        break
      case 'docker ps':
        term.writeln('\x1b[1;37mCONTAINER ID   IMAGE          COMMAND                  CREATED         STATUS         PORTS                    NAMES\x1b[0m')
        term.writeln('\x1b[32mabc123def456\x1b[0m   nginx:latest   "/docker-entrypoint.…"   2 hours ago     \x1b[1;32mUp 2 hours\x1b[0m     0.0.0.0:8080->80/tcp     web-server')
        term.writeln('\x1b[32mdef456ghi789\x1b[0m   postgres:13    "docker-entrypoint.s…"   2 hours ago     \x1b[1;32mUp 2 hours\x1b[0m     0.0.0.0:5432->5432/tcp   database')
        term.writeln('\x1b[32mghi789jkl012\x1b[0m   redis:7-alpine "docker-entrypoint.s…"   2 hours ago     \x1b[1;32mUp 2 hours\x1b[0m     0.0.0.0:6379->6379/tcp   cache')
        break
      case 'docker logs':
        term.writeln('\x1b[1;33mUsage:\x1b[0m docker logs <container_name>')
        term.writeln('\x1b[1;33mExample:\x1b[0m docker logs web-server')
        break
      case 'ls':
        term.writeln('\x1b[1;34mdevstack-manager/\x1b[0m')
        term.writeln('├── \x1b[1;34mbackend/\x1b[0m')
        term.writeln('├── \x1b[1;34mfrontend/\x1b[0m')
        term.writeln('├── \x1b[1;34mcommon/\x1b[0m')
        term.writeln('├── \x1b[1;36mdocker-compose.yml\x1b[0m')
        term.writeln('└── \x1b[1;36mpackage.json\x1b[0m')
        break
      case 'pwd':
        term.writeln('\x1b[1;34m/workspace/devstack-manager\x1b[0m')
        break
      case 'whoami':
        term.writeln('\x1b[1;32mdeveloper\x1b[0m')
        break
      case 'status':
        term.writeln('\x1b[1;36mDevStack Manager Status:\x1b[0m')
        term.writeln('  Backend:  \x1b[1;32m●\x1b[0m Running on port 8000')
        term.writeln('  Frontend: \x1b[1;32m●\x1b[0m Running on port 5173')
        term.writeln('  Docker:   \x1b[1;33m●\x1b[0m Limited access')
        term.writeln('  Profiles: \x1b[1;32m●\x1b[0m 2 profiles loaded')
        break
      case 'profiles':
        term.writeln('\x1b[1;36mAvailable Profiles:\x1b[0m')
        term.writeln('  \x1b[1;32m●\x1b[0m my-project-dev - Development environment')
        term.writeln('  \x1b[1;32m●\x1b[0m my-project-e2e - End-to-end testing environment')
        break
      case '':
        // Empty command, do nothing
        break
      default:
        if (command.startsWith('docker logs ')) {
          const containerName = command.split(' ')[2]
          if (containerName) {
            term.writeln(`\x1b[1;36mLogs for container: \x1b[1;33m${containerName}\x1b[0m`)
            term.writeln('\x1b[90m2024-01-15 10:30:15\x1b[0m [\x1b[1;32mINFO\x1b[0m] Application started')
            term.writeln('\x1b[90m2024-01-15 10:30:16\x1b[0m [\x1b[1;32mINFO\x1b[0m] Server listening on port 8080')
            term.writeln('\x1b[90m2024-01-15 10:30:17\x1b[0m [\x1b[1;32mINFO\x1b[0m] Database connection established')
            term.writeln('\x1b[90m2024-01-15 10:30:18\x1b[0m [\x1b[1;32mINFO\x1b[0m] Ready to accept connections')
          } else {
            term.writeln('\x1b[1;31mError:\x1b[0m Container name required')
            term.writeln('\x1b[1;33mUsage:\x1b[0m docker logs <container_name>')
          }
        } else if (command.startsWith('cd ')) {
          term.writeln('\x1b[1;33mNote:\x1b[0m Directory navigation is simulated in this demo terminal')
        } else {
          term.writeln(`\x1b[1;31mCommand not found:\x1b[0m ${command}`)
          term.writeln('Type "\x1b[1;33mhelp\x1b[0m" for available commands')
        }
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Terminal</h1>
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <div className={`w-2 h-2 rounded-full ${
              useWebSocket && isConnected ? 'bg-green-400' : 
              useWebSocket ? 'bg-yellow-400' : 'bg-gray-400'
            }`}></div>
            <span className="text-sm text-gray-600">
              {useWebSocket && isConnected ? 'Live Terminal' : 
               useWebSocket ? 'Connecting...' : 'Simulation Mode'}
            </span>
          </div>
          <div className="flex space-x-2">
            {!useWebSocket ? (
              <button
                onClick={connectWebSocket}
                className="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
              >
                Connect Live
              </button>
            ) : (
              <button
                onClick={disconnectWebSocket}
                className="px-3 py-1 text-xs bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
              >
                Disconnect
              </button>
            )}
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="bg-gray-800 px-4 py-2 flex items-center space-x-2">
          <div className="w-3 h-3 bg-red-500 rounded-full"></div>
          <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
          <div className="w-3 h-3 bg-green-500 rounded-full"></div>
          <span className="text-gray-300 text-sm ml-4">Terminal</span>
        </div>
        
        <div className="terminal-container bg-gray-900 rounded-lg overflow-hidden">
          <div
            ref={terminalRef}
            className="w-full"
            style={{ 
              height: '500px',
              minHeight: '500px'
            }}
          />
        </div>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h3 className="font-medium text-blue-900 mb-2">Terminal Features</h3>
        <ul className="text-sm text-blue-800 space-y-1">
          <li>• Type "help" to see available commands</li>
          <li>• Use "docker ps" to list containers</li>
          <li>• Use "docker logs &lt;container&gt;" to view logs</li>
          <li>• Press Ctrl+L to clear the terminal</li>
          <li>• Click "Connect Live" for real terminal access</li>
        </ul>
      </div>
    </div>
  )
}