import { AuthTokens } from '@/domain/entities/AuthTokens'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import * as Keychain from 'react-native-keychain'

const KEYCHAIN_SERVICE = 'com.xenios.auth'

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
      return JSON.parse(credentials.password) as AuthTokens
    } catch {
      return null
    }
  }

  async clearTokens(): Promise<void> {
    await Keychain.resetGenericPassword({ service: KEYCHAIN_SERVICE })
  }
}
