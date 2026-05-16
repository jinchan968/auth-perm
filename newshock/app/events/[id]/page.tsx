'use client';

import React from 'react';
import { Typography, Card, Row, Col, Spin, Descriptions, Tag, Button } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { useParams, useRouter } from 'next/navigation';
import { api, Event, displaySymbol } from '@/lib/api';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

export default function EventDetailPage() {
  const { lang } = useThemeContext();
  const router = useRouter();
  const params = useParams();
  const id = String(params.id);

  const { data: event, isLoading, isError, error, refetch } = useQuery<Event>({
    queryKey: ['event', id],
    queryFn: () => api.getEvent(id),
    enabled: !!id,
  });

  if (isError) {
    return <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', height: '100%', gap: 12 }}>
      <Typography.Text type="danger">{tt('loadError', lang)}: {error?.message}</Typography.Text>
      <Button onClick={() => refetch()}>{tt('retry', lang)}</Button>
    </div>;
  }

  if (isLoading || !event) {
    return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}><Spin size="large" /></div>;
  }

  return (
    <>
      <Typography.Title level={5} style={{ marginBottom: 16 }}>{tt('eventDetail', lang)}</Typography.Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card size="small" style={{ marginBottom: 16 }}>
            <div style={{ marginBottom: 12 }}>
              <Typography.Title level={4} style={{ margin: 0, marginBottom: 8 }}>{event.title}</Typography.Title>
              <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'center' }}>
                {event.channel && <Tag>{event.channel}</Tag>}
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <span style={{ fontSize: 12, color: 'var(--nshock-text-muted)' }}>{tt('importance', lang)}:</span>
                  <div className="importance" style={{ display: 'inline-flex' }}>
                    {[1,2,3,4,5].map((i) => <div key={i} className={`dot ${i <= event.importance ? 'filled' : ''}`} />)}
                  </div>
                </div>
              </div>
            </div>

            <Descriptions column={{ xs: 1, sm: 2 }} size="small" style={{ marginBottom: 12 }}>
              {event.event_time && (
                <Descriptions.Item label={tt('eventTime', lang)}>{event.event_time.slice(0, 16)}</Descriptions.Item>
              )}
              <Descriptions.Item label={tt('events', lang)}>{event.created_at?.slice(0, 16)}</Descriptions.Item>
              {event.theme_name && (
                <Descriptions.Item label={tt('themes', lang)}>
                  <Tag
                    color="purple"
                    style={{ cursor: 'pointer' }}
                    onClick={() => event.theme_id && router.push(`/themes/${event.theme_id}`)}
                  >
                    {event.theme_name}
                  </Tag>
                </Descriptions.Item>
              )}
            </Descriptions>

            {event.summary && (
              <Typography.Paragraph type="secondary" style={{ fontSize: 14, lineHeight: 1.8 }}>
                {event.summary}
              </Typography.Paragraph>
            )}
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title={<Typography.Title level={5} style={{ margin: 0 }}>{tt('associatedTickers', lang)}</Typography.Title>} size="small">
            {event.tickers && event.tickers.length > 0 ? event.tickers.map((ticker) => (
              <div
                key={ticker.symbol}
                className="ticker-row"
                onClick={() => router.push(`/tickers/${ticker.symbol}`)}
              >
                <Row align="middle" gutter={6}>
                  <Col>
                    <span style={{ fontWeight: 700, fontSize: 13 }}>{displaySymbol(ticker.symbol)}</span>
                    {ticker.market && <span style={{ marginLeft: 6, fontSize: 11, color: 'var(--nshock-text-muted)' }}>{ticker.market.toUpperCase()}</span>}
                  </Col>
                </Row>
                {ticker.name && <div style={{ fontSize: 12, color: 'var(--nshock-text-muted)', marginTop: 2 }}>{ticker.name}</div>}
              </div>
            )) : (
              <div style={{ color: 'var(--nshock-text-muted)', textAlign: 'center', padding: 24 }}>{tt('noTickers', lang)}</div>
            )}
          </Card>

          {event.theme_id && (
            <Card
              title={<Typography.Title level={5} style={{ margin: 0 }}>{tt('themes', lang)}</Typography.Title>}
              size="small"
              style={{ marginTop: 16 }}
            >
              <div
                className="ticker-row"
                onClick={() => router.push(`/themes/${event.theme_id}`)}
                style={{ cursor: 'pointer' }}
              >
                <span style={{ fontWeight: 600, fontSize: 14 }}>{event.theme_name}</span>
                <div style={{ fontSize: 12, color: 'var(--nshock-primary)', marginTop: 4 }}>{tt('viewTheme', lang)}</div>
              </div>
            </Card>
          )}
        </Col>
      </Row>
    </>
  );
}
