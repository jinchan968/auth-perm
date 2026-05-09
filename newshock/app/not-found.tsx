'use client';

import React from 'react';
import { Button, Typography } from 'antd';
import { useRouter } from 'next/navigation';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

export default function NotFound() {
  const router = useRouter();
  const { lang } = useThemeContext();

  return (
    <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', height: '100%', gap: 16 }}>
      <Typography.Title level={2} style={{ margin: 0, color: 'var(--nshock-text-muted)' }}>404</Typography.Title>
      <Typography.Text type="secondary">{tt('noData', lang)}</Typography.Text>
      <Button type="primary" onClick={() => router.push('/')}>{tt('radar', lang)}</Button>
    </div>
  );
}
