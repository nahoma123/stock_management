import { useEffect, useMemo, useState } from 'react';
import DOMPurify from 'dompurify';
import { marked } from 'marked';
import {
  BookOpen,
  Boxes,
  Code2,
  ChevronRight,
  FileText,
  Menu,
  Search,
  Users,
  X,
} from 'lucide-react';
import { systemDocument } from './systemDocument';

const markdownFiles = import.meta.glob(['../../docs/**/*.md', '!../../docs/archive/**'], {
  eager: true,
  query: '?raw',
  import: 'default',
});

const titleFromMarkdown = (content, fallback) => {
  const match = content.match(/^#\s+(.+)$/m);
  return match ? match[1].replace(/^Section \d+:\s*/, '') : fallback;
};

const slugFromPath = (path) => path
  .replace('../../docs/', '')
  .replace(/\.md$/, '')
  .replace(/\/README$/i, '/index')
  .toLowerCase();

const labelFromPath = (path) => path
  .split('/').pop()
  .replace(/\.md$/, '')
  .replace(/^\d+[_-]?/, '')
  .replace(/[_-]+/g, ' ')
  .replace(/\b\w/g, (letter) => letter.toUpperCase());

const sectionFromPath = (path) => {
  const relative = path.replace('../../docs/', '');
  if (relative.startsWith('archive/')) return 'Archive';
  if (relative === 'team/overview.md') return 'Start Here';
  if (relative.startsWith('team/')) return 'Role Guides';
  if (relative.startsWith('architecture/')) return 'System Design';
  if (/mobile/i.test(relative)) return 'API Reference';
  return 'Engineering Guides';
};

const documents = [...Object.entries(markdownFiles)
  .map(([path, content]) => ({
    path,
    content,
    slug: slugFromPath(path),
    title: titleFromMarkdown(content, labelFromPath(path)),
    section: sectionFromPath(path),
    area: path.includes('/team/') ? 'team' : 'developer',
  })), systemDocument]
  .sort((a, b) => a.path.localeCompare(b.path));

const visibleDocuments = documents.filter((document) => document.section !== 'Archive');
const teamHome = visibleDocuments.find((document) => document.slug === 'team/overview');
const developerHome = visibleDocuments.find((document) => document.slug.includes('01_introduction'));
const defaultDocument = teamHome || visibleDocuments[0];
const navigationSections = {
  team: ['Start Here', 'Role Guides'],
  developer: ['Engineering Guides', 'System Design', 'API Reference'],
};

marked.use({
  gfm: true,
  renderer: {
    heading({ tokens, depth }) {
      const text = this.parser.parseInline(tokens);
      const plainText = text.replace(/<[^>]+>/g, '');
      const id = plainText.toLowerCase().replace(/[^a-z0-9\s-]/g, '').trim().replace(/\s+/g, '-');
      return `<h${depth} id="${id}">${text}</h${depth}>`;
    },
  },
});

const readRoute = () => decodeURIComponent(window.location.hash.replace(/^#\/?/, '').split('?')[0]);

function App() {
  const [route, setRoute] = useState(readRoute());
  const [query, setQuery] = useState('');
  const [menuOpen, setMenuOpen] = useState(false);

  const activeDocument = documents.find((document) => document.slug === route) || defaultDocument;
  const activeArea = activeDocument?.area || 'team';
  const renderedContent = useMemo(
    () => DOMPurify.sanitize(marked.parse(activeDocument?.content || '# Documentation unavailable')),
    [activeDocument],
  );
  const headings = useMemo(() => {
    const matches = [...(activeDocument?.content || '').matchAll(/^(#{2,3})\s+(.+)$/gm)];
    return matches.map((match) => ({
      level: match[1].length,
      label: match[2].replace(/[*_`]/g, ''),
      id: match[2].toLowerCase().replace(/[^a-z0-9\s-]/g, '').trim().replace(/\s+/g, '-'),
    }));
  }, [activeDocument]);
  const filteredDocuments = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    const areaDocuments = visibleDocuments.filter((document) => document.area === activeArea);
    if (!normalizedQuery) return areaDocuments;
    return areaDocuments.filter((document) =>
      `${document.title} ${document.content}`.toLowerCase().includes(normalizedQuery),
    );
  }, [activeArea, query]);

  useEffect(() => {
    const onHashChange = () => {
      setRoute(readRoute());
      window.scrollTo({ top: 0 });
    };
    window.addEventListener('hashchange', onHashChange);
    if (!route && defaultDocument) window.history.replaceState(null, '', `#/${defaultDocument.slug}`);
    return () => window.removeEventListener('hashchange', onHashChange);
  }, [route]);

  const openDocument = (document) => {
    window.location.hash = `/${document.slug}`;
    setMenuOpen(false);
  };

  const openArea = (area) => {
    setQuery('');
    openDocument(area === 'team' ? teamHome : developerHome);
  };

  const handleArticleClick = (event) => {
    const link = event.target.closest('a');
    const href = link?.getAttribute('href');
    if (!href || !href.match(/\.md(?:#.*)?$/i)) return;
    const [filePath] = href.split('#');
    const fileName = filePath.split('/').pop().replace(/\.md$/i, '').toLowerCase();
    const targetDocument = documents.find((document) => document.slug.split('/').pop() === fileName);
    if (!targetDocument) return;
    event.preventDefault();
    openDocument(targetDocument);
  };

  return (
    <div className="docs-shell">
      <header className="topbar">
        <button className="icon-button menu-button" onClick={() => setMenuOpen(true)} aria-label="Open navigation">
          <Menu size={20} />
        </button>
        <a className="brand" href={`#/${defaultDocument?.slug || ''}`}>
          <span className="brand-mark"><Boxes size={20} /></span>
          <span><strong>Stock Platform</strong><small>{activeArea === 'team' ? 'Team documentation' : 'Developer documentation'}</small></span>
        </a>
        <div className="topbar-status"><span /> Source synced</div>
      </header>

      <aside className={`sidebar ${menuOpen ? 'is-open' : ''}`}>
        <div className="sidebar-header">
          <span>Documentation</span>
          <button className="icon-button close-button" onClick={() => setMenuOpen(false)} aria-label="Close navigation"><X size={20} /></button>
        </div>
        <label className="search-box">
          <Search size={17} />
          <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search docs" />
          {query && <button onClick={() => setQuery('')} aria-label="Clear search"><X size={15} /></button>}
        </label>
        <div className="area-switch" aria-label="Documentation area">
          <button className={activeArea === 'team' ? 'active' : ''} onClick={() => openArea('team')}>
            <Users size={16} /> Team Docs
          </button>
          <button className={activeArea === 'developer' ? 'active' : ''} onClick={() => openArea('developer')}>
            <Code2 size={16} /> Developer Docs
          </button>
        </div>
        <nav className="document-nav" aria-label="Documentation pages">
          {navigationSections[activeArea].map((section) => {
            const sectionDocuments = filteredDocuments.filter((document) => document.section === section);
            if (!sectionDocuments.length) return null;
            return <div className="nav-section" key={section}>
              <h2>{section}</h2>
              {sectionDocuments.map((document) => (
                <button
                  className={document.slug === activeDocument?.slug ? 'active' : ''}
                  key={document.slug}
                  onClick={() => openDocument(document)}
                >
                  <FileText size={16} />
                  <span>{document.title}</span>
                  <ChevronRight size={15} />
                </button>
              ))}
            </div>;
          })}
          {query && !filteredDocuments.length && <p className="empty-search">No documentation matches “{query}”.</p>}
        </nav>
        <div className="sidebar-footer"><BookOpen size={16} /> {filteredDocuments.length} pages in this area</div>
      </aside>

      {menuOpen && <button className="mobile-scrim" onClick={() => setMenuOpen(false)} aria-label="Close navigation" />}

      <main className="content-wrap">
        <article className="markdown-body" onClick={handleArticleClick} dangerouslySetInnerHTML={{ __html: renderedContent }} />
      </main>

      <aside className="page-outline">
        <span>On this page</span>
        {headings.length ? headings.map((heading) => (
          <button
            className={heading.level === 3 ? 'nested' : ''}
            onClick={() => document.getElementById(heading.id)?.scrollIntoView()}
            key={`${heading.id}-${heading.label}`}
          >
            {heading.label}
          </button>
        )) : <small>No sections</small>}
      </aside>
    </div>
  );
}

export default App;
