import { Listbox } from '@headlessui/react'
import { useState, useEffect } from 'react'

export default function ProfileSelector({ onSelect }) {
  const [profiles, setProfiles] = useState([])
  const [selected, setSelected] = useState(null)

  useEffect(() => {
    fetch('http://localhost:8000/api/profiles')
      .then(res => res.json())
      .then(data => {
        setProfiles(data)
        if (data.length) {
          setSelected(data[0])
          onSelect(data[0])
        }
      })
  }, [])

  return (
    <div className="w-64 mb-4">
      <Listbox value={selected} onChange={(profile) => {
        setSelected(profile)
        onSelect(profile)
      }}>
        <Listbox.Button className="w-full py-2 px-4 border rounded">{selected?.name || 'Select Profile'}</Listbox.Button>
        <Listbox.Options className="mt-1 border rounded bg-white">
          {profiles.map((profile, idx) => (
            <Listbox.Option key={idx} value={profile} className="cursor-pointer p-2 hover:bg-gray-100">
              {profile.name}
            </Listbox.Option>
          ))}
        </Listbox.Options>
      </Listbox>
    </div>
  )
}
