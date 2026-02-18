
import { NextResponse } from 'next/server'

const API_URL = process.env.NEXT_PUBLIC_API_URL

// Helper function to get auth headers
function getAuthHeaders(request: Request) {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  const authToken = request.headers.get('x-auth-token')
  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`
  }
  return headers
}

// GET /api/tenants - List tenants
export async function GET(request: Request) {
  const { searchParams } = new URL(request.url)
  const keyword = searchParams.get('keyword') || ''
  const status = searchParams.get('status') || ''
  const page = searchParams.get('page') || '1'
  const size = searchParams.get('size') || '10'

  try {
    const queryParams = new URLSearchParams()
    if (keyword) queryParams.set('keyword', keyword)
    if (status) queryParams.set('status', status)
    queryParams.set('page', page)
    queryParams.set('size', size)

    const apiRes = await fetch(`${API_URL}/tenants?${queryParams.toString()}`, {
      method: 'GET',
      headers: getAuthHeaders(request),
      credentials: 'include',
    })

    const data = await apiRes.json()

    if (!apiRes.ok) {
      return NextResponse.json({ message: data.message || 'Failed to fetch tenants' }, { status: apiRes.status })
    }

    return NextResponse.json(data)
  } catch (error) {
    console.error('Tenants list error:', error)
    return NextResponse.json({ message: 'An unexpected error occurred' }, { status: 500 })
  }
}

// POST /api/tenants - Create tenant
export async function POST(request: Request) {
  const body = await request.json()

  try {
    const apiRes = await fetch(`${API_URL}/tenants`, {
      method: 'POST',
      headers: getAuthHeaders(request),
      body: JSON.stringify(body),
      credentials: 'include',
    })

    const data = await apiRes.json()

    if (!apiRes.ok) {
      return NextResponse.json({ message: data.message || 'Failed to create tenant' }, { status: apiRes.status })
    }

    return NextResponse.json(data)
  } catch (error) {
    console.error('Create tenant error:', error)
    return NextResponse.json({ message: 'An unexpected error occurred' }, { status: 500 })
  }
}
