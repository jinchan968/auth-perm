const BASE_URL = '/api/v1/newshock';

function getTokenFromCookie(): string | null {
  if (typeof window === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)auth_token=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : null;
}

function getLoginURL(): string {
  if (typeof window === 'undefined') return '/login';
  const { protocol, hostname } = window.location;
  // 主 UI 默认在 3000 端口，可通过环境变量覆盖
  const mainPort = process.env.NEXT_PUBLIC_MAIN_PORT || '3000';
  return `${protocol}//${hostname}:${mainPort}/login`;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = getTokenFromCookie();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (options?.headers) {
    Object.assign(headers, options.headers);
  }
  if (token) {
    headers['x-auth-token'] = token;
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
  });

  if (res.status === 401) {
    window.location.href = getLoginURL();
    throw new Error('未登录，请先登录');
  }

  if (!res.ok) {
    let msg = `API error: ${res.status}`;
    try {
      const body = await res.json();
      if (body.error) msg = body.error;
      else if (body.message) msg = body.message;
    } catch {}
    throw new Error(msg);
  }
  const json = await res.json();
  return json.data ?? json;
}

export interface Theme {
  id: string;
  name: string;
  description: string;
  category: string;
  strength: number;
  strength_norm: number;
  classification_confidence: number;
  ticker_count: number;
  event_count: number;
  trend: 'rising' | 'stable' | 'declining';
  created_at: string;
  updated_at: string;
}

export interface Ticker {
  id: string;
  symbol: string;
  name: string;
  market: string;
  hot_score: number;
  mention_count: number;
}

export interface Event {
  id: string;
  title: string;
  summary: string;
  channel: string;
  importance: number;
  theme_id: string;
  theme_name: string;
  created_at: string;
  event_time: string;
  tickers: { id?: string; symbol: string; name?: string; market?: string }[];
}

export interface Regime {
  regime_type: 'risk_on' | 'risk_off' | 'neutral';
  confidence: number;
  summary: string;
  created_at: string;
}

export interface Stats {
  theme_count: number;
  ticker_count: number;
  event_count: number;
  avg_strength: number;
}

export interface Freshness {
  themes_updated: string;
  events_updated: string;
  tickers_updated: string;
}

export interface HomeData {
  top_themes: Theme[];
  top_tickers: Ticker[];
  recent_events: Event[];
  regime: Regime | null;
  stats: Stats;
  freshness: Freshness;
}

export interface Polymarket {
  condition_id: string;
  title: string;
  description: string;
  outcome: string;
  probability: number;
  volume: number;
  updated_at: string;
}

export interface ThemeDetail extends Theme {
  tickers?: Ticker[];
  events?: Event[];
  polymarket?: Polymarket[];
}

export interface TickerConcept {
  name: string;
  type: string; // concept/industry/region
}

export interface TickerDetail extends Ticker {
  themes?: Theme[];
  events?: Event[];
  daily?: TickerDaily[];
  concepts?: TickerConcept[];
}

export interface TickerDaily {
  date: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  amount: number;
  change_pct: number;
  turnover: number;
}

export interface SearchResults {
  themes: Theme[];
  tickers: Ticker[];
  events: Event[];
}

export interface PagedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface EdgeData {
  rising_themes: Theme[];
  hot_tickers: Ticker[];
  recent_events: Event[];
}

export interface PipelineStatus {
  news_total: number;
  news_unprocessed: number;
  theme_count: number;
  ticker_count: number;
  event_count: number;
  polymarket_count: number;
  latest_news_time: string | null;
  latest_event_time: string | null;
}

export const api = {
  getHome: () => request<HomeData>('/home'),
  getThemes: (params?: Record<string, string>) =>
    request<PagedResponse<Theme>>(`/themes${params ? '?' + new URLSearchParams(params) : ''}`),
  getTheme: (id: string) => request<ThemeDetail>(`/themes/${id}`),
  getTickers: (params?: Record<string, string>) =>
    request<PagedResponse<Ticker>>(`/tickers${params ? '?' + new URLSearchParams(params) : ''}`),
  getTicker: (symbol: string) => request<TickerDetail>(`/tickers/${symbol}`),
  getTickerDaily: (symbol: string, days?: number) =>
    request<TickerDaily[]>(`/tickers/${symbol}/daily${days ? `?days=${days}` : ''}`),
  getEvents: (params?: Record<string, string>) =>
    request<PagedResponse<Event>>(`/events${params ? '?' + new URLSearchParams(params) : ''}`),
  getEvent: (id: string) => request<Event>(`/events/${id}`),
  search: (keyword: string) => request<SearchResults>(`/search?keyword=${encodeURIComponent(keyword)}`),
  getPipeline: () => request<PipelineStatus>('/pipeline'),
  getEdge: () => request<EdgeData>('/edge'),
  generateThemeDescription: (id: string) =>
    request<{ description: string }>(`/themes/${id}/generate-description`, { method: 'POST' }),
  getPolymarket: () => request<Polymarket[]>('/polymarket'),
};
