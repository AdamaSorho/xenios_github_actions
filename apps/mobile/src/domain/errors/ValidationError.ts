export class ValidationError extends Error {
  public readonly errors: string[]
  constructor(errors: string[]) {
    super(errors.join(', '))
    this.name = 'ValidationError'
    this.errors = errors
  }
}
