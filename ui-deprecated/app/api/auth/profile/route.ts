import { NextResponse } from 'next/server'

const API_URL = process.env.NEXT_PUBLIC_API_URL

export async function GET(request: Request) {
  try {
    // Get token from x-auth-token header (sent by API client)
    const headers = new Headers()
    const authHeader = request.headers.get('x-auth-token')
    if (authHeader) {
      console.log('Profile GET: Received token:', authHeader)
      headers.set('Authorization', `Bearer ${authHeader}`)
    } else {
      console.log('Profile GET: No x-auth-token found')
    }

    // Forward the request to backend
    const apiRes = await fetch(`${API_URL}/auth/profile`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...Object.fromEntries(headers.entries()),
      },
    })

    const data = await apiRes.json()

    if (!apiRes.ok) {
      return NextResponse.json({ message: data.message || 'Failed to get profile' }, { status: apiRes.status })
    }

    return NextResponse.json(data)
  } catch (error) {
    console.error('Get profile proxy error:', error)
    return NextResponse.json({ message: 'An unexpected error occurred' }, { status: 500 })
  }
}

export async function PATCH(request: Request) {
  const body = await request.json()

  try {
    // Debug: Log all headers
    const allHeaders: Record<string, string> = {}
    request.headers.forEach((value, key) => {
      allHeaders[key] = value
    })
    console.log('Profile PATCH: All received headers:', allHeaders)

    // Get token from x-auth-token header (sent by API client)
    const headers = new Headers()
    const authHeader = request.headers.get('x-auth-token')
    console.log('Profile PATCH: x-auth-token header value:', authHeader)

    if (authHeader) {
      console.log('Profile PATCH: Received token:', authHeader)
      headers.set('Authorization', `Bearer ${authHeader}`)
    } else {
      console.log('Profile PATCH: No x-auth-token found')
    }

    // Forward the update profile request to backend
    const apiRes = await fetch(`${API_URL}/auth/profile`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        ...Object.fromEntries(headers.entries()),
      },
      body: JSON.stringify(body),
    })

    const data = await apiRes.json()

    if (!apiRes.ok) {
      return NextResponse.json({ message: data.message || 'Update failed' }, { status: apiRes.status })
    }

    return NextResponse.json(data)
  } catch (error) {
    console.error('Update profile proxy error:', error)
    return NextResponse.json({ message: 'An unexpected error occurred' }, { status: 500 })
  }
}
