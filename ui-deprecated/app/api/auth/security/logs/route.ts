import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

const TARGET_API_URL = process.env.NEXT_PUBLIC_API_URL

async function handler(req: NextRequest) {
  const searchParams = req.nextUrl.search
  const targetUrl = `${TARGET_API_URL}/auth/security/logs${searchParams}`

  const headers = new Headers(req.headers)
  headers.delete('host')

  const authHeader = req.headers.get('x-auth-token')
  if (authHeader) {
    headers.set('Authorization', `Bearer ${authHeader}`)
  }

  try {
    const response = await fetch(targetUrl, {
      method: req.method,
      headers,
    })
    return response
  } catch (error) {
    console.error('Security logs proxy error:', error)
    return NextResponse.json({ message: 'Proxy Error' }, { status: 502 })
  }
}

export async function GET(req: NextRequest) {
  return handler(req)
}
