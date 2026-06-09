'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import type { WSMessage } from '@/types/workflow'

interface Options {
  runId: string | null
  tenantId: string
  token: string
  onMessage?: (msg: WSMessage) => void
  onConnect?: () => void
  onDisconnect?: () => void
}

export function useWorkflowWS({ runId, tenantId, token, onMessage, onConnect, onDisconnect }: Options) {
  const [connected, setConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectCountRef = useRef(0)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const intentionalCloseRef = useRef(false)

  const onMessageRef = useRef(onMessage)
  const onConnectRef = useRef(onConnect)
  const onDisconnectRef = useRef(onDisconnect)
  onMessageRef.current = onMessage
  onConnectRef.current = onConnect
  onDisconnectRef.current = onDisconnect

  const connect = useCallback(() => {
    if (!runId || !token) return

    const query = new URLSearchParams({ token, tenant_id: tenantId })
    const wsUrl = `${process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'}/api/v1/ws/run/${encodeURIComponent(runId)}?${query.toString()}`
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      setConnected(true)
      intentionalCloseRef.current = false
      reconnectCountRef.current = 0
      onConnectRef.current?.()
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as WSMessage
        setLastMessage(msg)
        onMessageRef.current?.(msg)
      } catch {
        // ignore
      }
    }

    ws.onclose = () => {
      setConnected(false)
      onDisconnectRef.current?.()

      if (intentionalCloseRef.current) return

      const delay = Math.min(1000 * 2 ** reconnectCountRef.current, 30000)
      reconnectCountRef.current++
      timerRef.current = setTimeout(() => {
        connect()
      }, delay)
    }

    ws.onerror = () => {
      ws.close()
    }

    wsRef.current = ws
  }, [runId, tenantId, token])

  const disconnect = useCallback(() => {
    intentionalCloseRef.current = true
    if (timerRef.current) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setConnected(false)
  }, [])

  useEffect(() => {
    intentionalCloseRef.current = false
    connect()
    return () => {
      intentionalCloseRef.current = true
      disconnect()
    }
  }, [connect, disconnect])

  return { connected, lastMessage, disconnect }
}
