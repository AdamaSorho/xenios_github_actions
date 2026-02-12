/** @type {import('jest').Config} */
const config = {
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@xenios/api-client$': '<rootDir>/../../packages/api-client/src',
    '^@xenios/shared-types$': '<rootDir>/../../packages/shared-types/src',
  },
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', { tsconfig: { jsx: 'react-jsx', esModuleInterop: true, module: 'commonjs', moduleResolution: 'node', paths: { '@/*': ['./src/*'] }, baseUrl: '.' } }],
  },
  testMatch: ['**/__tests__/**/*.test.ts', '**/__tests__/**/*.test.tsx'],
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/app/layout.tsx',
    '!src/app/page.tsx',
    '!src/app/providers.tsx',
    '!src/app/**/page.tsx',
    '!src/app/**/layout.tsx',
    '!src/infrastructure/container.ts',
    '!src/presentation/hooks/useUser.ts',
    '!src/infrastructure/repositories/ApiUserRepository.ts',
  ],
  coverageReporters: ['text', 'lcov', 'json-summary'],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80,
    },
  },
}

module.exports = config
