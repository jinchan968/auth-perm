'use client';

import React from 'react';
import { Typography, Card, Row, Col, Spin, Empty, Button } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { api, Polymarket } from '@/lib/api';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

export default function MarketsPage() {
  const { lang } = useThemeContext();
  const { data: regime, isPending: regimePending, isError: regimeError, error: regimeErr } = useQuery({ queryKey: ['home'], queryFn: api.getHome, select: (d) => d.regime });
  const { data: polymarkets, isPending: pmPending, isError: pmError, error: pmErr, refetch: refetchPm } = useQuery<Polymarket[]>({ queryKey: ['polymarket'], queryFn: api.getPolymarket });

  const isLoading = (regimePending && !regime) || (pmPending && !polymarkets);
  const isError = regimeError || pmError;
  const error = regimeErr || pmErr;

  if (isError) {
    return <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', height: '100%', gap: 12 }}>
      <Typography.Text type="danger">{tt('loadError', lang)}: {error?.message}</Typography.Text>
      <Button onClick={() => refetchPm()}>{tt('retry', lang)}</Button>
    </div>;
  }

  if (isLoading) {
    return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}><Spin size="large" /></div>;
  }

  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} lg={8}>
        <Card title={<Typography.Title level={5} style={{ margin: 0 }}>{tt('marketRegime', lang)}</Typography.Title>} size="small">
          {regime ? (
            <div>
              <div style={{ textAlign: 'center', marginBottom: 16 }}>
                <span className={`regime-badge ${regime.regime_type}`} style={{ fontSize: 16, padding: '6px 24px' }}>{tt(regime.regime_type, lang)}</span>
              </div>
              <div style={{ textAlign: 'center', marginBottom: 12 }}>
                <Typography.Text style={{ fontSize: 28, fontWeight: 700 }}>{(regime.confidence * 100).toFixed(0)}%</Typography.Text>
                <div style={{ fontSize: 12, color: 'var(--nshock-text-muted)' }}>{tt('probability', lang)}</div>
              </div>
              <Typography.Paragraph type="secondary" style={{ fontSize: 13 }}>{regime.summary}</Typography.Paragraph>
              <div style={{ fontSize: 11, color: 'var(--nshock-text-muted)', textAlign: 'right' }}>
                {regime.created_at?.slice(0, 16)}
              </div>
            </div>
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={tt('noData', lang)} />
          )}
        </Card>
      </Col>

      <Col xs={24} lg={16}>
        <Card title={<Typography.Title level={5} style={{ margin: 0 }}>{tt('polymarket', lang)}</Typography.Title>} size="small">
          {!polymarkets || polymarkets.length === 0 ? (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={tt('noPolymarket', lang)} />
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {polymarkets.map((pm) => (
                <div key={pm.condition_id} style={{ padding: '10px 14px', borderRadius: 8, background: 'var(--nshock-bg-elevated)' }}>
                  <div style={{ fontWeight: 600, fontSize: 14, marginBottom: 8 }}>{pm.title}</div>
                  {pm.description && (
                    <div style={{ fontSize: 12, color: 'var(--nshock-text-muted)', marginBottom: 8, lineHeight: 1.5 }}>
                      {pm.description.slice(0, 120)}{pm.description.length > 120 ? '...' : ''}
                    </div>
                  )}
                  <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, flex: 1 }}>
                      <span style={{ fontSize: 22, fontWeight: 700, color: 'var(--nshock-primary)' }}>{(pm.probability * 100).toFixed(0)}%</span>
                      <div className="strength-bar" style={{ flex: 1, maxWidth: 160 }}>
                        <div className="fill" style={{ width: `${Math.min(pm.probability * 100, 100)}%` }} />
                      </div>
                    </div>
                    <div style={{ textAlign: 'right' }}>
                      {pm.outcome && <div style={{ fontSize: 12, color: 'var(--nshock-text-muted)' }}>{pm.outcome}</div>}
                      {pm.volume > 0 && <div style={{ fontSize: 11, color: 'var(--nshock-text-muted)' }}>{tt('volume', lang)} ${(pm.volume / 1000).toFixed(0)}k</div>}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      </Col>
    </Row>
  );
}
