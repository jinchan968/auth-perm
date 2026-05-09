'use client';

import React, { useState, useEffect, useRef } from 'react';
import { Spin } from 'antd';
import { useRouter } from 'next/navigation';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';
import { api, SearchResults } from '@/lib/api';

interface SearchModalProps {
  open: boolean;
  onClose: () => void;
}

export default function SearchModal({ open, onClose }: SearchModalProps) {
  const { lang } = useThemeContext();
  const router = useRouter();
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResults | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      setQuery('');
      setResults(null);
      setLoading(false);
      setError(false);
      setTimeout(() => inputRef.current?.focus(), 100);
    }
  }, [open]);

  useEffect(() => {
    if (!query.trim()) {
      setResults(null);
      setError(false);
      return;
    }
    setLoading(true);
    setError(false);
    const timer = setTimeout(async () => {
      try {
        const data = await api.search(query);
        setResults(data);
      } catch {
        setError(true);
      }
      setLoading(false);
    }, 200);
    return () => clearTimeout(timer);
  }, [query]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (open) {
      window.addEventListener('keydown', handler);
      return () => window.removeEventListener('keydown', handler);
    }
  }, [open, onClose]);

  if (!open) return null;

  const hasResults = results && (results.themes.length > 0 || results.tickers.length > 0 || results.events.length > 0);

  return (
    <div
      style={{
        position: 'fixed', inset: 0, zIndex: 1000,
        background: 'rgba(0,0,0,0.5)', backdropFilter: 'blur(4px)',
        display: 'flex', justifyContent: 'center', paddingTop: '10vh',
        paddingLeft: 16, paddingRight: 16,
      }}
      onClick={onClose}
    >
      <div
        style={{
          width: '100%', maxWidth: 520, maxHeight: '70vh',
          background: 'var(--nshock-bg-card)', border: '1px solid var(--nshock-border)',
          borderRadius: 12, overflow: 'hidden', display: 'flex', flexDirection: 'column',
          boxShadow: '0 20px 60px rgba(0,0,0,0.3)',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--nshock-border)' }}>
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={tt('searchGlobal', lang)}
            style={{
              width: '100%', border: 'none', outline: 'none', fontSize: 16,
              background: 'transparent', color: 'var(--nshock-text)', fontFamily: 'inherit',
            }}
          />
        </div>
        <div style={{ flex: 1, overflow: 'auto', padding: 8 }}>
          {!query.trim() && !results && (
            <div style={{ padding: 12, color: 'var(--nshock-text-muted)', fontSize: 13 }}>
              {tt('typeToSearch', lang)}
            </div>
          )}
          {loading && (
            <div style={{ padding: 12, textAlign: 'center' }}><Spin size="small" /></div>
          )}
          {error && (
            <div style={{ padding: 12, color: 'var(--nshock-danger)', fontSize: 13 }}>
              {tt('searchError', lang)}
            </div>
          )}
          {results && !loading && !hasResults && (
            <div style={{ padding: 12, color: 'var(--nshock-text-muted)', fontSize: 13 }}>
              {tt('noResults', lang)}
            </div>
          )}
          {results?.themes && results.themes.length > 0 && (
            <div style={{ marginBottom: 12 }}>
              <div style={{ fontSize: 11, fontWeight: 700, textTransform: 'uppercase', color: 'var(--nshock-text-muted)', padding: '0 12px', marginBottom: 4 }}>
                {tt('themes', lang)}
              </div>
              {results.themes.slice(0, 5).map((t) => (
                <div key={t.id} className="search-result-item" onClick={() => { onClose(); router.push(`/themes/${t.id}`); }}>
                  <span style={{ fontWeight: 600, fontSize: 14 }}>{t.name}</span>
                  <span style={{ marginLeft: 'auto', fontWeight: 700, fontSize: 13, color: 'var(--nshock-primary)' }}>{t.strength?.toFixed(0)}</span>
                </div>
              ))}
            </div>
          )}
          {results?.tickers && results.tickers.length > 0 && (
            <div style={{ marginBottom: 12 }}>
              <div style={{ fontSize: 11, fontWeight: 700, textTransform: 'uppercase', color: 'var(--nshock-text-muted)', padding: '0 12px', marginBottom: 4 }}>
                {tt('tickers', lang)}
              </div>
              {results.tickers.slice(0, 5).map((t) => (
                <div key={t.id} className="search-result-item" onClick={() => { onClose(); router.push(`/tickers/${t.symbol}`); }}>
                  <span style={{ fontWeight: 700, fontSize: 14 }}>{t.symbol}</span>
                  <span style={{ fontSize: 13, color: 'var(--nshock-text-muted)' }}>{t.name}</span>
                  <span style={{ marginLeft: 'auto', fontWeight: 700, fontSize: 13, color: 'var(--nshock-primary)' }}>{t.hot_score?.toFixed(0)}</span>
                </div>
              ))}
            </div>
          )}
          {results?.events && results.events.length > 0 && (
            <div>
              <div style={{ fontSize: 11, fontWeight: 700, textTransform: 'uppercase', color: 'var(--nshock-text-muted)', padding: '0 12px', marginBottom: 4 }}>
                {tt('events', lang)}
              </div>
              {results.events.slice(0, 5).map((e) => (
                <div key={e.id} className="search-result-item" onClick={() => { onClose(); router.push(`/events/${e.id}`); }}>
                  <span style={{ fontSize: 13 }}>{e.title}</span>
                  <span style={{ marginLeft: 'auto', fontSize: 11, color: 'var(--nshock-text-muted)' }}>{e.created_at?.slice(0, 16)}</span>
                </div>
              ))}
            </div>
          )}
        </div>
        <div style={{ padding: '8px 16px', borderTop: '1px solid var(--nshock-border)', fontSize: 11, color: 'var(--nshock-text-muted)' }}>
          {tt('escToClose', lang)}
        </div>
      </div>
    </div>
  );
}
