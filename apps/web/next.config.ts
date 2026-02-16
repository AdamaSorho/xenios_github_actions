import type { NextConfig } from 'next'

const apiUrl = process.env.NEXT_PUBLIC_API_URL
if (!apiUrl && process.env.NODE_ENV === 'production') {
  console.warn(
    'WARNING: NEXT_PUBLIC_API_URL is not set. CSP connect-src will default to http://localhost:8080. ' +
    'Set NEXT_PUBLIC_API_URL in production to restrict connections to the correct API origin.'
  )
}
// CSP connect-src needs origin only (no path), e.g. https://example.com
let connectSrc = 'http://localhost:8080'
if (apiUrl) {
  try {
    connectSrc = new URL(apiUrl).origin
  } catch {
    connectSrc = apiUrl
  }
}

const securityHeaders = [
  {
    key: 'X-Content-Type-Options',
    value: 'nosniff',
  },
  {
    key: 'X-Frame-Options',
    value: 'DENY',
  },
  {
    key: 'X-XSS-Protection',
    value: '1; mode=block',
  },
  {
    key: 'Referrer-Policy',
    value: 'strict-origin-when-cross-origin',
  },
  {
    key: 'Content-Security-Policy',
    value: `default-src 'self'; script-src 'self' 'unsafe-inline' https://vercel.live; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self' ${connectSrc} https://vercel.live; frame-ancestors 'none';`,
  },
]

const nextConfig: NextConfig = {
  reactStrictMode: true,
  transpilePackages: ['@xenios/api-client', '@xenios/shared-types', '@xenios/ui-kit'],
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: securityHeaders,
      },
    ]
  },
}

export default nextConfig
