'use client';

import React, { useRef, useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

function matchRoute(current: string, href: string): boolean {
  if (href === '/') return current === '/';
  return current === href || current.startsWith(href + '/');
}

const NAV_ITEMS = [
  { href: '/', key: 'radar', icon: '⊙' },
  { href: '/themes', key: 'themes', icon: '◆' },
  { href: '/tickers', key: 'tickers', icon: '▲' },
  { href: '/markets', key: 'markets', icon: '≡' },
  { href: '/edge', key: 'edge', icon: '⚡' },
  { href: '/events', key: 'events', icon: '◷' },
];

export default function Sidebar() {
  const pathname = usePathname();
  const { lang } = useThemeContext();
  const activeRef = useRef<HTMLAnchorElement>(null);
  const [indicator, setIndicator] = useState({ top: 0, height: 0 });

  useEffect(() => {
    const idx = NAV_ITEMS.findIndex((item) => matchRoute(pathname, item.href));
    const el = document.querySelector(`.sidebar-nav a[data-index="${idx >= 0 ? idx : 0}"]`);
    if (el) {
      const rect = el.getBoundingClientRect();
      const parent = el.parentElement?.getBoundingClientRect();
      if (parent) {
        setIndicator({ top: rect.top - parent.top, height: rect.height });
      }
    }
  }, [pathname]);

  return (
    <aside className="sidebar">
      <Link href="/" className="sidebar-brand">
        <span style={{ fontSize: 22 }}>📡</span>
        <span>News<em>shock</em></span>
      </Link>

      <nav className="sidebar-nav">
        <span
          className="sidebar-nav-indicator"
          style={{ transform: `translateY(${indicator.top}px)`, height: indicator.height }}
        />
        {NAV_ITEMS.map((item, i) => {
          const active = matchRoute(pathname, item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              data-index={i}
              className={active ? 'active' : ''}
              ref={active ? activeRef : undefined}
            >
              <span style={{ fontSize: 14, width: 20, textAlign: 'center' }}>{item.icon}</span>
              <span>{tt(item.key, lang)}</span>
            </Link>
          );
        })}
      </nav>

      <div style={{ marginTop: 'auto', paddingTop: 16, borderTop: '1px solid var(--nshock-border)' }}>
        <div style={{ padding: '9px 12px', fontSize: 11, color: 'var(--nshock-text-muted)' }}>
          Newshock v0.1
        </div>
      </div>
    </aside>
  );
}
