'use client';

import React from 'react';
import { Typography, Card, Row, Col, Spin, Descriptions, Tag, Button } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { useParams, useRouter } from 'next/navigation';
import { api, TickerDetail } from '@/lib/api';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

export default function TickerDetailPage() {
  const { lang } = useThemeContext();
  const router = useRouter();
  const params = useParams();
  const symbol = decodeURIComponent(params.symbol as string);

  const { data: ticker, isLoading, isError, error, refetch } = useQuery({ queryKey: ['ticker', symbol], queryFn: () => api.getTicker(symbol), enabled: !!symbol });

  if (isError) {
    return <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', height: '100%', gap: 12 }}>
      <Typography.Text type="danger">{tt('loadError', lang)}: {error?.message}</Typography.Text>
      <Button onClick={() => refetch()}>{tt('retry', lang)}</Button>
    </div>;
  }

  if (isLoading || !ticker) {
    return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}><Spin size="large" /></div>;
  }

  const t = ticker as TickerDetail;

  return (
    <>
      <Typography.Title level={5} style={{ margin: 0, marginBottom: 16 }}>{t.symbol} — {t.name}</Typography.Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card size="small" style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2 }} size="small">
              <Descriptions.Item label={tt('symbol', lang)}><span style={{ fontWeight: 700, fontSize: 16 }}>{t.symbol}</span></Descriptions.Item>
              <Descriptions.Item label={tt('name', lang)}>{t.name}</Descriptions.Item>
              <Descriptions.Item label={tt('filterMarket', lang)}><Tag>{t.market?.toUpperCase()}</Tag></Descriptions.Item>
              <Descriptions.Item label={tt('hotScore', lang)}><span style={{ fontWeight: 700, fontSize: 18 }}>{t.hot_score?.toFixed(0)}</span></Descriptions.Item>
              {t.mention_count > 0 && <Descriptions.Item label={tt('mentions', lang)}>{t.mention_count}</Descriptions.Item>}
            </Descriptions>
          </Card>

          <Card title={<Typography.Title level={5} style={{ margin: 0 }}>{tt('recentEvents', lang)}</Typography.Title>} size="small">
            {t.events && t.events.length > 0 ? t.events.map((event) => (
              <div key={event.id} className="event-item" style={{ cursor: 'pointer' }} onClick={() => router.push(`/events/${event.id}`)}>
                <div className="event-meta">
                  <span>{event.event_time?.slice(0, 16) || event.created_at?.slice(0, 16)}</span>
                  {event.channel && <><span>·</span><span>{event.channel}</span></>}
                  {event.theme_name && <><span>·</span><span style={{ color: 'var(--nshock-primary)', fontSize: 10 }}>{event.theme_name}</span></>}
                  <div className="importance">{[1,2,3,4,5].map((i) => <div key={i} className={`dot ${i <= event.importance ? 'filled' : ''}`} />)}</div>
                </div>
                <div className="event-title">{event.title}</div>
                {event.summary && <div className="event-summary">{event.summary.slice(0, 150)}</div>}
              </div>
            )) : <div style={{ color: 'var(--nshock-text-muted)', textAlign: 'center', padding: 24 }}>{tt('noEvents', lang)}</div>}
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title={<Typography.Title level={5} style={{ margin: 0 }}>{tt('themes', lang)}</Typography.Title>} size="small">
            {t.themes && t.themes.length > 0 ? t.themes.map((theme) => (
              <div key={theme.id} className="ticker-row" onClick={() => router.push(`/themes/${theme.id}`)}>
                <div style={{ fontWeight: 600, fontSize: 14 }}>{theme.name}</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 4 }}>
                  <div className="strength-bar" style={{ width: 80 }}><div className="fill" style={{ width: `${Math.min(theme.strength_norm, 100)}%` }} /></div>
                  <span style={{ fontSize: 11, color: 'var(--nshock-text-muted)' }}>{theme.strength.toFixed(0)}</span>
                  <span className={`regime-badge ${theme.trend}`} style={{ fontSize: 10 }}>{tt(theme.trend, lang)}</span>
                </div>
              </div>
            )) : <div style={{ color: 'var(--nshock-text-muted)', textAlign: 'center', padding: 24 }}>{tt('noThemes', lang)}</div>}
          </Card>
        </Col>
      </Row>
    </>
  );
}
