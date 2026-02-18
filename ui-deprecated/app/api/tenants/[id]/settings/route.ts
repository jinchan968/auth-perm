
import { NextResponse } from 'next/server'

const API_URL = process.env.NEXT_PUBLIC_API_URL

// GET /api/tenants/[id]/settings - Get tenant settings
export async function GET(
  request: Request,
  { params }: { params: { id: string } }
) {
  const { id } = params

  try {
    const apiRes = await fetch(`${API_URL}/tenants/${id}/settings`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    const data = await apiRes.json()

    if (!apiRes.ok) {
      return NextResponse.json({ message: data.message || 'Failed to fetch tenant settings' }, { status: apiRes.status })
    }

    return NextResponse.json(data)
  } catch (error) {
    console.error('Get tenant settings error:', error)
    return NextResponse.json({ message: 'An unexpected error occurred' }, { status: 500 })
  }
}

// PUT /api/tenants/[id]/settings - Update tenant settings
export async function PUT(
  request: Request,
  { params }: { params: { id: string } }
) {
  const { id } = params
  const body = await request.json()

  try {
    const apiRes = await fetch(`${API_URL}/tenants/${id}/settings`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
      credentials: 'include',
    })

    const data = await apiRes.json()

    if (!apiRes.ok) {
      return NextResponse.json({ message: data.message || 'Failed to update tenant settings' }, { status: apiRes.status })
    }

    return NextResponse.json(data)
  } catch (error) {
    console.error('Update tenant settings error:', error)
    return NextResponse.json({ message: 'An unexpected error occurred' }, { status: 500 })
  }
}
