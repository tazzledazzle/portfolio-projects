import React, { useState, useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Sidebar from '../components/Sidebar'
import Dashboard from '../components/Dashboard'
import ProfileManager from '../components/ProfileManager'
import ServiceManager from '../components/ServiceManager'
import LogViewer from '../components/LogViewer'
import MockServiceManager from '../components/MockServiceManager'
import Terminal from '../components/Terminal'

function App() {
  const [currentView, setCurrentView] = useState('dashboard')
  const [profiles, setProfiles] = useState([])
  const [services, setServices] = useState({ docker: [], vms: [], mocks: [] })
  const [selectedProfile, setSelectedProfile] = useState(null)

  useEffect(() => {
    fetchProfiles()
    fetchServices()
    
    // Set up periodic refresh
    const interval = setInterval(() => {
      fetchServices()
    }, 5000)
    
    return () => clearInterval(interval)
  }, [])

  const fetchProfiles = async () => {
    try {
      const response = await fetch('/api/profiles')
      const data = await response.json()
      setProfiles(data.profiles || [])
    } catch (error) {
      console.error('Failed to fetch profiles:', error)
    }
  }

  const fetchServices = async () => {
    try {
      const response = await fetch('/api/services')
      const data = await response.json()
      setServices(data)
    } catch (error) {
      console.error('Failed to fetch services:', error)
    }
  }

  const handleProfileSelect = (profile) => {
    setSelectedProfile(profile)
    setCurrentView('profile')
  }

  const renderCurrentView = () => {
    switch (currentView) {
      case 'dashboard':
        return (
          <Dashboard 
            profiles={profiles}
            services={services}
            onProfileSelect={handleProfileSelect}
            onRefresh={() => {
              fetchProfiles()
              fetchServices()
            }}
          />
        )
      case 'profiles':
        return (
          <ProfileManager 
            profiles={profiles}
            onProfileSelect={handleProfileSelect}
            onRefresh={fetchProfiles}
          />
        )
      case 'profile':
        return selectedProfile ? (
          <ServiceManager 
            profile={selectedProfile}
            onBack={() => setCurrentView('dashboard')}
            onRefresh={fetchServices}
          />
        ) : (
          <div>No profile selected</div>
        )
      case 'services':
        return (
          <ServiceManager 
            services={services}
            onRefresh={fetchServices}
          />
        )
      case 'logs':
        return <LogViewer services={services} />
      case 'mocks':
        return (
          <MockServiceManager 
            services={services.mocks}
            onRefresh={fetchServices}
          />
        )
      case 'terminal':
        return <Terminal />
      default:
        return <Dashboard profiles={profiles} services={services} />
    }
  }

  return (
    <div className="flex h-screen bg-gray-100">
      <Sidebar 
        currentView={currentView}
        onViewChange={setCurrentView}
        profiles={profiles}
      />
      
      <main className="flex-1 overflow-hidden">
        <div className="h-full p-6">
          {renderCurrentView()}
        </div>
      </main>
    </div>
  )
}

export default App
