/**
 * API Client for Xenios Backend
 *
 * Web and Mobile apps use this client to communicate with the Backend API.
 * This is the ONLY way frontend apps should access data - never directly to the database.
 */

export interface ApiResponse<T> {
  data: T | null
  error: string | null
  ok: boolean
}

export interface ApiClientConfig {
  baseUrl: string
  headers?: Record<string, string>
}

class ApiClient {
  private baseUrl: string
  private headers: Record<string, string>

  constructor() {
    // Default configuration - can be overridden via configure()
    this.baseUrl = this.getDefaultBaseUrl()
    this.headers = {
      'Content-Type': 'application/json',
    }
  }

  private getDefaultBaseUrl(): string {
    // Check for environment variable (works in both web and mobile)
    if (typeof process !== 'undefined' && process.env) {
      const envUrl =
        process.env.NEXT_PUBLIC_API_URL ||
        process.env.EXPO_PUBLIC_API_URL ||
        process.env.API_URL
      if (envUrl) return envUrl
    }

    // Default to localhost for development
    return 'http://localhost:8080/api'
  }

  configure(config: Partial<ApiClientConfig>): void {
    if (config.baseUrl) {
      this.baseUrl = config.baseUrl
    }
    if (config.headers) {
      this.headers = { ...this.headers, ...config.headers }
    }
  }

  setAuthToken(token: string): void {
    this.headers['Authorization'] = `Bearer ${token}`
  }

  clearAuthToken(): void {
    delete this.headers['Authorization']
  }

  async get<T>(path: string): Promise<ApiResponse<T>> {
    return this.request<T>('GET', path)
  }

  async post<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>('POST', path, body)
  }

  async put<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>('PUT', path, body)
  }

  async patch<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>('PATCH', path, body)
  }

  async delete<T = void>(path: string): Promise<ApiResponse<T>> {
    return this.request<T>('DELETE', path)
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseUrl}${path}`

    try {
      const response = await fetch(url, {
        method,
        headers: this.headers,
        body: body ? JSON.stringify(body) : undefined,
      })

      // Handle non-JSON responses
      const contentType = response.headers.get('content-type')
      if (!contentType || !contentType.includes('application/json')) {
        if (!response.ok) {
          return {
            data: null,
            error: response.statusText || 'Request failed',
            ok: false,
          }
        }
        return {
          data: null,
          error: null,
          ok: true,
        }
      }

      const data = await response.json()

      if (!response.ok) {
        return {
          data: null,
          error: data.error || data.message || response.statusText,
          ok: false,
        }
      }

      return {
        data: data as T,
        error: null,
        ok: true,
      }
    } catch (error) {
      return {
        data: null,
        error: error instanceof Error ? error.message : 'Network error',
        ok: false,
      }
    }
  }
}

// Singleton instance
export const apiClient = new ApiClient()
