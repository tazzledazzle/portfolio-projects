import React from 'react'
import { 
  PlayIcon, 
  StopIcon, 
  ArrowPathIcon,
  ChartBarIcon,
  ServerIcon,
  CloudIcon
} from '@heroicons/react/24/outline'

export default function Dashboard({ profiles, services, onProfileSelect, onRefresh }) {
  const totalServices = (services.docker?.length || 0) + (services.vms?.length || 0) + (services.mocks?.length || 0)
  const runningServices = [
    ...(services.docker || []).filter(s => s.status === 'running'),
    ...(services.vms || []).filter(s => s.status === 'running'),
    ...(services.mocks || []).filter(s => s.status === 'running')
  ].length

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

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <button
          onClick={onRefresh}
          className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
        >
          <ArrowPathIcon className="w-4 h-4 mr-2" />
          Refresh
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-white p-6 rounded-lg shadow">
          <div className="flex items-center">
            <ChartBarIcon className="w-8 h-8 text-blue-600" />
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Total Profiles</p>
              <p className="text-2xl font-bold text-gray-900">{profiles.length}</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <div className="flex items-center">
            <ServerIcon className="w-8 h-8 text-green-600" />
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Total Services</p>
              <p className="text-2xl font-bold text-gray-900">{totalServices}</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <div className="flex items-center">
            <PlayIcon className="w-8 h-8 text-green-600" />
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Running</p>
              <p className="text-2xl font-bold text-gray-900">{runningServices}</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <div className="flex items-center">
            <CloudIcon className="w-8 h-8 text-purple-600" />
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Mock Services</p>
              <p className="text-2xl font-bold text-gray-900">{services.mocks?.length || 0}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Profiles Section */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-medium text-gray-900">Environment Profiles</h2>
        </div>
        <div className="p-6">
          {profiles.length === 0 ? (
            <p className="text-gray-500 text-center py-8">No profiles configured</p>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {profiles.map((profile) => (
                <div
                  key={profile.name}
                  className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer"
                  onClick={() => onProfileSelect(profile)}
                >
                  <h3 className="font-medium text-gray-900">{profile.name}</h3>
                  <p className="text-sm text-gray-600 mt-1">{profile.description}</p>
                  <div className="mt-3 flex items-center justify-between">
                    <span className="text-xs text-gray-500">
                      {profile.services_count} services
                    </span>
                    <button className="text-blue-600 hover:text-blue-800 text-sm font-medium">
                      Manage →
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Recent Services */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-medium text-gray-900">Recent Services</h2>
        </div>
        <div className="p-6">
          {totalServices === 0 ? (
            <p className="text-gray-500 text-center py-8">No services running</p>
          ) : (
            <div className="space-y-3">
              {[...services.docker || [], ...services.vms || [], ...services.mocks || []]
                .slice(0, 10)
                .map((service, index) => (
                  <div key={`${service.name}-${index}`} className="flex items-center justify-between py-2">
                    <div className="flex items-center">
                      <div className="w-2 h-2 rounded-full mr-3 bg-gray-400"></div>
                      <span className="font-medium text-gray-900">{service.name}</span>
                      <span className="ml-2 text-sm text-gray-500">
                        ({service.provider || service.type || 'docker'})
                      </span>
                    </div>
                    <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(service.status)}`}>
                      {service.status}
                    </span>
                  </div>
                ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}