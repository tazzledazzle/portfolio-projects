import { useEffect, useRef, useState } from 'react'

export default function useWebSocketLog(containerName) {
  const [logs, setLogs] = useState([])
  const ws = useRef(null)

  useEffect(() => {
    if (!containerName) return
    ws.current = new WebSocket(`ws://localhost:8000/ws/logs/${containerName}`)
    ws.current.onmessage = (e) => setLogs(prev => [...prev, e.data])
    return () => ws.current && ws.current.close()
  }, [containerName])

  return logs
}
