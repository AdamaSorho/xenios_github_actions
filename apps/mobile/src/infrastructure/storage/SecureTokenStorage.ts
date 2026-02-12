import { AuthTokens } from '@/domain/entities/AuthTokens'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import * as Keychain from 'react-native-keychain'

const KEYCHAIN_SERVICE = 'com.xenios.auth'

function isValidAuthTokens(value: unknown): value is AuthTokens {
  return (
    typeof value === 'object' &&
    value !== null &&
    typeof (value as AuthTokens).accessToken === 'string' &&
    (value as AuthTokens).accessToken.length > 0 &&
    typeof (value as AuthTokens).refreshToken === 'string' &&
    (value as AuthTokens).refreshToken.length > 0
  )
}

export class SecureTokenStorage implements TokenStorageRepository {
  async saveTokens(tokens: AuthTokens): Promise<void> {
    const data = JSON.stringify(tokens)
    await Keychain.setGenericPassword('auth_tokens', data, {
      service: KEYCHAIN_SERVICE,
    })
  }

  async getTokens(): Promise<AuthTokens | null> {
    const credentials = await Keychain.getGenericPassword({
      service: KEYCHAIN_SERVICE,
    })

    if (!credentials) {
      return null
    }

    try {
      const parsed: unknown = JSON.parse(credentials.password)
      if (!isValidAuthTokens(parsed)) {
        return null
      }
      return parsed
    } catch {
      return null
    }
  }

  async clearTokens(): Promise<void> {
    await Keychain.resetGenericPassword({ service: KEYCHAIN_SERVICE })
  }
}
