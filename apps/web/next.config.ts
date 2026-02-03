import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  reactStrictMode: true,
  transpilePackages: ['@xenios/api-client', '@xenios/shared-types', '@xenios/ui-kit'],
}

export default nextConfig
