import React, { useEffect, useRef, useState } from 'react'
import { Terminal as XTerm } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'

export default function Terminal() {
  const terminalRef = useRef(null)
  const xtermRef = useRef(null)
  const fitAddonRef = useRef(null)
  const [isConnected, setIsConnected] = useState(false)

  useEffect(() => {
    if (terminalRef.current && !xtermRef.current) {
      // Initialize xterm
      xtermRef.current = new XTerm({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
        theme: {
          background: '#1e1e1e',
          foreground: '#ffffff',
          cursor: '#ffffff',
          selection: '#3a3a3a'
        }
      })

      // Initialize fit addon
      fitAddonRef.current = new FitAddon()
      xtermRef.current.loadAddon(fitAddonRef.current)

      // Open terminal
      xtermRef.current.open(terminalRef.current)
      fitAddonRef.current.fit()

      // Welcome message
      xtermRef.current.writeln('Welcome to DevStack Manager Terminal')
      xtermRef.current.writeln('This is a simulated terminal for demonstration purposes.')
      xtermRef.current.writeln('')
      xtermRef.current.write('$ ')

      // Handle input
      let currentLine = ''
      xtermRef.current.onData((data) => {
        const code = data.charCodeAt(0)
        
        if (code === 13) { // Enter
          xtermRef.current.writeln('')
          handleCommand(currentLine.trim())
          currentLine = ''
          xtermRef.current.write('$ ')
        } else if (code === 127) { // Backspace
          if (currentLine.length > 0) {
            currentLine = currentLine.slice(0, -1)
            xtermRef.current.write('\b \b')
          }
        } else if (code >= 32) { // Printable characters
          currentLine += data
          xtermRef.current.write(data)
        }
      })

      // Handle resize
      const handleResize = () => {
        if (fitAddonRef.current) {
          fitAddonRef.current.fit()
        }
      }

      window.addEventListener('resize', handleResize)

      return () => {
        window.removeEventListener('resize', handleResize)
        if (xtermRef.current) {
          xtermRef.current.dispose()
        }
      }
    }
  }, [])

  const handleCommand = (command) => {
    const term = xtermRef.current
    
    switch (command.toLowerCase()) {
      case 'help':
        term.writeln('Available commands:')
        term.writeln('  help          - Show this help message')
        term.writeln('  clear         - Clear the terminal')
        term.writeln('  docker ps     - List Docker containers')
        term.writeln('  docker logs   - Show container logs')
        term.writeln('  ls            - List files (simulated)')
        term.writeln('  pwd           - Show current directory')
        term.writeln('  whoami        - Show current user')
        break
        
      case 'clear':
        term.clear()
        break
        
      case 'docker ps':
        term.writeln('CONTAINER ID   IMAGE          COMMAND                  CREATED         STATUS         PORTS                    NAMES')
        term.writeln('abc123def456   nginx:latest   "/docker-entrypoint.…"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp     web-server')
        term.writeln('def456ghi789   postgres:13    "docker-entrypoint.s…"   2 hours ago     Up 2 hours     0.0.0.0:5432->5432/tcp   database')
        break
        
      case 'docker logs':
        term.writeln('Usage: docker logs <container_name>')
        term.writeln('Example: docker logs web-server')
        break
        
      case 'ls':
        term.writeln('devstack-manager/')
        term.writeln('├── backend/')
        term.writeln('├── frontend/')
        term.writeln('├── common/')
        term.writeln('├── docker-compose.yml')
        term.writeln('└── package.json')
        break
        
      case 'pwd':
        term.writeln('/workspace/devstack-manager')
        break
        
      case 'whoami':
        term.writeln('developer')
        break
        
      case '':
        // Empty command, do nothing
        break
        
      default:
        if (command.startsWith('docker logs ')) {
          const containerName = command.split(' ')[2]
          term.writeln(`Logs for container: ${containerName}`)
          term.writeln('2024-01-15 10:30:15 [INFO] Application started')
          term.writeln('2024-01-15 10:30:16 [INFO] Server listening on port 8080')
          term.writeln('2024-01-15 10:30:17 [INFO] Database connection established')
        } else {
          term.writeln(`Command not found: ${command}`)
          term.writeln('Type "help" for available commands')
        }
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Terminal</h1>
        <div className="flex items-center space-x-2">
          <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-400' : 'bg-gray-400'}`}></div>
          <span className="text-sm text-gray-600">
            {isConnected ? 'Connected' : 'Local Terminal'}
          </span>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="bg-gray-800 px-4 py-2 flex items-center space-x-2">
          <div className="w-3 h-3 bg-red-500 rounded-full"></div>
          <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
          <div className="w-3 h-3 bg-green-500 rounded-full"></div>
          <span className="text-gray-300 text-sm ml-4">Terminal</span>
        </div>
        
        <div className="terminal-container">
          <div
            ref={terminalRef}
            className="w-full h-96"
            style={{ height: '400px' }}
          />
        </div>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h3 className="font-medium text-blue-900 mb-2">Terminal Features</h3>
        <ul className="text-sm text-blue-800 space-y-1">
          <li>• Type "help" to see available commands</li>
          <li>• Use "docker ps" to list containers</li>
          <li>• Use "docker logs &lt;container&gt;" to view logs</li>
          <li>• This is a simulated terminal for demonstration</li>
        </ul>
      </div>
    </div>
  )
}