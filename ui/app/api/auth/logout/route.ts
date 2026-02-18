
import { NextResponse } from 'next/server'
import { serialize } from 'cookie'

export async function POST() {
  const cookieName = process.env.AUTH_COOKIE_NAME || 'auth_token'

  // Clear the authentication cookie by setting its maxAge to -1
  // Note: sameSite must match the login cookie setting
  const cookie = serialize(cookieName, '', {
    httpOnly: true,
    secure: false, // Set to true in production with HTTPS
    expires: new Date(0), // Set to a past date
    path: '/',
    sameSite: 'none', // Must match login cookie's sameSite setting
  });

  const response = NextResponse.json({ message: 'Logged out successfully' })
  response.headers.set('Set-Cookie', cookie)

  return response
}
