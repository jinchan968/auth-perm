'use client';

import React, { useState, useEffect } from 'react';
import { Typography, Card, Row, Col, Spin, Descriptions, Tag, Button, message } from 'antd';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useRouter } from 'next/navigation';
import { api, ThemeDetail, Polymarket, displaySymbol } from '@/lib/api';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

const WATCHLIST_KEY = 'newshock-watchlist';

export default function ThemeDetailPage() {
  const { lang } = useThemeContext();
  const router = useRouter();
  const params = useParams();
  const id = String(params.id);
  const [isWatchlisted, setIsWatchlisted] = useState(false);

  useEffect(() => {
    try { const list: string[] = JSON.parse(localStorage.getItem(WATCHLIST_KEY) || '[]'); setIsWatchlisted(list.includes(String(id))); } catch {}
  }, [id]);

  const toggleWatchlist = () => {
    try {
      const list: string[] = JSON.parse(localStorage.getItem(WATCHLIST_KEY) || '[]');
      const sid = String(id);
      const next = list.includes(sid) ? list.filter((x) => x !== sid) : [...list, sid];
      localStorage.setItem(WATCHLIST_KEY, JSON.stringify(next));
      setIsWatchlisted((prev) => !prev);
    } catch {}
  };

  const queryClient = useQueryClient();
  const { data: theme, isLoading, isError, error, refetch } = useQuery({ queryKey: ['theme', id], queryFn: () => api.getTheme(id), enabled: !!id });

  const generateMutation = useMutation({
    mutationFn: () => api.generateThemeDescription(id),
    onSuccess: (data) => {
      queryClient.setQueryData(['theme', id], (old: ThemeDetail | undefined) => old ? { ...old, description: data.description } : old);
      message.success(tt('descriptionGenerated', lang));
    },
    onError: () => message.error(tt('descriptionGenFailed', lang)),
  });

  if (isError) {
    return <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', height: '100%', gap: 12 }}>
      <Typography.Text type="danger">{tt('loadError', lang)}: {error?.message}</Typography.Text>
      <Button onClick={() => refetch()}>{tt('retry', lang)}</Button>
    </div>;
  }

  if (isLoading || !theme) {
    return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}><Spin size="large" /></div>;
  }

  const t = theme as ThemeDetail;

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Typography.Title level={5} style={{ margin: 0 }}>{t.name}</Typography.Title>
        <Button size="small" type={isWatchlisted ? 'primary' : 'default'} onClick={toggleWatchlist}>
          {isWatchlisted ? tt('removeWatchlist', lang) : tt('addWatchlist', lang)}
        </Button>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card size="small" style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2 }} size="small">
              <Descriptions.Item label={tt('themes', lang)}>{t.name}</Descriptions.Item>
              <Descriptions.Item label={tt('category', lang)}><Tag>{tt(t.category, lang)}</Tag></Descriptions.Item>
              <Descriptions.Item label={tt('strength', lang)}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <div className="strength-bar" style={{ width: 120 }}><div className="fill" style={{ width: `${Math.min(t.strength_norm, 100)}%` }} /></div>
                  <span>{t.strength.toFixed(1)}</span>
                  <span style={{ color: 'var(--nshock-text-muted)', fontSize: 12 }}>({t.strength_norm.toFixed(0)}%)</span>
                </div>
              </Descriptions.Item>
              <Descriptions.Item label={tt('trend', lang)}><span className={`regime-badge ${t.trend}`}>{tt(t.trend, lang)}</span></Descriptions.Item>
              <Descriptions.Item label={tt('tickers_label', lang).trim()}>{t.ticker_count}</Descriptions.Item>
              <Descriptions.Item label={tt('events_label', lang).trim()}>{t.event_count}</Descriptions.Item>
              {t.classification_confidence > 0 && <Descriptions.Item label={tt('confidence', lang)}>{(t.classification_confidence * 100).toFixed(0)}%</Descriptions.Item>}
            </Descriptions>
            <div style={{ marginTop: 12 }}>
              {t.description && <Typography.Paragraph type="secondary" style={{ fontSize: 13, marginBottom: 8 }}>{t.description}</Typography.Paragraph>}
              <Button
                size="small"
                loading={generateMutation.isPending}
                onClick={() => generateMutation.mutate()}
              >
                {generateMutation.isPending ? tt('generating', lang) : tt('generateDescription', lang)}
              </Button>
            </div>
          </Card>

          <Card title={<Typography.Title level={5} style={{ margin: 0 }}>{tt('recentEvents', lang)}</Typography.Title>} size="small">
            {t.events && t.events.length > 0 ? t.events.map((event) => (
              <div key={event.id} className="event-item" style={{ cursor: 'pointer' }} onClick={() => router.push(`/events/${event.id}`)}>
                <div className="event-meta">
                  <span>{event.event_time?.slice(0, 16) || event.created_at?.slice(0, 16)}</span>
                  {event.channel && <><span>·</span><span>{event.channel}</span></>}
                  <div className="importance">{[1,2,3,4,5].map((i) => <div key={i} className={`dot ${i <= event.importance ? 'filled' : ''}`} />)}</div>
                </div>
                <div className="event-title">{event.title}</div>
                {event.summary && <div className="event-summary">{event.summary.slice(0, 150)}</div>}
                {event.tickers?.length > 0 && <div className="event-tickers">{event.tickers.map((tk) => <span key={tk.symbol} className="ticker-theme-tag">{tk.symbol}</span>)}</div>}
              </div>
            )) : <div style={{ color: 'var(--nshock-text-muted)', textAlign: 'center', padding: 24 }}>{tt('noEvents', lang)}</div>}
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title={<Typography.Title level={5} style={{ margin: 0 }}>{tt('associatedTickers', lang)}</Typography.Title>} size="small">
            {t.tickers && t.tickers.length > 0 ? t.tickers.map((ticker) => (
              <div key={ticker.id} className="ticker-row" onClick={() => router.push(`/tickers/${ticker.symbol}`)}>
                <Row align="middle" gutter={6}>
                  <Col>
                    <span style={{ fontWeight: 700, fontSize: 13 }}>{displaySymbol(ticker.symbol)}</span>
                    <span style={{ marginLeft: 6, fontSize: 11, color: 'var(--nshock-text-muted)' }}>{ticker.market?.toUpperCase()}</span>
                  </Col>
                </Row>
                <div style={{ fontSize: 12, color: 'var(--nshock-text-muted)', marginTop: 2 }}>{ticker.name}</div>
              </div>
            )) : <div style={{ color: 'var(--nshock-text-muted)', textAlign: 'center', padding: 24 }}>{tt('noTickers', lang)}</div>}
          </Card>

          <Card title={<Typography.Title level={5} style={{ margin: 0 }}>{tt('polymarket', lang)}</Typography.Title>} size="small" style={{ marginTop: 16 }}>
            {t.polymarket && t.polymarket.length > 0 ? t.polymarket.map((pm) => (
              <div key={pm.condition_id} style={{ padding: '8px 0', borderBottom: '1px solid var(--nshock-border)' }}>
                <div style={{ fontSize: 13, fontWeight: 600, marginBottom: 4 }}>{pm.title}</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <span style={{ fontSize: 20, fontWeight: 700, color: 'var(--nshock-primary)' }}>{(pm.probability * 100).toFixed(0)}%</span>
                  <div className="strength-bar" style={{ flex: 1, maxWidth: 100 }}>
                    <div className="fill" style={{ width: `${Math.min(pm.probability * 100, 100)}%` }} />
                  </div>
                  {pm.volume > 0 && <span style={{ fontSize: 11, color: 'var(--nshock-text-muted)' }}>${(pm.volume / 1000).toFixed(0)}k</span>}
                </div>
              </div>
            )) : <div style={{ color: 'var(--nshock-text-muted)', textAlign: 'center', padding: 24 }}>{tt('noPolymarket', lang)}</div>}
          </Card>
        </Col>
      </Row>
    </>
  );
}
