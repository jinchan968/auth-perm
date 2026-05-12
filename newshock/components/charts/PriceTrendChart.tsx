'use client';

import React, { useState } from 'react';
import { Card, Radio, Space, Typography, Empty } from 'antd';
import { Line, Column } from '@ant-design/charts';
import { TickerDaily } from '@/lib/api';
import { tt } from '@/lib/i18n';

interface Props {
  data: TickerDaily[];
  lang: string;
}

export default function PriceTrendChart({ data, lang }: Props) {
  const [range, setRange] = useState(90);

  if (!data || data.length === 0) {
    return (
      <Card size="small">
        <Empty description={tt('noData', lang)} />
      </Card>
    );
  }

  const filtered = data.slice(-range);

  const priceData = filtered.map(d => ({
    date: d.date,
    value: d.close,
  }));

  const volumeData = filtered.map(d => ({
    date: d.date,
    value: d.volume,
    color: d.change_pct >= 0 ? '#ef5350' : '#26a69a',
  }));

  const firstClose = filtered[0]?.close || 0;
  const lastClose = filtered[filtered.length - 1]?.close || 0;
  const priceColor = lastClose >= firstClose ? '#ef5350' : '#26a69a';

  const priceConfig = {
    data: priceData,
    xField: 'date',
    yField: 'value',
    color: priceColor,
    height: 200,
    smooth: true,
    point: { size: 0 },
    tooltip: {
      title: (d: { date: string }) => d.date,
      items: [{ field: 'value', name: tt('close', lang), valueFormatter: (v: number) => v?.toFixed(2) }],
    },
    axis: {
      x: { labelAutoRotate: false, label: { style: { fontSize: 10 } } },
      y: { title: tt('close', lang), label: { style: { fontSize: 10 } } },
    },
    interaction: { tooltip: { render: (e: Event, { title, items }: { title: string; items: { name: string; value: string }[] }) => { return `<div><strong>${title}</strong>${items.map(i => `<div>${i.name}: ${i.value}</div>`).join('')}</div>`; } } },
  };

  const volumeConfig = {
    data: volumeData,
    xField: 'date',
    yField: 'value',
    color: ({ color }: { color: string }) => color,
    height: 80,
    style: { radiusTopLeft: 2, radiusTopRight: 2 },
    axis: {
      x: { labelAutoRotate: false, label: { style: { fontSize: 10 } } },
      y: { title: tt('vol', lang), label: { style: { fontSize: 10 } } },
    },
    tooltip: {
      title: (d: { date: string }) => d.date,
      items: [{ field: 'value', name: tt('vol', lang), valueFormatter: (v: number) => v?.toLocaleString() }],
    },
  };

  return (
    <Card
      size="small"
      title={
        <Space>
          <Typography.Text strong>{tt('priceTrend', lang)}</Typography.Text>
          <Radio.Group value={range} onChange={e => setRange(e.target.value)} size="small">
            <Radio.Button value={30}>30{tt('days', lang)}</Radio.Button>
            <Radio.Button value={90}>90{tt('days', lang)}</Radio.Button>
            <Radio.Button value={180}>180{tt('days', lang)}</Radio.Button>
            <Radio.Button value={365}>1{tt('year', lang)}</Radio.Button>
          </Radio.Group>
        </Space>
      }
    >
      <Line {...priceConfig} />
      <div style={{ marginTop: 8 }}>
        <Column {...volumeConfig} />
      </div>
    </Card>
  );
}
