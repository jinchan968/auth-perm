'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { Typography, Card, Row, Col, Spin, Segmented, Button } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { api, HomeData } from '@/lib/api';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';
import { timeAgo } from '@/lib/time';

const WATCHLIST_KEY = 'newshock-watchlist';

function getWatchlist(): string[] {
  if (typeof window === 'undefined') return [];
  try { return JSON.parse(localStorage.getItem(WATCHLIST_KEY) || '[]'); } catch { return []; }
}

function saveWatchlist(ids: string[]) {
  if (typeof window === 'undefined') return;
  localStorage.setItem(WATCHLIST_KEY, JSON.stringify(ids));
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <Card size="small" styles={{ body: { padding: '14px 18px' } }}>
      <div style={{ fontSize: 12, color: 'var(--nshock-text-muted)', textTransform: 'uppercase' }}>{label}</div>
      <div style={{ fontSize: 28, fontWeight: 700 }}>{value.toLocaleString()}</div>
    </Card>
  );
}

export default function RadarPage() {
  const { lang } = useThemeContext();
  const router = useRouter();
  const { data, isPending, isError, error, refetch } = useQuery({ queryKey: ['home'], queryFn: api.getHome });
  const { data: pipeline } = useQuery({ queryKey: ['pipeline'], queryFn: api.getPipeline, refetchInterval: 60000 });
  const [leaderTab, setLeaderTab] = useState<string>('themes');
  const [watchlistIds, setWatchlistIds] = useState<string[]>([]);

  useEffect(() => { setWatchlistIds(getWatchlist()); }, []);

  const toggleWatchlist = useCallback((themeId: string) => {
    setWatchlistIds((prev) => {
      const next = prev.includes(themeId) ? prev.filter((id) => id !== themeId) : [...prev, themeId];
      saveWatchlist(next);
      return next;
    });
  }, []);

  if (isError) {
    return <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', height: '100%', gap: 12 }}>
      <Typography.Text type="danger">{tt('loadError', lang)}: {error?.message}</Typography.Text>
      <Button onClick={() => refetch()}>{tt('retry', lang)}</Button>
    </div>;
  }

  if (isPending && !data) {
    return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}><Spin size="large" /></div>;
  }

  const watchlistThemes = data.top_themes.filter((t) => watchlistIds.includes(String(t.id)));

  return (
    <>
      <Typography.Text type="secondary" style={{ fontSize: 13, marginBottom: 4, display: 'block' }}>
        {tt('liveSignals', lang)}
      </Typography.Text>

      {data.freshness && (
        <div style={{ fontSize: 11, color: 'var(--nshock-text-muted)', marginBottom: 16 }}>
          {tt('freshness', lang)}：
          {data.freshness.themes_updated && <>{tt('themeCount', lang)} {timeAgo(data.freshness.themes_updated, lang)}</>}
          {data.freshness.events_updated && <> · {tt('events', lang)} {timeAgo(data.freshness.events_updated, lang)}</>}
          {data.freshness.tickers_updated && <> · {tt('tickerCount', lang)} {timeAgo(data.freshness.tickers_updated, lang)}</>}
        </div>
      )}

      {pipeline && (
        <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', fontSize: 11, color: 'var(--nshock-text-muted)', marginBottom: 12 }}>
          <span>{tt('pipeline', lang)}:</span>
          <span>{tt('newsTotal', lang)} {pipeline.news_total.toLocaleString()}</span>
          {pipeline.news_unprocessed > 0 && <span style={{ color: 'var(--nshock-warning, #faad14)' }}>{tt('unprocessed', lang)} {pipeline.news_unprocessed}</span>}
          <span>{tt('themeCount', lang)} {pipeline.theme_count}</span>
          <span>{tt('tickerCount', lang)} {pipeline.ticker_count}</span>
          <span>{tt('eventCount', lang)} {pipeline.event_count}</span>
          <span>{tt('polymarketCount', lang)} {pipeline.polymarket_count}</span>
        </div>
      )}

      {watchlistThemes.length > 0 && (
        <div style={{ marginBottom: 16 }}>
          <Typography.Text style={{ fontSize: 13, fontWeight: 600, marginBottom: 8, display: 'block' }}>
            {tt('watchlist', lang)} ({watchlistThemes.length})
          </Typography.Text>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            {watchlistThemes.map((theme) => (
              <div key={theme.id} onClick={() => router.push(`/themes/${theme.id}`)}
                style={{ padding: '6px 12px', borderRadius: 8, background: 'var(--nshock-bg-elevated)', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
                <span style={{ fontWeight: 600 }}>{theme.name}</span>
                <span style={{ fontSize: 11, color: 'var(--nshock-text-muted)' }}>{tt('strength', lang)} {theme.strength_norm.toFixed(0)}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={8}><StatCard label={tt('themeCount', lang)} value={data.stats.theme_count} /></Col>
        <Col xs={24} sm={8}><StatCard label={tt('tickerCount', lang)} value={data.stats.ticker_count} /></Col>
        <Col xs={24} sm={8}><StatCard label={tt('eventCount', lang)} value={data.stats.event_count} /></Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={17}>
          <Card title={<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography.Title level={5} style={{ margin: 0, fontWeight: 600 }}>{tt('eventStream', lang)}</Typography.Title>
            <span style={{ fontSize: 12, color: 'var(--nshock-primary)', cursor: 'pointer' }} onClick={() => router.push('/events')}>{tt('viewAll', lang)}</span>
          </div>} size="small" style={{ marginBottom: 16 }}>
            <div style={{ maxHeight: 420, overflow: 'auto' }}>
              {data.recent_events.slice(0, 15).map((event) => (
                <div key={event.id} className="event-item" style={{ cursor: 'pointer' }} onClick={() => router.push(`/events/${event.id}`)}>
                  <div className="event-meta">
                    <span>{event.event_time?.slice(0, 16) || event.created_at?.slice(0, 16)}</span>
                    {event.channel && <><span>·</span><span>{event.channel}</span></>}
                    {event.theme_name && <><span>·</span><span style={{ color: 'var(--nshock-primary)', fontSize: 10 }}>{event.theme_name}</span></>}
                    <div className="importance">{[1,2,3,4,5].map((i) => <div key={i} className={`dot ${i <= event.importance ? 'filled' : ''}`} />)}</div>
                  </div>
                  <div className="event-title">{event.title}</div>
                  {event.summary && <div className="event-summary">{event.summary.slice(0, 130)}</div>}
                  {event.tickers?.length > 0 && <div className="event-tickers">{event.tickers.map((t) => <span key={t.symbol} className="ticker-theme-tag">{t.symbol}</span>)}</div>}
                </div>
              ))}
            </div>
          </Card>

          <Card title={<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
            <Typography.Title level={5} style={{ margin: 0, fontWeight: 600 }}>{tt('thematicLeaders', lang)}</Typography.Title>
            <Segmented size="small" value={leaderTab} onChange={(v) => setLeaderTab(v as string)}
              options={[{ label: tt('themeBoard', lang), value: 'themes' }, { label: tt('tickerBoard', lang), value: 'tickers' }]} />
          </div>} size="small">
            {leaderTab === 'themes' ? (
              data.top_themes.slice(0, 10).map((theme) => (
                <div key={theme.id} className="ticker-row" style={{ padding: '10px 0' }}>
                  <Row justify="space-between" align="middle" style={{ width: '100%' }} gutter={8}>
                    <Col flex="1" style={{ minWidth: 0 }}>
                      <div style={{ fontWeight: 600, fontSize: 14 }}>{theme.name}</div>
                      <div style={{ fontSize: 12, color: 'var(--nshock-text-muted)', marginTop: 2 }}>
                        {theme.ticker_count}{tt('tickers_label', lang)} · {theme.event_count}{tt('events_label', lang)}
                      </div>
                    </Col>
                    <Col style={{ textAlign: 'right', minWidth: 70 }}>
                      <div className="strength-bar" style={{ width: 100 }}><div className="fill" style={{ width: `${Math.min(theme.strength_norm, 100)}%` }} /></div>
                      <div style={{ fontSize: 11, color: 'var(--nshock-text-muted)', textAlign: 'right', marginTop: 2 }}>{theme.strength.toFixed(0)}</div>
                    </Col>
                  </Row>
                </div>
              ))
            ) : (
              data.top_tickers.slice(0, 10).map((ticker) => (
                <div key={ticker.id} className="ticker-row" style={{ padding: '10px 0' }} onClick={() => router.push(`/tickers/${ticker.symbol}`)}>
                  <Row justify="space-between" align="middle" style={{ width: '100%' }}>
                    <Col>
                      <span style={{ fontWeight: 700, fontSize: 14 }}>{ticker.symbol}</span>
                      <span style={{ marginLeft: 8, fontSize: 12, color: 'var(--nshock-text-muted)' }}>{ticker.name}</span>
                    </Col>
                    <Col style={{ textAlign: 'right' }}>
                      <span className="ticker-theme-tag">{ticker.market?.toUpperCase()}</span>
                      <div style={{ fontWeight: 700, fontSize: 14, marginTop: 2 }}>{ticker.hot_score?.toFixed(0)}</div>
                    </Col>
                  </Row>
                </div>
              ))
            )}
          </Card>
        </Col>

        <Col xs={24} lg={7}>
          <Card title={<Typography.Title level={5} style={{ margin: 0, fontWeight: 600 }}>{tt('marketRegime', lang)}</Typography.Title>} size="small" style={{ marginBottom: 16 }}>
            {data.regime ? (
              <div>
                <span className={`regime-badge ${data.regime.regime_type}`}>{tt(data.regime.regime_type, lang)}</span>
                <Typography.Text type="secondary" style={{ fontSize: 13, display: 'block', marginTop: 8 }}>{data.regime.summary}</Typography.Text>
              </div>
            ) : <Typography.Text type="secondary">{tt('noData', lang)}</Typography.Text>}
          </Card>

          <Card title={<Typography.Title level={5} style={{ margin: 0, fontWeight: 600 }}>{tt('events1w', lang)}</Typography.Title>} size="small" style={{ marginBottom: 16 }}>
            {data.recent_events.slice(0, 8).map((event) => (
              <div key={event.id} className="event-item" style={{ padding: '8px 0', cursor: 'pointer' }} onClick={() => router.push(`/events/${event.id}`)}>
                <div style={{ fontSize: 11, color: 'var(--nshock-text-muted)', marginBottom: 2 }}>
                  {event.event_time?.slice(5, 16) || event.created_at?.slice(5, 16)}
                  <div className="importance" style={{ display: 'inline-flex', marginLeft: 8 }}>{[1,2,3,4,5].map((i) => <div key={i} className={`dot ${i <= event.importance ? 'filled' : ''}`} />)}</div>
                </div>
                <div style={{ fontWeight: 600, fontSize: 13 }}>{event.title}</div>
              </div>
            ))}
          </Card>

          <Card title={<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography.Title level={5} style={{ margin: 0, fontWeight: 600 }}>{tt('tickerBoard', lang)}</Typography.Title>
            <span style={{ fontSize: 12, color: 'var(--nshock-primary)', cursor: 'pointer' }} onClick={() => router.push('/tickers')}>{tt('viewAll', lang)}</span>
          </div>} size="small">
            {data.top_tickers.slice(0, 8).map((ticker) => (
              <div key={ticker.id} className="ticker-row" style={{ padding: '8px 0' }} onClick={() => router.push(`/tickers/${ticker.symbol}`)}>
                <Row align="middle" gutter={6}>
                  <Col>
                    <span style={{ fontWeight: 700, fontSize: 13 }}>{ticker.symbol}</span>
                    <span style={{ marginLeft: 6, fontSize: 11, color: 'var(--nshock-text-muted)' }}>{ticker.market?.toUpperCase()}</span>
                  </Col>
                </Row>
                <div style={{ fontWeight: 700, fontSize: 13, marginTop: 2 }}>{ticker.hot_score?.toFixed(0)}</div>
              </div>
            ))}
          </Card>
        </Col>
      </Row>
    </>
  );
}
