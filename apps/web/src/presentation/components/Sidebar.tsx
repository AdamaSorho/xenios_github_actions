'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useAuth } from '@/presentation/hooks/useAuth'

interface NavItem {
  label: string
  href: string
  icon: string
}

const navItems: NavItem[] = [
  { label: 'Dashboard', href: '/dashboard', icon: '📊' },
  { label: 'Insights', href: '/insights', icon: '💡' },
  { label: 'Clients', href: '/clients', icon: '👥' },
  { label: 'Sessions', href: '/sessions', icon: '📋' },
  { label: 'Analytics', href: '/analytics', icon: '📈' },
  { label: 'Settings', href: '/settings', icon: '⚙️' },
]

export function Sidebar() {
  const pathname = usePathname()
  const { user, logout } = useAuth()

  const isActive = (href: string) =>
    pathname === href || pathname.startsWith(href + '/')

  return (
    <aside
      className="flex h-screen w-64 flex-col border-r border-gray-200 bg-white"
      data-testid="sidebar"
    >
      <div className="border-b border-gray-200 p-6">
        <h1 className="text-xl font-bold text-gray-900">Xenios</h1>
      </div>

      <nav className="flex-1 space-y-1 p-4" data-testid="sidebar-nav">
        {navItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={`flex items-center rounded-md px-3 py-2 text-sm font-medium transition-colors ${
              isActive(item.href)
                ? 'bg-blue-50 text-blue-700'
                : 'text-gray-700 hover:bg-gray-100'
            }`}
            data-testid={`nav-${item.label.toLowerCase()}`}
          >
            <span className="mr-3">{item.icon}</span>
            {item.label}
          </Link>
        ))}
      </nav>

      <div className="border-t border-gray-200 p-4">
        {user && (
          <div className="mb-3">
            <p className="text-sm font-medium text-gray-900" data-testid="user-name">
              {user.name}
            </p>
            <p className="text-xs text-gray-500" data-testid="user-email">
              {user.email}
            </p>
          </div>
        )}
        <button
          onClick={logout}
          className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
          data-testid="logout-button"
        >
          Log out
        </button>
      </div>
    </aside>
  )
}
