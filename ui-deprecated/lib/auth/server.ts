
import { cookies } from 'next/headers'
import { User } from '@/lib/api/auth' // Assuming User type is defined here

const API_URL = process.env.NEXT_PUBLIC_API_URL

// This is a server-side only function
export async function getCurrentUser(): Promise<User | null> {
  const cookieStore = cookies()
  const cookieName = process.env.AUTH_COOKIE_NAME || 'auth_token'
  const tokenCookie = cookieStore.get(cookieName)

  if (!tokenCookie) {
    return null
  }

  try {
    const response = await fetch(`${API_URL}/auth/profile`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${tokenCookie.value}`,
      },
      // Use cache: 'no-store' to ensure we always get the latest user data
      cache: 'no-store',
    })

    if (!response.ok) {
      console.error('Failed to fetch current user:', response.statusText)
      return null
    }

    const user = await response.json()

    // Map backend UserResponse to client-side User type
    // Backend returns: UserResponse { id, username, nickname, avatar, phone, identifier_type, identifier_value, status, created_at, updated_at }
    // Client expects: User { id, username, email, name, avatar, roles, profile }
    const userProfile: User = {
        id: user.id,
        username: user.username,
        // Extract email from identifier_value if identifier_type is email, otherwise use empty string
        email: user.identifier_type === 'email' ? user.identifier_value : '',
        name: user.nickname || user.username,
        avatar: user.avatar || '',
        roles: [], // Backend doesn't return roles, need separate call if needed
        profile: {
            phone: user.phone || '',
            identifierType: user.identifier_type,
            identifierValue: user.identifier_value,
            status: user.status
        }
    };


    return userProfile

  } catch (error) {
    console.error('Error fetching current user:', error)
    return null
  }
}
