import { parse } from 'yaml';
import composeSource from '../../docker-compose.yml?raw';
import routesSource from '../../backend/main.go?raw';
import dockerSource from '../../backend/services/docker.go?raw';

const compose = parse(composeSource);

const services = Object.entries(compose.services || {}).map(([name, config]) => ({
  name,
  image: config.image || config.build?.dockerfile || 'Project build',
  purpose: {
    db: 'PostgreSQL control-plane and tenant databases',
    backend: 'Go provisioning, monitoring, and lifecycle API',
    frontend: 'React superadmin application',
    documentation: 'This source-synchronized documentation site',
    traefik: 'Local reverse proxy and hostname routing',
  }[name] || 'Platform service',
}));

const routes = [...routesSource.matchAll(/api\.(GET|POST|PUT|DELETE|PATCH)\("([^"]+)"/g)]
  .map((match) => ({ method: match[1], path: `/api${match[2]}` }));

const moduleMatch = dockerSource.match(/"--init",\s*"([^"]+)"/);
const modules = moduleMatch ? moduleMatch[1].split(',') : [];

const table = (headers, rows) => [
  `| ${headers.join(' | ')} |`,
  `| ${headers.map(() => '---').join(' | ')} |`,
  ...rows.map((row) => `| ${row.join(' | ')} |`),
].join('\n');

export const systemDocument = {
  path: 'generated/system-map.md',
  slug: 'system-map',
  title: 'System Map',
  section: 'Architecture',
  content: `# System Map

> This page is generated from the current Compose and Go sources. It refreshes automatically while the documentation development server is running.

## Platform services

${table(['Service', 'Runtime', 'Responsibility'], services.map((service) => [
  `\`${service.name}\``,
  `\`${service.image}\``,
  service.purpose,
]))}

## Backend API

${table(['Method', 'Route'], routes.map((route) => [`\`${route.method}\``, `\`${route.path}\``]))}

## Default tenant modules

New tenants are initialized with the following modules from the current provisioning command:

${modules.map((module) => `- \`${module}\``).join('\n')}

## Documentation update model

- Edit Markdown in \`docs/\` for maintained guides.
- Change Compose, backend routes, or the provisioning module list and this page updates from source.
- Run \`npm run dev\` in \`documentation/\` for hot reload.
- Run \`npm run build\` or rebuild the documentation container for a production snapshot.
`,
};
