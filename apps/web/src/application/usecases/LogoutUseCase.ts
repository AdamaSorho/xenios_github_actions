import { AuthRepository } from '@/domain/repositories/AuthRepository'

export class LogoutUseCase {
  constructor(private readonly authRepo: AuthRepository) {}

  async execute(): Promise<void> {
    return this.authRepo.logout()
  }
}
