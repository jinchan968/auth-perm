import { NextResponse } from 'next/server'

const API_URL = process.env.NEXT_PUBLIC_API_URL

export async function POST(request: Request) {
  const body = await request.json()

  try {
    // 1. Forward the register request to the backend API
    const apiRes = await fetch(`${API_URL}/auth/public/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })

    const data = await apiRes.json()

    // 2. Check if the backend register was successful
    if (!apiRes.ok) {
      // 后端返回格式：{ code, msg, error }
      // error 字段包含具体错误信息（如"用户名已存在"），msg 是通用消息（如"注册失败"）
      // 优先显示具体错误信息
      let errorMessage = data.error || data.msg || 'Registration failed'
      // 如果 error 包含前缀（如 "BUSINESS: 用户名已存在"），去除前缀
      if (errorMessage.includes(': ')) {
        errorMessage = errorMessage.split(': ').slice(1).join(': ')
      }
      return NextResponse.json({ message: errorMessage }, { status: apiRes.status })
    }

    console.log('Register route: Backend response data:', data)
    console.log('Register route: Backend response structure:', JSON.stringify(data, null, 2))

    // 后端返回格式：{ code, msg, data }
    // 需要从 data 字段中提取实际的用户数据
    if (!data.data) {
      console.error('Register route: No data field in backend response!')
      return NextResponse.json({ message: '服务器返回格式错误' }, { status: 500 })
    }

    const backendData = data.data
    const { token, expires_at, ...userData } = backendData

    console.log('Register route: Extracted token:', token)
    console.log('Register route: Expires at:', expires_at)
    console.log('Register route: Response payload:', { ...userData, expires_at, token })

    // 3. Return the user data to the client, with the token
    // Wrap in { data: ... } for frontend consistency
    const responseData = { data: { ...userData, expires_at, token } }
    const response = NextResponse.json(responseData)

    return response

  } catch (error) {
    console.error('Register proxy error:', error)
    return NextResponse.json({ message: 'An unexpected error occurred' }, { status: 500 })
  }
}
