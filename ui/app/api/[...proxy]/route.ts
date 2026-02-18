
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

const TARGET_API_URL = process.env.NEXT_PUBLIC_API_URL

// This route acts as a proxy for all authenticated API calls.
// It reads the auth token from localStorage and adds it as Authorization header.
// This solves the cross-port cookie limitation issue.

async function proxyHandler(req: NextRequest) {
  // Construct the target URL
  // Strip /api prefix (TARGET_API_URL already includes /api/v1)
  const path = req.nextUrl.pathname.replace('/api', '')
  const targetUrl = `${TARGET_API_URL}${path}${req.nextUrl.search}`

  // Clone headers
  const headers = new Headers(req.headers)

  // Remove the host header, it will be replaced by the target host
  headers.delete('host')

  // Add the auth token from localStorage as Authorization header
  // Note: In Next.js API routes, we can't directly access localStorage
  // So we need to pass it via a special header from the client
  const authHeader = req.headers.get('x-auth-token')
  if (authHeader) {
    console.log('Proxy: Sending token:', authHeader)
    headers.set('Authorization', `Bearer ${authHeader}`)
  } else {
    console.log('Proxy: No x-auth-token found in request headers')
  }

  console.log('Proxy: Target URL:', targetUrl)
  console.log('Proxy: Method:', req.method)

  try {
    const response = await fetch(targetUrl, {
      method: req.method,
      headers: headers,
      body: req.body,
      // Pass duplex for streaming request bodies
      // @ts-ignore
      duplex: 'half',
    })

    console.log('Proxy: Response status:', response.status)

    // Stream the response back to the client
    return response

  } catch (error) {
    console.error(`API Proxy error for ${req.method} ${targetUrl}:`, error)
    return NextResponse.json({ message: 'API Proxy Error' }, { status: 502 }) // Bad Gateway
  }
}

export async function GET(req: NextRequest) {
  return proxyHandler(req)
}

export async function POST(req: NextRequest) {
  return proxyHandler(req)
}

export async function PUT(req: NextRequest) {
  return proxyHandler(req)
}

export async function DELETE(req: NextRequest) {
  return proxyHandler(req)
}

export async function PATCH(req: NextRequest) {
  return proxyHandler(req)
}

export async function OPTIONS(req: NextRequest) {
  const response = new NextResponse(null, { status: 204 })
  response.headers.set('Access-Control-Allow-Origin', '*')
  response.headers.set('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, PATCH, OPTIONS')
  response.headers.set('Access-Control-Allow-Headers', 'Content-Type, Authorization, x-auth-token')
  return response
}
