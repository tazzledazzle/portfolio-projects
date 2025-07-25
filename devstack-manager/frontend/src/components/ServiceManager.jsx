import React, { useState, useEffect } from 'react'
import { 
  PlayIcon, 
  StopIcon, 
  ArrowPathIcon,
  DocumentTextIcon,
  CommandLineIcon,
  ChartBarIcon,
  ArrowLeftIcon
} from '@heroicons/react/24/outline'

export default function ServiceManager({ profile, services, onBack, onRefresh }) {
  const [serviceStatuses, setServiceStatuses] = useState({})
  const [isLoading, setIsLoading] = useState({})

  useEffect(() => {
    if (profile) {
      fetchProfileStatus()
    }
  }, [profile])

  const fetchProfileStatus = async () => {
    if (!profile?.profile?.name) return
    
    try {
      const response = await fetch(`/api/profiles/${profile.profile.name}/status`)
      if (response.ok) {
        const data = await response.json()
        const statusMap = {}
        data.services?.forEach(service => {
          statusMap[service.name] = service
        })
        setServiceStatuses(statusMap)
      }
    } catch (error) {
      console.error('Failed to fetch profile status:', error)
    }
  }

  const toggleService = async (serviceName, action, serviceType = 'docker') => {
    setIsLoading(prev => ({ ...prev, [serviceName]: true }))
    
    try {
      const response = await fetch(`/api/services/${serviceName}/${action}?service_type=${serviceType}`, {
        method: 'POST'
      })
      const result = await response.json()
      console.log(`Service ${action} result:`, result)
      
      // Refresh status
      await fetchProfileStatus()
      if (onRefresh) onRefresh()
    } catch (error) {
      console.error(`Failed to ${action} service:`, error)
    } finally {
      setIsLoading(prev => ({ ...prev, [serviceName]: false }))
    }
  }

  const getStatusColor = (status) => {
    switch (status?.toLowerCase()) {
      case 'running':
      case 'started':
        return 'text-green-600 bg-green-100'
      case 'stopped':
      case 'exited':
        return 'text-red-600 bg-red-100'
      case 'starting':
        return 'text-yellow-600 bg-yellow-100'
      default:
        return 'text-gray-600 bg-gray-100'
    }
  }

  const getServiceIcon = (type) => {
    switch (type) {
      case 'docker':
        return '🐳'
      case 'vm':
        return '💻'
      case 'mock':
        return '🎭'
      default:
        return '⚙️'
    }
  }

  const allServices = profile ? profile.services : [
    ...(services?.docker || []),
    ...(services?.vms || []),
    ...(services?.mocks || [])
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center">
          {onBack && (
            <button
              onClick={onBack}
              className="mr-4 p-2 text-gray-600 hover:text-gray-900 transition-colors"
            >
              <ArrowLeftIcon className="w-5 h-5" />
            </button>
          )}
          <div>
            <h1 className="text-2xl font-bold text-gray-900">
              {profile ? `${profile.profile.name} Services` : 'Service Manager'}
            </h1>
            {profile?.profile?.description && (
              <p className="text-gray-600">{profile.profile.description}</p>
            )}
          </div>
        </div>
        <button
          onClick={() => {
            fetchProfileStatus()
            if (onRefresh) onRefresh()
          }}
          className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
        >
          <ArrowPathIcon className="w-4 h-4 mr-2" />
          Refresh
        </button>
      </div>

      {/* Services Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
        {allServices.map((service, index) => {
          const serviceName = service.name
          const serviceType = service.type || 'docker'
          const status = serviceStatuses[serviceName] || service
          const isServiceLoading = isLoading[serviceName]

          return (
            <div key={`${serviceName}-${index}`} className="bg-white rounded-lg shadow-md p-6">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center">
                  <span className="text-2xl mr-3">{getServiceIcon(serviceType)}</span>
                  <div>
                    <h3 className="font-semibold text-gray-900">{serviceName}</h3>
                    <p className="text-sm text-gray-500">{serviceType}</p>
                  </div>
                </div>
                <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(status?.status)}`}>
                  {status?.status || 'unknown'}
                </span>
              </div>

              {/* Service Details */}
              <div className="space-y-2 mb-4">
                {service.image && (
                  <div className="text-sm">
                    <span className="font-medium text-gray-700">Image:</span>
                    <span className="ml-2 text-gray-600">{service.image}</span>
                  </div>
                )}
                {service.ports && service.ports.length > 0 && (
                  <div className="text-sm">
                    <span className="font-medium text-gray-700">Ports:</span>
                    <span className="ml-2 text-gray-600">{service.ports.join(', ')}</span>
                  </div>
                )}
                {status?.cpu_usage !== undefined && (
                  <div className="text-sm">
                    <span className="font-medium text-gray-700">CPU:</span>
                    <span className="ml-2 text-gray-600">{status.cpu_usage}%</span>
                  </div>
                )}
                {status?.memory_usage?.percentage !== undefined && (
                  <div className="text-sm">
                    <span className="font-medium text-gray-700">Memory:</span>
                    <span className="ml-2 text-gray-600">
                      {status.memory_usage.usage_mb}MB ({status.memory_usage.percentage}%)
                    </span>
                  </div>
                )}
              </div>

              {/* Action Buttons */}
              <div className="flex space-x-2">
                <button
                  onClick={() => toggleService(serviceName, 'start', serviceType)}
                  disabled={isServiceLoading || status?.status === 'running'}
                  className="flex-1 flex items-center justify-center px-3 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {isServiceLoading ? (
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                  ) : (
                    <>
                      <PlayIcon className="w-4 h-4 mr-1" />
                      Start
                    </>
                  )}
                </button>
                
                <button
                  onClick={() => toggleService(serviceName, 'stop', serviceType)}
                  disabled={isServiceLoading || status?.status === 'stopped'}
                  className="flex-1 flex items-center justify-center px-3 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {isServiceLoading ? (
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                  ) : (
                    <>
                      <StopIcon className="w-4 h-4 mr-1" />
                      Stop
                    </>
                  )}
                </button>
                
                <button
                  className="px-3 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors"
                  title="View Logs"
                >
                  <DocumentTextIcon className="w-4 h-4" />
                </button>
                
                {serviceType === 'docker' && (
                  <button
                    className="px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                    title="Terminal"
                  >
                    <CommandLineIcon className="w-4 h-4" />
                  </button>
                )}
              </div>

              {/* Hot Reload Indicator */}
              {service.hot_reload && (
                <div className="mt-3 flex items-center text-sm text-blue-600">
                  <div className="w-2 h-2 bg-blue-600 rounded-full mr-2 animate-pulse"></div>
                  Hot reload enabled
                </div>
              )}
            </div>
          )
        })}
      </div>

      {allServices.length === 0 && (
        <div className="text-center py-12">
          <p className="text-gray-500">No services configured</p>
        </div>
      )}
    </div>
  )
}