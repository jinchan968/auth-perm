'use client'

import { useState } from 'react'
import Image from 'next/image'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { generateActualImage } from '@/lib/api/multimodal'
import { useTenant } from '@/lib/tenant-context'
import type { ImageGenerateRequest } from '@/types/multimodal'
import { Check, Copy, Download, ExternalLink, Image as ImageIcon, Loader2, Wand2 } from 'lucide-react'

const SIZES: Array<{ value: NonNullable<ImageGenerateRequest['size']>; label: string }> = [
  { value: 'auto', label: '自动' },
  { value: '1024x1024', label: '方图' },
  { value: '1536x1024', label: '横图' },
  { value: '1024x1536', label: '竖图' },
]

const QUALITIES: Array<{ value: NonNullable<ImageGenerateRequest['quality']>; label: string }> = [
  { value: 'auto', label: '自动' },
  { value: 'low', label: '低' },
  { value: 'medium', label: '中' },
  { value: 'high', label: '高' },
]

export default function ImageGeneratePanel() {
  const [prompt, setPrompt] = useState('')
  const [size, setSize] = useState<NonNullable<ImageGenerateRequest['size']>>('auto')
  const [quality, setQuality] = useState<NonNullable<ImageGenerateRequest['quality']>>('auto')
  const [imageDataURL, setImageDataURL] = useState('')
  const [b64JSON, setB64JSON] = useState('')
  const [revisedPrompt, setRevisedPrompt] = useState('')
  const [modelID, setModelID] = useState('')
  const [duration, setDuration] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState<'data' | 'base64' | null>(null)
  const { selectedTenantId: tenantId } = useTenant()

  const handleSubmit = async () => {
    if (!prompt.trim()) {
      setError('请输入图片描述')
      return
    }
    setLoading(true)
    setError('')
    setImageDataURL('')
    setB64JSON('')
    setRevisedPrompt('')
    setDuration(0)
    try {
      const res = await generateActualImage(tenantId, {
        prompt: prompt.trim(),
        size,
        quality,
      })
      setImageDataURL(res.image_data_url)
      setB64JSON(res.b64_json)
      setRevisedPrompt(res.revised_prompt || '')
      setDuration(res.duration_ms)
      setModelID(res.model_id)
    } catch (err: any) {
      setError(err.message || '生成图片失败')
    } finally {
      setLoading(false)
    }
  }

  const copyText = async (value: string, kind: 'data' | 'base64') => {
    try {
      await navigator.clipboard.writeText(value)
    } catch {
      const textarea = document.createElement('textarea')
      textarea.value = value
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    setCopied(kind)
    setTimeout(() => setCopied(null), 2000)
  }

  const handleDownload = () => {
    if (!imageDataURL) return
    const link = document.createElement('a')
    link.href = imageDataURL
    link.download = `gpt-image-${Date.now()}.png`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  const handleOpenImage = () => {
    if (!imageDataURL) return
    const win = window.open()
    if (win) {
      win.document.write(`<img src="${imageDataURL}" style="max-width:100%;height:auto;" alt="generated image" />`)
    }
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
        <div className="mb-4 flex items-center gap-2">
          <ImageIcon className="h-5 w-5 text-slate-600" />
          <h3 className="text-lg font-semibold text-slate-900">生成图片</h3>
        </div>
        <p className="mb-4 text-sm text-slate-500">
          输入图片描述，使用后端配置的 GPT Image 模型生成可预览和下载的图片。
        </p>

        <Textarea
          placeholder="例如：清晨的玻璃温室里，一张极简木桌上摆着蓝色陶瓷花瓶，柔和自然光，电影感构图"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          rows={5}
        />

        <div className="mt-4 grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-2 block text-sm font-medium text-slate-700">尺寸</label>
            <div className="flex flex-wrap gap-2">
              {SIZES.map((item) => (
                <Button
                  key={item.value}
                  type="button"
                  variant={size === item.value ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setSize(item.value)}
                >
                  {item.label}
                </Button>
              ))}
            </div>
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium text-slate-700">质量</label>
            <div className="flex flex-wrap gap-2">
              {QUALITIES.map((item) => (
                <Button
                  key={item.value}
                  type="button"
                  variant={quality === item.value ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setQuality(item.value)}
                >
                  {item.label}
                </Button>
              ))}
            </div>
          </div>
        </div>

        <Button onClick={handleSubmit} disabled={loading || !prompt.trim()} className="mt-5 w-full">
          {loading ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              生成中...
            </>
          ) : (
            <>
              <Wand2 className="h-4 w-4 mr-2" />
              生成图片
            </>
          )}
        </Button>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {error}
        </div>
      )}

      {imageDataURL && (
        <div className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
          <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 className="text-lg font-semibold text-slate-900">生成结果</h3>
              <p className="mt-1 text-xs text-slate-400">
                {modelID || 'gpt-image-2'}
                {duration > 0 ? ` · ${(duration / 1000).toFixed(1)}s` : ''}
              </p>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Button size="sm" variant="outline" onClick={handleOpenImage}>
                <ExternalLink className="h-4 w-4 mr-1" />
                查看
              </Button>
              <Button size="sm" variant="outline" onClick={handleDownload}>
                <Download className="h-4 w-4 mr-1" />
                下载
              </Button>
              <Button size="sm" variant="outline" onClick={() => copyText(imageDataURL, 'data')}>
                {copied === 'data' ? <Check className="h-4 w-4 mr-1" /> : <Copy className="h-4 w-4 mr-1" />}
                Data URL
              </Button>
              <Button size="sm" variant="outline" onClick={() => copyText(b64JSON, 'base64')}>
                {copied === 'base64' ? <Check className="h-4 w-4 mr-1" /> : <Copy className="h-4 w-4 mr-1" />}
                Base64
              </Button>
            </div>
          </div>

          <div className="overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
            <Image
              src={imageDataURL}
              alt="生成图片"
              width={1024}
              height={1024}
              unoptimized
              className="mx-auto max-h-[720px] w-full object-contain"
            />
          </div>

          {revisedPrompt && (
            <div className="mt-4 rounded-lg bg-slate-50 p-4 text-sm text-slate-600">
              <div className="mb-1 font-medium text-slate-700">模型优化后的提示词</div>
              {revisedPrompt}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
