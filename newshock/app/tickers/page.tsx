'use client';

import React, { useState } from 'react';
import { Typography, Card, Row, Col, Spin, Select, Input, Pagination, Button } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { api, Ticker, displaySymbol } from '@/lib/api';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

const MARKETS = ['us', 'cn', 'hk', 'kr'];

export default function TickersPage() {
  const { lang } = useThemeContext();
  const router = useRouter();
  const [market, setMarket] = useState('');
  const [keyword, setKeyword] = useState('');
  const [page, setPage] = useState(1);
  const pageSize = 24;

  const { data, isPending, isError, error, refetch } = useQuery({
    queryKey: ['tickers', market, keyword, page],
    queryFn: () => api.getTickers({ ...(market ? { market } : {}), ...(keyword ? { keyword } : {}), page: String(page), page_size: String(pageSize) }),
  });

  const tickers: Ticker[] = data?.items ?? [];
  const total = data?.total ?? 0;

  return (
    <>
      <div style={{ display: 'flex', gap: 12, marginBottom: 20, flexWrap: 'wrap' }}>
        <Select placeholder={tt('filterMarket', lang)} allowClear style={{ minWidth: 140 }} value={market || undefined}
          onChange={(v) => { setMarket(v ?? ''); setPage(1); }}
          options={[{ label: tt('allMarkets', lang), value: '' }, ...MARKETS.map((m) => ({ label: tt(`market_${m}`, lang), value: m }))]} />
        <Input.Search placeholder={tt('searchPlaceholder', lang)} allowClear style={{ maxWidth: 300 }} onSearch={(v) => { setKeyword(v); setPage(1); }} />
      </div>

      {isError ? <div style={{ textAlign: 'center', padding: 60 }}><Typography.Text type="danger">{tt('loadError', lang)}: {error?.message}</Typography.Text><br /><Button style={{ marginTop: 12 }} onClick={() => refetch()}>{tt('retry', lang)}</Button></div>
      : isPending && !tickers.length ? <div style={{ display: 'flex', justifyContent: 'center', padding: 60 }}><Spin size="large" /></div>
      : tickers.length === 0 ? <div style={{ textAlign: 'center', padding: 60, color: 'var(--nshock-text-muted)' }}>{tt('noTickers', lang)}</div>
      : <>
        <Row gutter={[16, 16]}>
          {tickers.map((ticker) => (
            <Col xs={24} sm={12} lg={8} xl={6} key={ticker.id}>
              <Card hoverable size="small" onClick={() => router.push(`/tickers/${ticker.symbol}`)} style={{ height: '100%' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <div>
                    <div style={{ fontWeight: 700, fontSize: 16 }}>{displaySymbol(ticker.symbol)}</div>
                    <div style={{ fontSize: 12, color: 'var(--nshock-text-muted)', marginTop: 2 }}>{ticker.name}</div>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <span className="ticker-theme-tag">{ticker.market?.toUpperCase()}</span>
                    <div style={{ fontWeight: 700, fontSize: 18, marginTop: 4 }}>{ticker.hot_score?.toFixed(0)}</div>
                    <div style={{ fontSize: 10, color: 'var(--nshock-text-muted)' }}>{tt('hotScore', lang)}</div>
                  </div>
                </div>
                {ticker.mention_count > 0 && <div style={{ fontSize: 11, color: 'var(--nshock-text-muted)', marginTop: 8 }}>{ticker.mention_count} {tt('mentions', lang)}</div>}
              </Card>
            </Col>
          ))}
        </Row>
        {total > pageSize && <div style={{ display: 'flex', justifyContent: 'center', marginTop: 24 }}><Pagination current={page} pageSize={pageSize} total={total} onChange={setPage} showSizeChanger={false} showTotal={(t) => `${t} ${tt('tickers', lang)}`} /></div>}
      </>}
    </>
  );
}
