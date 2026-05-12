'use client';

import React from 'react';
import { Typography, Card, Row, Col, Spin, Empty, Tag, Button } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { api, EdgeData } from '@/lib/api';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

export default function EdgePage() {
  const { lang } = useThemeContext();
  const router = useRouter();
  const { data, isPending, isError, error, refetch } = useQuery<EdgeData>({ queryKey: ['edge'], queryFn: api.getEdge });

  if (isError) {
    return <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', height: '100%', gap: 12 }}>
      <Typography.Text type="danger">{tt('loadError', lang)}: {error?.message}</Typography.Text>
      <Button onClick={() => refetch()}>{tt('retry', lang)}</Button>
    </div>;
  }

  if (isPending && !data) {
    return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}><Spin size="large" /></div>;
  }

  if (!data) return <Empty description={tt('noData', lang)} />;

  const hasData = data.rising_themes.length > 0 || data.hot_tickers.length > 0 || data.recent_events.length > 0;

  return (
    <>
      <Typography.Text type="secondary" style={{ fontSize: 13, marginBottom: 4, display: 'block' }}>
        {tt('edgeDesc', lang)}
      </Typography.Text>

      {!hasData ? (
        <Card size="small" style={{ marginTop: 16 }}>
          <Empty description={tt('noData', lang)} />
        </Card>
      ) : (
        <Row gutter={[16, 16]} style={{ marginTop: 12 }}>
          {/* Rising Themes */}
          <Col xs={24} lg={8}>
            <Card
              title={<Typography.Title level={5} style={{ margin: 0, fontWeight: 600 }}>{tt('risingThemes', lang)}</Typography.Title>}
              size="small"
            >
              {data.rising_themes.length === 0 ? (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={tt('noThemes', lang)} />
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  {data.rising_themes.map((theme) => (
                    <div
                      key={theme.id}
                      className="ticker-row"
                      style={{ padding: '10px 12px', cursor: 'pointer', borderRadius: 8, background: 'var(--nshock-bg-elevated)' }}
                      onClick={() => router.push(`/themes/${theme.id}`)}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
                        <div style={{ fontWeight: 600, fontSize: 14, flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {theme.name}
                        </div>
                        <Tag color="green" style={{ fontSize: 10, lineHeight: '18px', marginLeft: 8 }}>{tt('emerging', lang)}</Tag>
                      </div>
                      <div style={{ fontSize: 12, color: 'var(--nshock-text-muted)', marginBottom: 6 }}>
                        {theme.category && <>{tt(theme.category, lang)} · </>}
                        {theme.ticker_count}{tt('tickers_label', lang)} · {theme.event_count}{tt('events_label', lang)}
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <div className="strength-bar" style={{ flex: 1 }}>
                          <div className="fill" style={{ width: `${Math.min(theme.strength_norm, 100)}%` }} />
                        </div>
                        <span style={{ fontSize: 11, color: 'var(--nshock-text-muted)', minWidth: 28, textAlign: 'right' }}>
                          {theme.strength_norm.toFixed(0)}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </Col>

          {/* Hot Tickers */}
          <Col xs={24} lg={8}>
            <Card
              title={<Typography.Title level={5} style={{ margin: 0, fontWeight: 600 }}>{tt('hotTickers', lang)}</Typography.Title>}
              size="small"
            >
              {data.hot_tickers.length === 0 ? (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={tt('noTickers', lang)} />
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  {data.hot_tickers.map((ticker) => (
                    <div
                      key={ticker.id}
                      className="ticker-row"
                      style={{ padding: '10px 12px', cursor: 'pointer', borderRadius: 8, background: 'var(--nshock-bg-elevated)' }}
                      onClick={() => router.push(`/tickers/${ticker.symbol}`)}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div>
                          <span style={{ fontWeight: 700, fontSize: 14 }}>{ticker.symbol}</span>
                          <span style={{ marginLeft: 8, fontSize: 12, color: 'var(--nshock-text-muted)' }}>{ticker.name}</span>
                        </div>
                        <div style={{ textAlign: 'right' }}>
                          <div style={{ fontWeight: 700, fontSize: 16, color: 'var(--nshock-primary)' }}>{ticker.hot_score?.toFixed(0)}</div>
                          <div style={{ fontSize: 10, color: 'var(--nshock-text-muted)' }}>{tt('hotScore', lang)}</div>
                        </div>
                      </div>
                      <div style={{ display: 'flex', gap: 12, marginTop: 4, fontSize: 11, color: 'var(--nshock-text-muted)' }}>
                        <span>{ticker.market?.toUpperCase()}</span>
                        <span>{ticker.mention_count} {tt('mentions', lang)}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </Col>

          {/* Signal Events */}
          <Col xs={24} lg={8}>
            <Card
              title={<Typography.Title level={5} style={{ margin: 0, fontWeight: 600 }}>{tt('signalEvents', lang)}</Typography.Title>}
              size="small"
            >
              {data.recent_events.length === 0 ? (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={tt('noEvents', lang)} />
              ) : (
                <div>
                  {data.recent_events.map((event) => (
                    <div key={event.id} className="event-item" style={{ padding: '10px 0' }}>
                      <div className="event-meta">
                        <span>{event.event_time?.slice(0, 16) || event.created_at?.slice(0, 16)}</span>
                        {event.channel && <><span>·</span><span>{event.channel}</span></>}
                        {event.theme_name && (
                          <><span>·</span>
                          <span
                            style={{ color: 'var(--nshock-primary)', fontSize: 10, cursor: 'pointer' }}
                            onClick={() => event.theme_id && router.push(`/themes/${event.theme_id}`)}
                          >
                            {event.theme_name}
                          </span></>
                        )}
                        <div className="importance">
                          {[1,2,3,4,5].map((i) => <div key={i} className={`dot ${i <= event.importance ? 'filled' : ''}`} />)}
                        </div>
                      </div>
                      <div className="event-title">{event.title}</div>
                      {event.summary && <div className="event-summary">{event.summary.slice(0, 100)}</div>}
                      {event.tickers?.length > 0 && (
                        <div className="event-tickers">
                          {event.tickers.map((t) => (
                            <span key={t.symbol} className="ticker-theme-tag">{t.symbol}</span>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </Col>
        </Row>
      )}
    </>
  );
}
