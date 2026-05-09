'use client';

import React, { useState } from 'react';
import { Typography, Card, Row, Col, Spin, Select, Input, Pagination, Button } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { api, Theme } from '@/lib/api';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

const CATEGORIES = ['geopolitical', 'ai_semi', 'macro_monetary', 'supply_chain', 'defense', 'energy', 'earnings_event', 'exploratory', 'pharma', 'regulatory'];

export default function ThemesPage() {
  const { lang } = useThemeContext();
  const router = useRouter();
  const [category, setCategory] = useState('');
  const [keyword, setKeyword] = useState('');
  const [page, setPage] = useState(1);
  const pageSize = 24;

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['themes', category, keyword, page],
    queryFn: () => api.getThemes({ ...(category ? { category } : {}), ...(keyword ? { keyword } : {}), page: String(page), page_size: String(pageSize) }),
  });

  const themes: Theme[] = data?.items ?? [];
  const total = data?.total ?? 0;

  return (
    <>
      <div style={{ display: 'flex', gap: 12, marginBottom: 20, flexWrap: 'wrap' }}>
        <Select placeholder={tt('filterMarket', lang)} allowClear style={{ minWidth: 160 }} value={category || undefined}
          onChange={(v) => { setCategory(v ?? ''); setPage(1); }} options={CATEGORIES.map((c) => ({ label: tt(c, lang), value: c }))} />
        <Input.Search placeholder={tt('searchThemes', lang)} allowClear style={{ maxWidth: 300 }} onSearch={(v) => { setKeyword(v); setPage(1); }} />
      </div>

      {isError ? <div style={{ textAlign: 'center', padding: 60 }}><Typography.Text type="danger">{tt('loadError', lang)}: {error?.message}</Typography.Text><br /><Button style={{ marginTop: 12 }} onClick={() => refetch()}>{tt('retry', lang)}</Button></div>
      : isLoading ? <div style={{ display: 'flex', justifyContent: 'center', padding: 60 }}><Spin size="large" /></div>
      : themes.length === 0 ? <div style={{ textAlign: 'center', padding: 60, color: 'var(--nshock-text-muted)' }}>{tt('noThemes', lang)}</div>
      : <>
        <Row gutter={[16, 16]}>
          {themes.map((theme) => (
            <Col xs={24} sm={12} lg={8} xl={6} key={theme.id}>
              <Card hoverable size="small" onClick={() => router.push(`/themes/${theme.id}`)} style={{ height: '100%' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 8 }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 600, fontSize: 14, marginBottom: 4 }}>{theme.name}</div>
                    <span className="ticker-theme-tag">{tt(theme.category, lang)}</span>
                  </div>
                  <div style={{ textAlign: 'right', marginLeft: 8 }}>
                    <div className="strength-bar" style={{ width: 80 }}><div className="fill" style={{ width: `${Math.min(theme.strength_norm, 100)}%` }} /></div>
                    <div style={{ fontSize: 11, color: 'var(--nshock-text-muted)', marginTop: 2 }}>{theme.strength.toFixed(0)}</div>
                  </div>
                </div>
                {theme.description && <div style={{ fontSize: 12, color: 'var(--nshock-text-muted)', lineHeight: 1.5, marginBottom: 8 }}>{theme.description.slice(0, 100)}</div>}
                <div style={{ fontSize: 11, color: 'var(--nshock-text-muted)' }}>
                  {theme.ticker_count}{tt('tickers_label', lang)} · {theme.event_count}{tt('events_label', lang)}
                  <span className={`regime-badge ${theme.trend}`} style={{ marginLeft: 8, fontSize: 10 }}>{tt(theme.trend, lang)}</span>
                </div>
              </Card>
            </Col>
          ))}
        </Row>
        {total > pageSize && <div style={{ display: 'flex', justifyContent: 'center', marginTop: 24 }}><Pagination current={page} pageSize={pageSize} total={total} onChange={setPage} showSizeChanger={false} showTotal={(t) => `${t} ${tt('themes', lang)}`} /></div>}
      </>}
    </>
  );
}
