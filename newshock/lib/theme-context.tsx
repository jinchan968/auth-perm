'use client';

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';

interface ThemeContextValue {
  lang: string;
  toggleLang: () => void;
  theme: 'dark' | 'light';
  toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextValue>({
  lang: 'zh',
  toggleLang: () => {},
  theme: 'dark',
  toggleTheme: () => {},
});

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [lang, setLang] = useState('zh');
  const [theme, setTheme] = useState<'dark' | 'light'>('dark');
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const savedLang = localStorage.getItem('newshock-lang');
    if (savedLang === 'zh' || savedLang === 'en') setLang(savedLang);
    const savedTheme = localStorage.getItem('theme-mode');
    if (savedTheme === 'light' || savedTheme === 'dark') {
      setTheme(savedTheme);
      document.documentElement.dataset.theme = savedTheme;
    }
    setMounted(true);
  }, []);

  useEffect(() => {
    if (mounted) {
      document.documentElement.dataset.theme = theme;
      document.cookie = `theme-mode=${theme};path=/;max-age=31536000;samesite=lax`;
    }
  }, [theme, mounted]);

  useEffect(() => {
    if (mounted) {
      document.documentElement.lang = lang === 'en' ? 'en' : 'zh-CN';
    }
  }, [lang, mounted]);

  const toggleLang = useCallback(() => {
    setLang((prev) => {
      const next = prev === 'zh' ? 'en' : 'zh';
      localStorage.setItem('newshock-lang', next);
      return next;
    });
  }, []);

  const toggleTheme = useCallback(() => {
    setTheme((prev) => {
      const next = prev === 'dark' ? 'light' : 'dark';
      localStorage.setItem('theme-mode', next);
      return next;
    });
  }, []);

  if (!mounted) return null;

  return (
    <ThemeContext.Provider value={{ lang, toggleLang, theme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useThemeContext() {
  return useContext(ThemeContext);
}
