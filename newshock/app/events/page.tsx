'use client';

import React, { useState } from 'react';
import { Typography, Card, Spin, Select, Input, Pagination, Button } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { api, Event } from '@/lib/api';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

export default function EventsPage() {
  const { lang } = useThemeContext();
  const router = useRouter();
  const [channel, setChannel] = useState('');
  const [keyword, setKeyword] = useState('');
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['events', channel, keyword, page],
    queryFn: () => api.getEvents({ ...(channel ? { channel } : {}), ...(keyword ? { keyword } : {}), page: String(page), page_size: String(pageSize) }),
  });

  const events: Event[] = data?.items ?? [];
  const total = data?.total ?? 0;

  return (
    <>
      <div style={{ display: 'flex', gap: 12, marginBottom: 20, flexWrap: 'wrap' }}>
        <Select placeholder={tt('channel', lang)} allowClear style={{ minWidth: 160 }} value={channel || undefined}
          onChange={(v) => { setChannel(v ?? ''); setPage(1); }}
          options={[{ label: tt('global_macro', lang), value: 'global_macro' }, { label: tt('industry_news', lang), value: 'industry_news' }, { label: tt('market_flow', lang), value: 'market_flow' }]} />
        <Input.Search placeholder={tt('searchPlaceholder', lang)} allowClear style={{ maxWidth: 300 }} onSearch={(v) => { setKeyword(v); setPage(1); }} />
      </div>

      {isError ? <div style={{ textAlign: 'center', padding: 60 }}><Typography.Text type="danger">{tt('loadError', lang)}: {error?.message}</Typography.Text><br /><Button style={{ marginTop: 12 }} onClick={() => refetch()}>{tt('retry', lang)}</Button></div>
      : isLoading ? <div style={{ display: 'flex', justifyContent: 'center', padding: 60 }}><Spin size="large" /></div>
      : events.length === 0 ? <div style={{ textAlign: 'center', padding: 60, color: 'var(--nshock-text-muted)' }}>{tt('noEvents', lang)}</div>
      : <>
        <Card size="small">
          {events.map((event) => (
            <div key={event.id} className="event-item" style={{ cursor: 'pointer' }} onClick={() => router.push(`/events/${event.id}`)}>
              <div className="event-meta">
                <span>{event.event_time?.slice(0, 16) || event.created_at?.slice(0, 16)}</span>
                {event.channel && <><span>·</span><span>{event.channel}</span></>}
                {event.theme_name && <><span>·</span><span style={{ color: 'var(--nshock-primary)', fontSize: 10 }}>{event.theme_name}</span></>}
                <div className="importance">{[1,2,3,4,5].map((i) => <div key={i} className={`dot ${i <= event.importance ? 'filled' : ''}`} />)}</div>
              </div>
              <div className="event-title">{event.title}</div>
              {event.summary && <div className="event-summary">{event.summary.slice(0, 200)}</div>}
              {event.tickers?.length > 0 && (
                <div className="event-tickers">
                  {event.tickers.map((t) => <span key={t.symbol} className="ticker-theme-tag" onClick={(e) => { e.stopPropagation(); router.push(`/tickers/${t.symbol}`); }}>{t.symbol}</span>)}
                </div>
              )}
            </div>
          ))}
        </Card>
        {total > pageSize && <div style={{ display: 'flex', justifyContent: 'center', marginTop: 24 }}><Pagination current={page} pageSize={pageSize} total={total} onChange={setPage} showSizeChanger={false} showTotal={(t) => `${t} ${tt('events', lang)}`} /></div>}
      </>}
    </>
  );
}
