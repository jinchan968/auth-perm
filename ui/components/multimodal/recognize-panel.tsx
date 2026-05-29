'use client'

import { useState, useRef, useCallback, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { recognizeImage } from '@/lib/api/multimodal'
import { useTenant } from '@/lib/tenant-context'
import ReactMarkdown from 'react-markdown'
import { Upload, X, Loader2, Image as ImageIcon } from 'lucide-react'

interface ImageFile {
  file: File
  preview: string
  dataUrl: string
}

export default function RecognizePanel() {
  const [images, setImages] = useState<ImageFile[]>([])
  const [prompt, setPrompt] = useState('')
  const [result, setResult] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [duration, setDuration] = useState(0)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const { selectedTenantId: tenantId } = useTenant()

  useEffect(() => {
    return () => {
      images.forEach((img) => URL.revokeObjectURL(img.preview))
    }
  }, [])

  const handleFiles = useCallback(async (files: FileList | File[]) => {
    const newImages: ImageFile[] = []
    let hasError = false
    for (const file of Array.from(files)) {
      if (!file.type.startsWith('image/')) continue
      if (file.size > 10 * 1024 * 1024) {
        if (!hasError) {
          setError('图片大小不能超过 10MB')
          hasError = true
        }
        continue
      }
      const dataUrl = await readFileAsDataUrl(file)
      newImages.push({ file, preview: URL.createObjectURL(file), dataUrl })
    }
    setImages((prev) => {
      const combined = [...prev, ...newImages]
      if (combined.length > 5) {
        const excess = combined.slice(5)
        excess.forEach((img) => URL.revokeObjectURL(img.preview))
      }
      return combined.slice(0, 5)
    })
    if (newImages.length > 0 && !hasError) {
      setError('')
    }
  }, [])

  const removeImage = (index: number) => {
    setImages((prev) => {
      URL.revokeObjectURL(prev[index].preview)
      return prev.filter((_, i) => i !== index)
    })
  }

  const handleSubmit = async () => {
    if (images.length === 0) {
      setError('请先上传图片')
      return
    }
    setLoading(true)
    setError('')
    setResult('')
    try {
      const res = await recognizeImage(tenantId, {
        prompt: prompt || undefined,
        images: images.map((img) => img.dataUrl),
      })
      setResult(res.content)
      setDuration(res.duration_ms)
    } catch (err: any) {
      setError(err.message || '识图失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
        <h3 className="text-lg font-semibold text-slate-900 mb-4">上传图片</h3>

        <div
          className="border-2 border-dashed border-slate-300 rounded-lg p-8 text-center cursor-pointer hover:border-blue-400 hover:bg-blue-50/50 transition-colors"
          onClick={() => fileInputRef.current?.click()}
          onDragOver={(e) => e.preventDefault()}
          onDrop={(e) => {
            e.preventDefault()
            handleFiles(e.dataTransfer.files)
          }}
        >
          <Upload className="h-10 w-10 mx-auto text-slate-400 mb-3" />
          <p className="text-sm text-slate-600">点击或拖拽图片到此处上传</p>
          <p className="text-xs text-slate-400 mt-1">支持 JPG、PNG、GIF、WebP，最大 10MB，最多 5 张</p>
        </div>

        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          multiple
          className="hidden"
          onChange={(e) => e.target.files && handleFiles(e.target.files)}
        />

        {images.length > 0 && (
          <div className="flex gap-3 mt-4 flex-wrap">
            {images.map((img, i) => (
              <div key={i} className="relative group">
                <img
                  src={img.preview}
                  alt={`图片 ${i + 1}`}
                  className="h-24 w-24 object-cover rounded-lg border border-slate-200"
                />
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    removeImage(i)
                  }}
                  className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full p-0.5 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <X className="h-3 w-3" />
                </button>
              </div>
            ))}
          </div>
        )}

        <div className="mt-4">
          <Textarea
            placeholder="输入你想了解的问题（可选，留空则默认分析图片内容）"
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            rows={3}
          />
        </div>

        <Button
          onClick={handleSubmit}
          disabled={loading || images.length === 0}
          className="mt-4 w-full"
        >
          {loading ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              分析中...
            </>
          ) : (
            <>
              <ImageIcon className="h-4 w-4 mr-2" />
              开始识图
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
            <h3 className="text-lg font-semibold text-slate-900">分析结果</h3>
            {duration > 0 && (
              <span className="text-xs text-slate-400">{(duration / 1000).toFixed(1)}s</span>
            )}
          </div>
          <div className="prose prose-slate max-w-none">
            <ReactMarkdown>{result}</ReactMarkdown>
          </div>
        </div>
      )}
    </div>
  )
}

function readFileAsDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}
