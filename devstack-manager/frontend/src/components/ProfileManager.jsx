import React, { useState } from 'react'
import { 
  PlayIcon, 
  StopIcon, 
  EyeIcon,
  PlusIcon,
  TrashIcon
} from '@heroicons/react/24/outline'

export default function ProfileManager({ profiles, onProfileSelect, onRefresh }) {
  const [selectedProfile, setSelectedProfile] = useState(null)
  const [isStarting, setIsStarting] = useState({})

  const startProfile = async (profileName) => {
    setIsStarting(prev => ({ ...prev, [profileName]: true }))
    try {
      const response = await fetch(`/api/profiles/${profileName}/start`, {
        method: 'POST'
      })
      const result = await response.json()
      console.log('Profile start result:', result)
      onRefresh()
    } catch (error) {
      console.error('Failed to start profile:', error)
    } finally {
      setIsStarting(prev => ({ ...prev, [profileName]: false }))
    }
  }

  const stopProfile = async (profileName) => {
    setIsStarting(prev => ({ ...prev, [profileName]: true }))
    try {
      const response = await fetch(`/api/profiles/${profileName}/stop`, {
        method: 'POST'
      })
      const result = await response.json()
      console.log('Profile stop result:', result)
      onRefresh()
    } catch (error) {
      console.error('Failed to stop profile:', error)
    } finally {
      setIsStarting(prev => ({ ...prev, [profileName]: false }))
    }
  }

  const viewProfile = async (profileName) => {
    try {
      const response = await fetch(`/api/profiles/${profileName}`)
      const profile = await response.json()
      setSelectedProfile(profile)
    } catch (error) {
      console.error('Failed to fetch profile details:', error)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Environment Profiles</h1>
        <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors">
          <PlusIcon className="w-4 h-4 mr-2" />
          New Profile
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Profiles List */}
        <div className="bg-white rounded-lg shadow">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Available Profiles</h2>
          </div>
          <div className="p-6">
            {profiles.length === 0 ? (
              <p className="text-gray-500 text-center py-8">No profiles configured</p>
            ) : (
              <div className="space-y-4">
                {profiles.map((profile) => (
                  <div
                    key={profile.name}
                    className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="font-medium text-gray-900">{profile.name}</h3>
                        <p className="text-sm text-gray-600 mt-1">{profile.description}</p>
                        <p className="text-xs text-gray-500 mt-2">
                          {profile.services_count} services • Version {profile.version || '1.0'}
                        </p>
                      </div>
                      <div className="flex space-x-2">
                        <button
                          onClick={() => viewProfile(profile.name)}
                          className="p-2 text-gray-600 hover:text-blue-600 transition-colors"
                          title="View Details"
                        >
                          <EyeIcon className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => startProfile(profile.name)}
                          disabled={isStarting[profile.name]}
                          className="p-2 text-green-600 hover:text-green-700 transition-colors disabled:opacity-50"
                          title="Start Profile"
                        >
                          <PlayIcon className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => stopProfile(profile.name)}
                          disabled={isStarting[profile.name]}
                          className="p-2 text-red-600 hover:text-red-700 transition-colors disabled:opacity-50"
                          title="Stop Profile"
                        >
                          <StopIcon className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                    
                    {isStarting[profile.name] && (
                      <div className="mt-3 flex items-center text-sm text-blue-600">
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
                        Processing...
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Profile Details */}
        <div className="bg-white rounded-lg shadow">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Profile Details</h2>
          </div>
          <div className="p-6">
            {selectedProfile ? (
              <div className="space-y-4">
                <div>
                  <h3 className="font-medium text-gray-900">{selectedProfile.profile?.name}</h3>
                  <p className="text-sm text-gray-600">{selectedProfile.profile?.description}</p>
                </div>

                <div>
                  <h4 className="font-medium text-gray-900 mb-2">Services ({selectedProfile.services?.length || 0})</h4>
                  <div className="space-y-2">
                    {selectedProfile.services?.map((service, index) => (
                      <div key={index} className="flex items-center justify-between p-3 bg-gray-50 rounded">
                        <div>
                          <span className="font-medium">{service.name}</span>
                          <span className="ml-2 text-sm text-gray-500">({service.type})</span>
                        </div>
                        <div className="text-sm text-gray-600">
                          {service.ports?.join(', ') || 'No ports'}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {selectedProfile.hooks && (
                  <div>
                    <h4 className="font-medium text-gray-900 mb-2">Hooks</h4>
                    <div className="space-y-2 text-sm">
                      {selectedProfile.hooks.pre_start && (
                        <div>
                          <span className="font-medium text-gray-700">Pre-start:</span>
                          <ul className="ml-4 mt-1 space-y-1">
                            {selectedProfile.hooks.pre_start.map((hook, index) => (
                              <li key={index} className="text-gray-600 font-mono text-xs bg-gray-100 p-1 rounded">
                                {hook}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}
                      {selectedProfile.hooks.post_start && (
                        <div>
                          <span className="font-medium text-gray-700">Post-start:</span>
                          <ul className="ml-4 mt-1 space-y-1">
                            {selectedProfile.hooks.post_start.map((hook, index) => (
                              <li key={index} className="text-gray-600 font-mono text-xs bg-gray-100 p-1 rounded">
                                {hook}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}
                    </div>
                  </div>
                )}

                <div className="pt-4 border-t">
                  <button
                    onClick={() => onProfileSelect(selectedProfile)}
                    className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                  >
                    Manage Services
                  </button>
                </div>
              </div>
            ) : (
              <p className="text-gray-500 text-center py-8">Select a profile to view details</p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}