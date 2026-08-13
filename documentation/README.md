# Documentation Website

This app renders the Markdown files in `../docs` in two separate areas:

- **Team Docs** for sales, product/testing, and operations
- **Developer Docs** for setup, architecture, APIs, and implementation

Developers can switch between both areas. Team users can bookmark their role page without seeing the developer navigation.

```bash
npm install
npm run dev
```

Open `http://localhost:4174`. Markdown, Compose, API-route, and default-module changes hot reload during
development. Use `npm run build` for a production snapshot, or start the Compose service at
`http://docs.localhost:8090`.

## Direct team links

- Team home: `http://localhost:4174/#/team/overview`
- Sales: `http://localhost:4174/#/team/sales`
- Product and testing: `http://localhost:4174/#/team/product-testing`
- Operations: `http://localhost:4174/#/team/operations`
- Developer docs: `http://localhost:4174/#/01_introduction_saas_overview`
