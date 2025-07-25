import React, { useState, useEffect, useRef } from 'react'
import { 
  DocumentTextIcon,
  ArrowDownIcon,
  XMarkIcon,
  PauseIcon,
  PlayIcon
} from '@heroicons/react/24/outline'

export default function LogViewer({ services }) {
  const [selectedService, setSelectedService] = useState(null)
  const [logs, setLogs] = useState([])
  const [isConnected, setIsConnected] = useState(false)
  const [isPaused, setIsPaused] = useState(false)
  const [autoScroll, setAutoScroll] = useState(true)
  const logsEndRef = useRef(null)
  const wsRef = useRef(null)

  const allServices = [
    ...(services?.docker || []),
    ...(services?.vms || []),
    ...(services?.mocks || [])
  ]

  useEffect(() => {
    if (selectedService && !isPaused) {
      connectWebSocket(selectedService.name)
    }

    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [selectedService, isPaused])

  useEffect(() => {
    if (autoScroll && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [logs, autoScroll])

  const connectWebSocket = (serviceName) => {
    if (wsRef.current) {
      wsRef.current.close()
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws/logs/${serviceName}`
    
    wsRef.current = new WebSocket(wsUrl)

    wsRef.current.onopen = () => {
      setIsConnected(true)
      console.log(`Connected to logs for ${serviceName}`)
    }

    wsRef.current.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'log' && data.message) {
          setLogs(prev => [...prev, {
            id: Date.now() + Math.random(),
            message: data.message,
            timestamp: new Date().toISOString(),
            type: getLogType(data.message)
          }])
        }
      } catch (error) {
        console.error('Error parsing log message:', error)
      }
    }

    wsRef.current.onclose = () => {
      setIsConnected(false)
      console.log(`Disconnected from logs for ${serviceName}`)
    }

    wsRef.current.onerror = (error) => {
      console.error('WebSocket error:', error)
      setIsConnected(false)
    }
  }

  const getLogType = (message) => {
    const lowerMessage = message.toLowerCase()
    if (lowerMessage.includes('error') || lowerMessage.includes('failed') || lowerMessage.includes('exception')) {
      return 'error'
    }
    if (lowerMessage.includes('warn') || lowerMessage.includes('warning')) {
      return 'warning'
    }
    return 'info'
  }

  const getLogLineClass = (type) => {
    switch (type) {
      case 'error':
        return 'log-line error'
      case 'warning':
        return 'log-line warning'
      default:
        return 'log-line'
    }
  }

  const clearLogs = () => {
    setLogs([])
  }

  const togglePause = () => {
    setIsPaused(!isPaused)
  }

  const scrollToBottom = () => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Log Viewer</h1>
        <div className="flex items-center space-x-2">
          {selectedService && (
            <>
              <button
                onClick={togglePause}
                className={`flex items-center px-3 py-2 rounded-md transition-colors ${
                  isPaused 
                    ? 'bg-green-600 text-white hover:bg-green-700' 
                    : 'bg-yellow-600 text-white hover:bg-yellow-700'
                }`}
              >
                {isPaused ? (
                  <>
                    <PlayIcon className="w-4 h-4 mr-1" />
                    Resume
                  </>
                ) : (
                  <>
                    <PauseIcon className="w-4 h-4 mr-1" />
                    Pause
                  </>
                )}
              </button>
              <button
                onClick={clearLogs}
                className="px-3 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors"
              >
                Clear
              </button>
            </>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Service Selection */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg shadow">
            <div className="px-4 py-3 border-b border-gray-200">
              <h2 className="font-medium text-gray-900">Services</h2>
            </div>
            <div className="p-4">
              <div className="space-y-2">
                {allServices.map((service, index) => (
                  <button
                    key={`${service.name}-${index}`}
                    onClick={() => {
                      setSelectedService(service)
                      setLogs([])
                    }}
                    className={`w-full text-left px-3 py-2 rounded-md transition-colors ${
                      selectedService?.name === service.name
                        ? 'bg-blue-100 text-blue-700'
                        : 'text-gray-700 hover:bg-gray-100'
                    }`}
                  >
                    <div className="flex items-center">
                      <DocumentTextIcon className="w-4 h-4 mr-2" />
                      <span className="font-medium">{service.name}</span>
                    </div>
                    <div className="text-xs text-gray-500 ml-6">
                      {service.type || 'docker'}
                    </div>
                  </button>
                ))}
              </div>
              
              {allServices.length === 0 && (
                <p className="text-gray-500 text-sm text-center py-4">
                  No services available
                </p>
              )}
            </div>
          </div>
        </div>

        {/* Log Display */}
        <div className="lg:col-span-3">
          <div className="bg-white rounded-lg shadow h-96">
            <div className="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
              <div className="flex items-center">
                <h2 className="font-medium text-gray-900">
                  {selectedService ? `Logs: ${selectedService.name}` : 'Select a service'}
                </h2>
                {selectedService && (
                  <div className="ml-3 flex items-center">
                    <div className={`w-2 h-2 rounded-full mr-2 ${
                      isConnected ? 'bg-green-400' : 'bg-red-400'
                    }`}></div>
                    <span className="text-sm text-gray-500">
                      {isConnected ? 'Connected' : 'Disconnected'}
                    </span>
                  </div>
                )}
              </div>
              
              <div className="flex items-center space-x-2">
                <label className="flex items-center text-sm text-gray-600">
                  <input
                    type="checkbox"
                    checked={autoScroll}
                    onChange={(e) => setAutoScroll(e.target.checked)}
                    className="mr-1"
                  />
                  Auto-scroll
                </label>
                {logs.length > 0 && (
                  <button
                    onClick={scrollToBottom}
                    className="p-1 text-gray-600 hover:text-gray-900 transition-colors"
                    title="Scroll to bottom"
                  >
                    <ArrowDownIcon className="w-4 h-4" />
                  </button>
                )}
              </div>
            </div>
            
            <div className="h-80 overflow-y-auto custom-scrollbar bg-gray-50">
              {selectedService ? (
                logs.length > 0 ? (
                  <div className="p-2">
                    {logs.map((log) => (
                      <div key={log.id} className={getLogLineClass(log.type)}>
                        <span className="text-gray-500 text-xs mr-2">
                          {new Date(log.timestamp).toLocaleTimeString()}
                        </span>
                        <span>{log.message}</span>
                      </div>
                    ))}
                    <div ref={logsEndRef} />
                  </div>
                ) : (
                  <div className="flex items-center justify-center h-full text-gray-500">
                    {isPaused ? 'Logs paused' : 'Waiting for logs...'}
                  </div>
                )
              ) : (
                <div className="flex items-center justify-center h-full text-gray-500">
                  Select a service to view logs
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}