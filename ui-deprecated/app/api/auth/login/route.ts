
import { NextResponse } from 'next/server'
import { serialize } from 'cookie'

const API_URL = process.env.NEXT_PUBLIC_API_URL

export async function POST(request: Request) {
  const body = await request.json()

  // Get client info from headers
  const userAgent = request.headers.get('user-agent') || ''
  const forwardedFor = request.headers.get('x-forwarded-for') || ''
  const clientIP = forwardedFor.split(',')[0].trim() || 'unknown'

  // Ensure IPAddress and UserAgent are in the request body
  if (!body.ip_address && clientIP) {
    body.ip_address = clientIP
  }
  if (!body.user_agent && userAgent) {
    body.user_agent = userAgent
  }

  try {
    // 1. Forward the login request to the backend API
    const apiRes = await fetch(`${API_URL}/auth/public/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Forwarded-For': clientIP,
        'User-Agent': userAgent,
      },
      body: JSON.stringify(body),
    })

    const data = await apiRes.json()

    // 2. Check if the backend login was successful
    if (!apiRes.ok) {
      return NextResponse.json({ message: data.message || 'Authentication failed' }, { status: apiRes.status })
    }

    console.log('Login route: Backend response data:', data)
    console.log('Login route: Backend response structure:', JSON.stringify(data, null, 2))

    // 后端返回格式：{ code, msg, data }
    // 需要从 data 字段中提取实际的用户数据
    if (!data.data) {
      console.error('Login route: No data field in backend response!')
      return NextResponse.json({ message: '服务器返回格式错误' }, { status: 500 })
    }

    const backendData = data.data
    const { token, expires_at, ...userData } = backendData

    console.log('Login route: Extracted token:', token)
    console.log('Login route: Expires at:', expires_at)
    console.log('Login route: Response payload:', { ...userData, expires_at, token })

    // 3. Clear any existing auth cookies (leftover from previous tests)
    const clearCookie = serialize(process.env.AUTH_COOKIE_NAME || 'auth_token', '', {
      httpOnly: true,
      secure: false,
      expires: new Date(0),
      path: '/',
      sameSite: 'lax',
    })

    // 4. Return the user data to the client, with the token
    // Wrap in { data: ... } for frontend consistency
    const responseData = { data: { ...userData, expires_at, token } }
    const response = NextResponse.json(responseData)
    response.headers.set('Set-Cookie', clearCookie)

    return response

  } catch (error) {
    console.error('Login proxy error:', error)
    return NextResponse.json({ message: 'An unexpected error occurred' }, { status: 500 })
  }
}
