'use client';

import React from 'react';
import { Input, Button } from 'antd';
import { SearchOutlined, SunOutlined, MoonOutlined } from '@ant-design/icons';
import { useThemeContext } from '@/lib/theme-context';
import { tt } from '@/lib/i18n';

interface TopbarProps {
  onOpenSearch: () => void;
}

export default function Topbar({ onOpenSearch }: TopbarProps) {
  const { lang, toggleLang, theme, toggleTheme } = useThemeContext();
  const isDark = theme === 'dark';

  return (
    <div className="topbar-glass">
      <Input
        prefix={<SearchOutlined style={{ color: 'var(--nshock-text-muted)' }} />}
        suffix={<span style={{ fontSize: 11, color: 'var(--nshock-text-muted)', background: 'rgba(128,128,128,0.1)', padding: '2px 6px', borderRadius: 4 }}>⌘K</span>}
        placeholder={tt('searchGlobal', lang)}
        onClick={onOpenSearch}
        readOnly
        style={{ flex: 0.4, minWidth: 260, cursor: 'pointer' }}
      />
      <div className="topbar-actions">
        <Button type="text" onClick={toggleLang} className="topbar-iconbtn">
          <span className="topbar-iconbtn-inner">{lang === 'zh' ? 'EN' : '中'}</span>
        </Button>
        <Button
          type="text"
          icon={isDark ? <SunOutlined /> : <MoonOutlined />}
          onClick={toggleTheme}
          className="topbar-iconbtn"
          aria-label={isDark ? tt('switchToLight', lang) : tt('switchToDark', lang)}
        />
      </div>
    </div>
  );
}
