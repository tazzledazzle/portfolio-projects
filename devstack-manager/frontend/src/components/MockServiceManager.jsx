import React, { useState } from 'react'
import { 
  PlusIcon,
  TrashIcon,
  PlayIcon,
  StopIcon,
  PencilIcon,
  CloudIcon
} from '@heroicons/react/24/outline'

export default function MockServiceManager({ services, onRefresh }) {
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [editingService, setEditingService] = useState(null)
  const [newService, setNewService] = useState({
    name: '',
    port: 9000,
    routes: {}
  })
  const [newRoute, setNewRoute] = useState({
    path: '',
    method: 'GET',
    status: 200,
    headers: {},
    body: {}
  })

  const createMockService = async () => {
    try {
      const response = await fetch('/api/mock-services', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(newService)
      })
      
      if (response.ok) {
        setShowCreateForm(false)
        setNewService({ name: '', port: 9000, routes: {} })
        onRefresh()
      }
    } catch (error) {
      console.error('Failed to create mock service:', error)
    }
  }

  const deleteMockService = async (serviceName) => {
    try {
      const response = await fetch(`/api/mock-services/${serviceName}`, {
        method: 'DELETE'
      })
      
      if (response.ok) {
        onRefresh()
      }
    } catch (error) {
      console.error('Failed to delete mock service:', error)
    }
  }

  const toggleMockService = async (serviceName, action) => {
    try {
      const response = await fetch(`/api/services/${serviceName}/${action}?service_type=mock`, {
        method: 'POST'
      })
      
      if (response.ok) {
        onRefresh()
      }
    } catch (error) {
      console.error(`Failed to ${action} mock service:`, error)
    }
  }

  const addRoute = () => {
    if (newRoute.path) {
      setNewService(prev => ({
        ...prev,
        routes: {
          ...prev.routes,
          [newRoute.path]: {
            method: newRoute.method,
            status: newRoute.status,
            headers: newRoute.headers,
            body: newRoute.body
          }
        }
      }))
      setNewRoute({
        path: '',
        method: 'GET',
        status: 200,
        headers: {},
        body: {}
      })
    }
  }

  const removeRoute = (path) => {
    setNewService(prev => {
      const routes = { ...prev.routes }
      delete routes[path]
      return { ...prev, routes }
    })
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Mock Services</h1>
        <button
          onClick={() => setShowCreateForm(true)}
          className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
        >
          <PlusIcon className="w-4 h-4 mr-2" />
          New Mock Service
        </button>
      </div>

      {/* Create/Edit Form */}
      {showCreateForm && (
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">
            {editingService ? 'Edit Mock Service' : 'Create Mock Service'}
          </h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Service Name
              </label>
              <input
                type="text"
                value={newService.name}
                onChange={(e) => setNewService(prev => ({ ...prev, name: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="my-mock-service"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Port
              </label>
              <input
                type="number"
                value={newService.port}
                onChange={(e) => setNewService(prev => ({ ...prev, port: parseInt(e.target.value) }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="9000"
              />
            </div>
          </div>

          {/* Routes Section */}
          <div className="mb-6">
            <h3 className="text-md font-medium text-gray-900 mb-3">Routes</h3>
            
            {/* Add Route Form */}
            <div className="bg-gray-50 p-4 rounded-md mb-4">
              <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-3">
                <input
                  type="text"
                  value={newRoute.path}
                  onChange={(e) => setNewRoute(prev => ({ ...prev, path: e.target.value }))}
                  className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="/api/endpoint"
                />
                
                <select
                  value={newRoute.method}
                  onChange={(e) => setNewRoute(prev => ({ ...prev, method: e.target.value }))}
                  className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="GET">GET</option>
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                  <option value="DELETE">DELETE</option>
                  <option value="PATCH">PATCH</option>
                </select>
                
                <input
                  type="number"
                  value={newRoute.status}
                  onChange={(e) => setNewRoute(prev => ({ ...prev, status: parseInt(e.target.value) }))}
                  className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="200"
                />
                
                <button
                  onClick={addRoute}
                  className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors"
                >
                  Add Route
                </button>
              </div>
              
              <textarea
                value={typeof newRoute.body === 'string' ? newRoute.body : JSON.stringify(newRoute.body, null, 2)}
                onChange={(e) => {
                  try {
                    const parsed = JSON.parse(e.target.value)
                    setNewRoute(prev => ({ ...prev, body: parsed }))
                  } catch {
                    setNewRoute(prev => ({ ...prev, body: e.target.value }))
                  }
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                rows="3"
                placeholder='{"message": "Hello World"}'
              />
            </div>

            {/* Existing Routes */}
            <div className="space-y-2">
              {Object.entries(newService.routes).map(([path, config]) => (
                <div key={path} className="flex items-center justify-between p-3 bg-white border border-gray-200 rounded-md">
                  <div className="flex items-center space-x-4">
                    <span className={`px-2 py-1 text-xs font-medium rounded ${
                      config.method === 'GET' ? 'bg-blue-100 text-blue-800' :
                      config.method === 'POST' ? 'bg-green-100 text-green-800' :
                      config.method === 'PUT' ? 'bg-yellow-100 text-yellow-800' :
                      config.method === 'DELETE' ? 'bg-red-100 text-red-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {config.method}
                    </span>
                    <span className="font-medium">{path}</span>
                    <span className="text-sm text-gray-500">→ {config.status}</span>
                  </div>
                  <button
                    onClick={() => removeRoute(path)}
                    className="p-1 text-red-600 hover:text-red-800 transition-colors"
                  >
                    <TrashIcon className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-end space-x-3">
            <button
              onClick={() => {
                setShowCreateForm(false)
                setEditingService(null)
                setNewService({ name: '', port: 9000, routes: {} })
              }}
              className="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={createMockService}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
            >
              {editingService ? 'Update' : 'Create'} Service
            </button>
          </div>
        </div>
      )}

      {/* Services List */}
      <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
        {services.map((service, index) => (
          <div key={`${service.name}-${index}`} className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center">
                <CloudIcon className="w-8 h-8 text-purple-600 mr-3" />
                <div>
                  <h3 className="font-semibold text-gray-900">{service.name}</h3>
                  <p className="text-sm text-gray-500">Port: {service.port}</p>
                </div>
              </div>
              <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                service.status === 'running' 
                  ? 'bg-green-100 text-green-800'
                  : 'bg-red-100 text-red-800'
              }`}>
                {service.status}
              </span>
            </div>

            <div className="mb-4">
              <h4 className="text-sm font-medium text-gray-700 mb-2">Routes ({service.routes?.length || 0})</h4>
              <div className="space-y-1 max-h-32 overflow-y-auto">
                {service.routes?.map((route, routeIndex) => (
                  <div key={routeIndex} className="text-xs text-gray-600 bg-gray-50 p-2 rounded">
                    {route}
                  </div>
                )) || <p className="text-xs text-gray-500">No routes configured</p>}
              </div>
            </div>

            <div className="flex space-x-2">
              <button
                onClick={() => toggleMockService(service.name, 'start')}
                disabled={service.status === 'running'}
                className="flex-1 flex items-center justify-center px-3 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                <PlayIcon className="w-4 h-4 mr-1" />
                Start
              </button>
              
              <button
                onClick={() => toggleMockService(service.name, 'stop')}
                disabled={service.status === 'stopped'}
                className="flex-1 flex items-center justify-center px-3 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                <StopIcon className="w-4 h-4 mr-1" />
                Stop
              </button>
              
              <button
                onClick={() => deleteMockService(service.name)}
                className="px-3 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors"
                title="Delete Service"
              >
                <TrashIcon className="w-4 h-4" />
              </button>
            </div>
          </div>
        ))}
      </div>

      {services.length === 0 && !showCreateForm && (
        <div className="text-center py-12">
          <CloudIcon className="w-12 h-12 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-500">No mock services configured</p>
          <button
            onClick={() => setShowCreateForm(true)}
            className="mt-4 text-blue-600 hover:text-blue-800 font-medium"
          >
            Create your first mock service
          </button>
        </div>
      )}
    </div>
  )
}