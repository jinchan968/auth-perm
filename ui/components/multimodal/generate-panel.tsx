'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { generateImage } from '@/lib/api/multimodal'
import { useTenant } from '@/lib/tenant-context'
import ReactMarkdown from 'react-markdown'
import { Check, Copy, Loader2, Wand2 } from 'lucide-react'

const STYLES = [
  { value: 'realistic', label: '写实' },
  { value: 'artistic', label: '艺术' },
  { value: 'cartoon', label: '卡通' },
  { value: 'pixel', label: '像素' },
  { value: 'oil_painting', label: '油画' },
  { value: 'watercolor', label: '水彩' },
]

export default function GeneratePanel() {
  const [prompt, setPrompt] = useState('')
  const [style, setStyle] = useState('realistic')
  const [result, setResult] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [duration, setDuration] = useState(0)
  const [copied, setCopied] = useState(false)
  const { selectedTenantId: tenantId } = useTenant()

  const handleSubmit = async () => {
    if (!prompt.trim()) {
      setError('请输入图片描述')
      return
    }
    setLoading(true)
    setError('')
    setResult('')
    try {
      const res = await generateImage(tenantId, {
        prompt: prompt.trim(),
        style,
      })
      setResult(res.content)
      setDuration(res.duration_ms)
    } catch (err: any) {
      setError(err.message || '生成失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(result)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      const textarea = document.createElement('textarea')
      textarea.value = result
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
        <h3 className="text-lg font-semibold text-slate-900 mb-4">生成绘图提示词</h3>
        <p className="text-sm text-slate-500 mb-4">
          描述你想要的图片，AI 将生成适用于 Midjourney、DALL-E 等工具的英文绘图提示词。
        </p>

        <Textarea
          placeholder="描述你想要生成的图片，例如：一只橘猫坐在窗台上，窗外是雨天的城市夜景"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          rows={4}
        />

        <div className="mt-4">
          <label className="text-sm font-medium text-slate-700 mb-2 block">风格</label>
          <div className="flex flex-wrap gap-2">
            {STYLES.map((s) => (
              <Button
                key={s.value}
                variant={style === s.value ? 'default' : 'outline'}
                size="sm"
                onClick={() => setStyle(s.value)}
              >
                {s.label}
              </Button>
            ))}
          </div>
        </div>

        <Button
          onClick={handleSubmit}
          disabled={loading || !prompt.trim()}
          className="mt-4 w-full"
        >
          {loading ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              生成中...
            </>
          ) : (
            <>
              <Wand2 className="h-4 w-4 mr-2" />
              生成提示词
            </>
          )}
        </Button>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-700 text-sm">
          {error}
        </div>
      )}

      {result && (
        <div className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-slate-900">生成结果</h3>
            <div className="flex items-center gap-3">
              {duration > 0 && (
                <span className="text-xs text-slate-400">{(duration / 1000).toFixed(1)}s</span>
              )}
              <button
                onClick={handleCopy}
                className="flex items-center gap-1 text-xs text-slate-500 hover:text-slate-700 transition-colors"
              >
                {copied ? (
                  <>
                    <Check className="h-3.5 w-3.5" />
                    已复制
                  </>
                ) : (
                  <>
                    <Copy className="h-3.5 w-3.5" />
                    复制
                  </>
                )}
              </button>
            </div>
          </div>

          <div className="prose prose-slate max-w-none">
            <ReactMarkdown>{result}</ReactMarkdown>
          </div>
        </div>
      )}
    </div>
  )
}
