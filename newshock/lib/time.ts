export function timeAgo(dateStr: string, lang: string): string {
  if (!dateStr) return '';
  const now = Date.now();
  const then = new Date(dateStr.replace(' ', 'T')).getTime();
  const diff = Math.max(0, now - then);

  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return lang === 'zh' ? '刚刚' : 'just now';

  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return lang === 'zh' ? `${minutes}分钟前` : `${minutes}m ago`;

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return lang === 'zh' ? `${hours}小时前` : `${hours}h ago`;

  const days = Math.floor(hours / 24);
  if (days < 30) return lang === 'zh' ? `${days}天前` : `${days}d ago`;

  const months = Math.floor(days / 30);
  return lang === 'zh' ? `${months}个月前` : `${months}mo ago`;
}
