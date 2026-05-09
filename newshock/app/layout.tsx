'use client';

import React, { useState, useEffect } from 'react';
import { ConfigProvider, theme as antdTheme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useRouter, usePathname } from 'next/navigation';
import { ThemeProvider, useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';
import Sidebar from '@/components/layout/Sidebar';
import Topbar from '@/components/layout/Topbar';
import SearchModal from '@/components/layout/SearchModal';
import './globals.css';

const darkToken = {
  colorPrimary: '#6366f1',
  colorBgContainer: '#14141f',
  colorBgElevated: '#1a1a2e',
  colorBorderSecondary: 'rgba(255,255,255,0.06)',
  colorText: '#e4e4ed',
  colorTextSecondary: '#8888a0',
  borderRadius: 8,
  fontFamily: "-apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Segoe UI', sans-serif",
};

const lightToken = {
  colorPrimary: '#4f46e5',
  borderRadius: 8,
  fontFamily: "-apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Segoe UI', sans-serif",
};

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { staleTime: 60_000, retry: 2 },
    },
  });
}

let browserQueryClient: QueryClient | undefined;
function getQueryClient() {
  if (typeof window === 'undefined') return makeQueryClient();
  if (!browserQueryClient) browserQueryClient = makeQueryClient();
  return browserQueryClient;
}

function AntdProvider({ children }: { children: React.ReactNode }) {
  const { theme } = useThemeContext();
  const isDark = theme === 'dark';

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
        token: isDark ? darkToken : lightToken,
      }}
    >
      {children}
    </ConfigProvider>
  );
}

function AppShell({ children }: { children: React.ReactNode }) {
  const { lang } = useThemeContext();
  const router = useRouter();
  const pathname = usePathname();
  const [searchOpen, setSearchOpen] = useState(false);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setSearchOpen((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, []);

  return (
    <>
      <Sidebar />
      <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}>
        <Topbar onOpenSearch={() => setSearchOpen(true)} />
        <main className="main-content">
          {children}
        </main>
      </div>
      {/* Mobile Topbar */}
      <div className="topbar-mobile">
        <span className="topbar-mobile-brand">News<em>shock</em></span>
        <div className="topbar-mobile-actions">
          <button onClick={() => setSearchOpen(true)} style={{ width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center', borderRadius: 8, border: 'none', background: 'transparent', color: 'var(--nshock-text-muted)', cursor: 'pointer' }}>
            🔍
          </button>
        </div>
      </div>
      {/* Mobile Tabbar */}
      <nav className="mobile-tabbar">
        {[
          { href: '/', key: 'radar', icon: '⊙' },
          { href: '/themes', key: 'themes', icon: '◆' },
          { href: '/tickers', key: 'tickers', icon: '▲' },
          { href: '/markets', key: 'markets', icon: '≡' },
          { href: '/edge', key: 'edge', icon: '⚡' },
          { href: '/events', key: 'events', icon: '◷' },
        ].map((item) => {
          const active = item.href === '/' ? pathname === '/' : pathname.startsWith(item.href);
          return (
            <a key={item.href} onClick={() => router.push(item.href)} style={{ color: active ? 'var(--nshock-primary)' : undefined }}>
              <span>{item.icon}</span>
              <span>{tt(item.key, lang)}</span>
            </a>
          );
        })}
      </nav>
      <SearchModal open={searchOpen} onClose={() => setSearchOpen(false)} />
    </>
  );
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);

  return (
    <html lang="zh-CN" data-theme="dark">
      <head>
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>Newshock — 事件驱动主题投研雷达</title>
        <meta name="description" content="实时追踪市场主线、催化事件、标的暴露 — A股 / 美股 / 港股事件驱动主题投研工具" />
        <script
          dangerouslySetInnerHTML={{
            __html: `(function(){try{var m=localStorage.getItem('theme-mode');if(m==='dark'||m==='light')document.documentElement.dataset.theme=m;}catch(e){}})()`,
          }}
        />
      </head>
      <body className="antialiased">
        {mounted && (
          <QueryClientProvider client={getQueryClient()}>
            <ThemeProvider>
              <AntdProvider>
                <div className="app-shell">
                  <AppShell>{children}</AppShell>
                </div>
              </AntdProvider>
            </ThemeProvider>
          </QueryClientProvider>
        )}
      </body>
    </html>
  );
}
