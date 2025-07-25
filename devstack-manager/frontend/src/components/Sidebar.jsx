import React from 'react'
import {
  HomeIcon,
  CogIcon,
  ServerIcon,
  DocumentTextIcon,
  CommandLineIcon,
  CloudIcon
} from '@heroicons/react/24/outline'

const navigation = [
  { name: 'Dashboard', key: 'dashboard', icon: HomeIcon },
  { name: 'Profiles', key: 'profiles', icon: CogIcon },
  { name: 'Services', key: 'services', icon: ServerIcon },
  { name: 'Logs', key: 'logs', icon: DocumentTextIcon },
  { name: 'Mock Services', key: 'mocks', icon: CloudIcon },
  { name: 'Terminal', key: 'terminal', icon: CommandLineIcon },
]

export default function Sidebar({ currentView, onViewChange, profiles }) {
  return (
    <div className="w-64 bg-white shadow-lg">
      <div className="p-6">
        <h1 className="text-xl font-bold text-gray-900">DevStack Manager</h1>
      </div>
      
      <nav className="mt-6">
        <div className="px-3">
          {navigation.map((item) => {
            const Icon = item.icon
            return (
              <button
                key={item.key}
                onClick={() => onViewChange(item.key)}
                className={`w-full flex items-center px-3 py-2 text-sm font-medium rounded-md mb-1 transition-colors ${
                  currentView === item.key
                    ? 'bg-blue-100 text-blue-700'
                    : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                }`}
              >
                <Icon className="mr-3 h-5 w-5" />
                {item.name}
              </button>
            )
          })}
        </div>
        
        {profiles.length > 0 && (
          <div className="mt-8 px-3">
            <h3 className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">
              Quick Profiles
            </h3>
            <div className="mt-2 space-y-1">
              {profiles.slice(0, 5).map((profile) => (
                <button
                  key={profile.name}
                  onClick={() => onViewChange('profile')}
                  className="w-full flex items-center px-3 py-2 text-sm text-gray-600 rounded-md hover:bg-gray-50 hover:text-gray-900"
                >
                  <div className="w-2 h-2 bg-green-400 rounded-full mr-3"></div>
                  {profile.name}
                </button>
              ))}
            </div>
          </div>
        )}
      </nav>
    </div>
  )
}