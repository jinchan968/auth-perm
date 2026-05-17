export const onRequest = async ({ request }: { request: Request }) => {
  const url = new URL(request.url)
  const apiUrl = 'https://api.big-artist.top/api/v1'
  const target = `${apiUrl}${url.pathname.replace('/api', '')}${url.search}`

  const headers = new Headers(request.headers)
  headers.set('X-Forwarded-For', headers.get('CF-Connecting-IP') || '')

  return fetch(target, {
    method: request.method,
    headers,
    body: ['GET', 'HEAD'].includes(request.method) ? null : await request.clone().arrayBuffer(),
  })
}
