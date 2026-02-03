/** @type {import('dependency-cruiser').IConfiguration} */
module.exports = {
  forbidden: [
    // ─── CLEAN ARCHITECTURE RULES ───
    {
      name: 'domain-no-infrastructure',
      severity: 'error',
      comment: 'Domain layer cannot import from infrastructure',
      from: { path: 'domain' },
      to: { path: 'infrastructure' },
    },
    {
      name: 'domain-no-presentation',
      severity: 'error',
      comment: 'Domain layer cannot import from presentation/handlers',
      from: { path: 'domain' },
      to: { path: 'presentation|adapter/handler' },
    },
    {
      name: 'usecase-no-infrastructure',
      severity: 'error',
      comment: 'Use cases cannot import from infrastructure (use interfaces)',
      from: { path: 'usecase|application' },
      to: { path: 'infrastructure' },
    },
    {
      name: 'usecase-no-presentation',
      severity: 'error',
      comment: 'Use cases cannot import from presentation layer',
      from: { path: 'usecase|application' },
      to: { path: 'presentation|adapter/handler' },
    },

    // ─── DATABASE RULES ───
    // Web and Mobile must NEVER access database - only Backend does
    {
      name: 'web-no-supabase',
      severity: 'error',
      comment: 'Web app cannot import Supabase. Use Backend API.',
      from: { path: 'apps/web' },
      to: { path: '@supabase' },
    },
    {
      name: 'web-no-prisma',
      severity: 'error',
      comment: 'Web app cannot import Prisma. Use Backend API.',
      from: { path: 'apps/web' },
      to: { path: '@prisma|prisma' },
    },
    {
      name: 'web-no-typeorm',
      severity: 'error',
      comment: 'Web app cannot import TypeORM. Use Backend API.',
      from: { path: 'apps/web' },
      to: { path: 'typeorm' },
    },
    {
      name: 'web-no-drizzle',
      severity: 'error',
      comment: 'Web app cannot import Drizzle. Use Backend API.',
      from: { path: 'apps/web' },
      to: { path: 'drizzle-orm' },
    },
    {
      name: 'web-no-sequelize',
      severity: 'error',
      comment: 'Web app cannot import Sequelize. Use Backend API.',
      from: { path: 'apps/web' },
      to: { path: 'sequelize' },
    },
    {
      name: 'web-no-pg',
      severity: 'error',
      comment: 'Web app cannot import pg. Use Backend API.',
      from: { path: 'apps/web' },
      to: { path: '^pg$' },
    },
    {
      name: 'mobile-no-supabase',
      severity: 'error',
      comment: 'Mobile app cannot import Supabase. Use Backend API.',
      from: { path: 'apps/mobile' },
      to: { path: '@supabase' },
    },
    {
      name: 'mobile-no-prisma',
      severity: 'error',
      comment: 'Mobile app cannot import Prisma. Use Backend API.',
      from: { path: 'apps/mobile' },
      to: { path: '@prisma|prisma' },
    },
    {
      name: 'mobile-no-typeorm',
      severity: 'error',
      comment: 'Mobile app cannot import TypeORM. Use Backend API.',
      from: { path: 'apps/mobile' },
      to: { path: 'typeorm' },
    },
    {
      name: 'mobile-no-drizzle',
      severity: 'error',
      comment: 'Mobile app cannot import Drizzle. Use Backend API.',
      from: { path: 'apps/mobile' },
      to: { path: 'drizzle-orm' },
    },
    {
      name: 'mobile-no-sequelize',
      severity: 'error',
      comment: 'Mobile app cannot import Sequelize. Use Backend API.',
      from: { path: 'apps/mobile' },
      to: { path: 'sequelize' },
    },

    // ─── NO CIRCULAR DEPENDENCIES ───
    {
      name: 'no-circular',
      severity: 'error',
      comment: 'Circular dependencies are not allowed',
      from: {},
      to: {
        circular: true,
      },
    },
  ],
  options: {
    doNotFollow: {
      path: 'node_modules',
    },
    tsConfig: {
      fileName: './tsconfig.json',
    },
    reporterOptions: {
      dot: {
        collapsePattern: 'node_modules/(@[^/]+/[^/]+|[^/]+)',
      },
    },
  },
}
