'use client'

import { useState, useEffect, useCallback, useMemo } from 'react'
import {
  Sparkles, Loader2, Clock, AlertCircle, Brain,
  History, RotateCcw, MessageSquare, Maximize2, Minimize2,
} from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import { showError, showSuccess } from '@/lib/toast'
import * as journalApi from '@/lib/api/journal'
import type { AIPredictionResult, AIPredictionHistoryItem, AIPredictionModelsResponse } from '@/types/journal'

const DEFAULT_MODELS = [
  'glm-5.1',
  'kimi-k2.6',
  'mimo-v2.5-pro',
  'deepseek-v4-pro',
  'qwen3.6-plus',
]

const REPLACEABLE_OPTIONS: { id: string; name: string }[] = [
  { id: 'qwen3.6-plus', name: 'Qwen3.6 Plus（默认）' },
  { id: 'minimax-m2.7', name: 'MiniMax M2.7' },
  { id: 'minimax-m2.5', name: 'MiniMax M2.5' },
  { id: 'deepseek-v4-flash', name: 'DeepSeek V4 Flash' },
  { id: 'glm-5', name: 'GLM-5' },
  { id: 'kimi-k2.5', name: 'Kimi K2.5' },
]

const REASONING_OPTIONS: { value: 'low' | 'medium' | 'high'; label: string }[] = [
  { value: 'low', label: '轻度思考' },
  { value: 'medium', label: '中度思考' },
  { value: 'high', label: '深度思考' },
]

const CARD_ACCENTS = [
  'from-blue-500/10 to-blue-600/5 border-blue-200/50',
  'from-indigo-500/10 to-indigo-600/5 border-indigo-200/50',
  'from-emerald-500/10 to-emerald-600/5 border-emerald-200/50',
  'from-amber-500/10 to-amber-600/5 border-amber-200/50',
  'from-rose-500/10 to-rose-600/5 border-rose-200/50',
]

const MODEL_DISPLAY_NAMES: Record<string, string> = {
  'glm-5.1': 'GLM-5.1',
  'kimi-k2.6': 'Kimi K2.6',
  'mimo-v2.5-pro': 'MiMo-V2.5-Pro',
  'deepseek-v4-pro': 'DeepSeek V4 Pro',
  'deepseek-v4-flash': 'DeepSeek V4 Flash',
  'qwen3.6-plus': 'Qwen3.6 Plus',
  'minimax-m2.7': 'MiniMax M2.7',
  'minimax-m2.5': 'MiniMax M2.5',
  'glm-5': 'GLM-5',
  'kimi-k2.5': 'Kimi K2.5',
}

interface AIPredictionPanelProps {
  tenantId: string | null
  tenantReady: boolean
}

export default function AIPredictionPanel({ tenantId, tenantReady }: AIPredictionPanelProps) {
  const [question, setQuestion] = useState('')
  const [systemPrompt, setSystemPrompt] = useState('你是一个有帮助的助手。')
  const [showPromptInput, setShowPromptInput] = useState(false)
  const [cardModels, setCardModels] = useState<string[]>([...DEFAULT_MODELS])
  const [reasoningMode, setReasoningMode] = useState<'low' | 'medium' | 'high'>('low')
  const [results, setResults] = useState<(AIPredictionResult & { loading: boolean })[] | null>(null)
  const [predicting, setPredicting] = useState(false)
  const [history, setHistory] = useState<AIPredictionHistoryItem[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)
  const [showHistory, setShowHistory] = useState(false)
  const [historyPage, setHistoryPage] = useState(1)
  const [historyTotal, setHistoryTotal] = useState(0)
  const [modelConfig, setModelConfig] = useState<AIPredictionModelsResponse | null>(null)
  const [expandedIdx, setExpandedIdx] = useState<number | null>(null)
  const [quotas, setQuotas] = useState<Record<string, number>>({})
  const [dailyLimit, setDailyLimit] = useState(10)

  // Load model config from backend on mount
  useEffect(() => {
    if (!tenantId || modelConfig) return
    journalApi.getAIPredictionModels(tenantId)
      .then(setModelConfig)
      .catch(() => {})
  }, [tenantId, modelConfig])

  // Load quotas from backend on mount
  useEffect(() => {
    if (!tenantId) return
    journalApi.getAIPredictionQuotas(tenantId)
      .then((res) => {
        setQuotas(res.remaining)
        setDailyLimit(res.daily_limit)
      })
      .catch(() => {})
  }, [tenantId])

  const refreshQuotas = useCallback(() => {
    if (!tenantId) return
    journalApi.getAIPredictionQuotas(tenantId)
      .then((res) => {
        setQuotas(res.remaining)
        setDailyLimit(res.daily_limit)
      })
      .catch(() => {})
  }, [tenantId])

  const effectiveDefaults = modelConfig?.defaults || DEFAULT_MODELS
  const effectiveDisplayNames = useMemo(() => {
    const map: Record<string, string> = { ...MODEL_DISPLAY_NAMES }
    if (modelConfig?.all) {
      modelConfig.all.forEach((m) => { map[m.id] = m.name })
    }
    return map
  }, [modelConfig])

  const allModelOptions = useMemo(() => {
    if (modelConfig?.all) return modelConfig.all
    const seen = new Set<string>()
    const result: { id: string; name: string }[] = []
    for (const id of effectiveDefaults) {
      if (!seen.has(id)) { seen.add(id); result.push({ id, name: effectiveDisplayNames[id] || id }) }
    }
    for (const opt of REPLACEABLE_OPTIONS) {
      if (!seen.has(opt.id)) { seen.add(opt.id); result.push(opt) }
    }
    return result
  }, [modelConfig, effectiveDefaults, effectiveDisplayNames])

  // Sync cardModels when backend defaults load (only if user hasn't customized)
  useEffect(() => {
    if (!modelConfig?.defaults) return
    setCardModels((prev) => {
      if (prev.every((m, i) => m === DEFAULT_MODELS[i])) {
        return [...modelConfig.defaults!]
      }
      return prev
    })
  }, [modelConfig])

  const currentModels = cardModels

  const allQuotasExhausted = useMemo(() => {
    return currentModels.every((m) => (quotas[m] ?? dailyLimit) <= 0)
  }, [currentModels, quotas, dailyLimit])

  const resetResults = useCallback(() => {
    setResults(
      currentModels.map((modelId) => ({
        model_id: modelId,
        model_name: effectiveDisplayNames[modelId] || modelId,
        content: '',
        duration_ms: 0,
        loading: false,
      }))
    )
  }, [currentModels, effectiveDisplayNames])

  // Initialize results once on mount; do NOT auto-reset when cardModels change
  // to avoid wiping all card contents when switching a single model.
  useEffect(() => {
    if (results === null) {
      resetResults()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const fetchHistory = useCallback(async (page: number) => {
    if (!tenantId) return
    setHistoryLoading(true)
    try {
      const res = await journalApi.listAIPredictions({ tenant_id: tenantId, page, page_size: 10 })
      setHistory((prev) => page === 1 ? res.data : [...prev, ...res.data])
      setHistoryTotal(res.total)
    } catch (e: unknown) {
      // silent
    } finally {
      setHistoryLoading(false)
    }
  }, [tenantId])

  const handlePredict = async () => {
    if (!question.trim()) {
      showError('请输入问题')
      return
    }
    if (!tenantId) return

    setPredicting(true)
    setResults(
      currentModels.map((modelId) => ({
        model_id: modelId,
        model_name: effectiveDisplayNames[modelId] || modelId,
        content: '',
        duration_ms: 0,
        loading: true,
      }))
    )

    try {
      const res = await journalApi.createAIPrediction({
        tenant_id: tenantId,
        question: question.trim(),
        system_prompt: systemPrompt || '你是一个有帮助的助手。',
        models: currentModels,
        reasoning_mode: reasoningMode,
      })
      setResults(
        res.results.map((r) => ({ ...r, loading: false }))
      )
      setExpandedIdx(null)
      showSuccess('预测完成')
      setHistoryPage(1)
      fetchHistory(1)
      refreshQuotas()
    } catch (e: unknown) {
      showError(e instanceof Error ? e.message : '预测失败')
      setResults(
        currentModels.map((modelId) => ({
          model_id: modelId,
          model_name: effectiveDisplayNames[modelId] || modelId,
          content: '',
          duration_ms: 0,
          loading: false,
          error: '请求失败',
        }))
      )
    } finally {
      setPredicting(false)
    }
  }

  const handleRetrySingle = useCallback(async (idx: number) => {
    if (!tenantId || !question.trim()) return

    const modelId = cardModels[idx]
    const modelName = effectiveDisplayNames[modelId] || modelId
    setResults((prev) =>
      prev?.map((r, ri) =>
        ri === idx ? { ...r, model_id: modelId, model_name: modelName, content: '', error: undefined, duration_ms: 0, loading: true } : r
      ) ?? null
    )

    try {
      const res = await journalApi.createAIPrediction({
        tenant_id: tenantId,
        question: question.trim(),
        system_prompt: systemPrompt || '你是一个有帮助的助手。',
        models: [modelId],
        reasoning_mode: reasoningMode,
      })
      const result = res.results[0]
      if (!result) {
        showError('未返回结果')
        setResults((prev) =>
          prev?.map((r, ri) => (ri === idx ? { ...r, error: '未返回结果', loading: false } : r)) ?? null
        )
        return
      }
      setResults((prev) =>
        prev?.map((r, ri) => (ri === idx ? { ...result, loading: false } : r)) ?? null
      )
      refreshQuotas()
    } catch (e: unknown) {
      showError(e instanceof Error ? e.message : '重试失败')
      setResults((prev) =>
        prev?.map((r, ri) => (ri === idx ? { ...r, error: '请求失败', loading: false } : r)) ?? null
      )
    }
  }, [tenantId, question, cardModels, effectiveDisplayNames, systemPrompt, reasoningMode, refreshQuotas])

  const handleCardModelChange = useCallback((idx: number, v: string) => {
    if (cardModels.some((m, i) => i !== idx && m === v)) {
      showError(`模型 "${effectiveDisplayNames[v] || v}" 已被其他卡片选中`)
      return
    }
    setCardModels((prev) => {
      const next = [...prev]
      next[idx] = v
      return next
    })
    setResults((prev) =>
      prev?.map((r, ri) =>
        ri === idx ? { ...r, model_id: v, model_name: effectiveDisplayNames[v] || v, content: '', error: undefined, duration_ms: 0 } : r
      ) ?? null
    )
  }, [cardModels, effectiveDisplayNames])

  const loadHistoryItem = async (item: AIPredictionHistoryItem) => {
    if (!tenantId) return
    try {
      const detail = await journalApi.getAIPrediction(item.id, tenantId)
      setQuestion(detail.question)
      setSystemPrompt(detail.system_prompt || '你是一个有帮助的助手。')
      setCardModels(detail.model_snapshot.length > 0 ? detail.model_snapshot : [...DEFAULT_MODELS])
      setResults(
        detail.results.map((r) => ({ ...r, loading: false }))
      )
      setExpandedIdx(null)
      setShowHistory(false)
    } catch {
      showError('加载历史记录失败')
    }
  }

  return (
    <div className="space-y-4">
      {/* Input Area */}
      <Card>
        <CardContent className="pt-5 space-y-3">
          <div className="flex items-start gap-2">
            <Brain className="h-5 w-5 text-primary mt-2 shrink-0" />
            <div className="flex-1 space-y-3">
              <textarea
                    className="w-full min-h-[90px] rounded-lg border-2 border-slate-200 bg-white px-4 py-3 text-[15px] leading-relaxed placeholder:text-slate-300 focus:outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all duration-200 resize-y"
                    value={question}
                    onChange={(e) => setQuestion(e.target.value)}
                    placeholder="输入你想问的问题，多个AI将同时为你回答..."
                    disabled={predicting}
                  />
                  <div className="flex items-center gap-2">
                    <Button
                      size="sm"
                      variant="ghost"
                      className="text-xs text-slate-400 h-7"
                      onClick={() => setShowPromptInput(!showPromptInput)}
                    >
                      <MessageSquare className="h-3 w-3 mr-1" />
                      {showPromptInput ? '收起' : '自定义'} System Prompt
                    </Button>
                    <div className="flex items-center gap-2 ml-auto">
                      <span className="text-xs text-slate-400">思考程度：</span>
                      <Select
                        value={reasoningMode}
                        onValueChange={(v) => setReasoningMode(v as 'low' | 'medium' | 'high')}
                        disabled={predicting}
                      >
                        <SelectTrigger className="h-8 w-28 text-xs">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {REASONING_OPTIONS.map((opt) => (
                            <SelectItem key={opt.value} value={opt.value}>
                              {opt.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  {showPromptInput && (
                    <textarea
                      className="w-full min-h-[60px] rounded-lg border-2 border-slate-200 bg-slate-50 px-4 py-2.5 text-[13px] leading-relaxed placeholder:text-slate-300 focus:outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all duration-200 resize-y"
                      value={systemPrompt}
                      onChange={(e) => setSystemPrompt(e.target.value)}
                      placeholder="自定义 System Prompt..."
                      disabled={predicting}
                    />
                  )}
                </div>
              </div>
              <div className="flex items-center justify-between pt-1">
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-xs text-slate-400"
                  onClick={() => {
                    setShowHistory(!showHistory)
                    if (!showHistory && history.length === 0 && tenantId) {
                      fetchHistory(1)
                    }
                  }}
                >
                  <History className="h-3.5 w-3.5 mr-1" />
                  {showHistory ? '收起历史' : '历史记录'}
                  {historyTotal > 0 && (
                    <span className="ml-1 text-primary">({historyTotal})</span>
                  )}
                </Button>
                <Button
                  onClick={handlePredict}
                  disabled={predicting || !question.trim() || !tenantReady || allQuotasExhausted}
                  variant="default"
                >
                  {predicting ? (
                    <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />
                  ) : (
                    <Sparkles className="h-4 w-4 mr-1.5" />
                  )}
                  {predicting ? 'AI 思考中...' : '推理'}
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* History Section */}
          {showHistory && (
            <Card className="mt-3 border-slate-200/50">
              <CardHeader className="py-3">
                <CardTitle className="text-sm font-medium text-slate-600 flex items-center">
                  <History className="h-4 w-4 mr-1.5" />
                  历史记录
                </CardTitle>
              </CardHeader>
              <CardContent className="py-0 pb-3 max-h-52 overflow-y-auto space-y-1">
                {historyLoading && history.length === 0 ? (
                  <div className="flex justify-center py-4">
                    <Loader2 className="h-5 w-5 animate-spin text-slate-300" />
                  </div>
                ) : history.length === 0 ? (
                  <p className="text-xs text-slate-400 text-center py-3">暂无历史记录</p>
                ) : (
                  history.map((item) => (
                    <button
                      key={item.id}
                      className="w-full text-left px-3 py-2 rounded-lg hover:bg-slate-50 transition-colors flex items-center gap-2 group"
                      onClick={() => loadHistoryItem(item)}
                    >
                      <MessageSquare className="h-3.5 w-3.5 text-slate-300 shrink-0" />
                      <span className="text-sm text-slate-600 truncate flex-1">{item.question}</span>
                      <span className="text-xs text-slate-300 shrink-0">{new Date(item.created_at).toLocaleDateString()}</span>
                    </button>
                  ))
                )}
                {history.length < historyTotal && (
<Button
                  variant="ghost"
                  size="sm"
                  className="w-full text-xs text-slate-400"
                  onClick={() => {
                    const nextPage = historyPage + 1
                    setHistoryPage(nextPage)
                    fetchHistory(nextPage)
                  }}
                  disabled={historyLoading}
                >
                  {historyLoading ? '加载中...' : '加载更多'}
                </Button>
                )}
              </CardContent>
            </Card>
          )}

          {/* Results Grid */}
          {results && (
            <div className="mt-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-3">
              {results.map((result, idx) => (
                <Card
                  key={result.model_id}
                  className={`border overflow-hidden transition-all duration-300 ${
                    CARD_ACCENTS[idx] || 'border-slate-200/50'
                  } ${result.loading ? 'animate-pulse' : 'animate-in fade-in slide-in-from-bottom-2'} ${expandedIdx === idx ? 'col-span-full' : ''}`}
                  style={result.loading ? {} : { animationDelay: `${idx * 100}ms` }}
                >
                  <CardHeader className="py-2.5 px-4 border-b">
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-1.5 min-w-0">
                        <span className={`w-2 h-2 rounded-full shrink-0 ${
                          ['bg-blue-400', 'bg-indigo-400', 'bg-emerald-400', 'bg-amber-400', 'bg-rose-400'][idx]
                        }`} />
                        <Select
                          value={cardModels[idx]}
                          onValueChange={(v) => handleCardModelChange(idx, v)}
                          disabled={predicting || result.loading}
                        >
                          <SelectTrigger className="h-6 border-0 bg-transparent text-sm font-semibold p-0 m-0 [&>span]:truncate max-w-[140px] focus:ring-0">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {allModelOptions.map((opt) => (
                              <SelectItem key={opt.id} value={opt.id}>
                                {opt.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <span className={`text-[10px] ml-0.5 ${
                          (quotas[cardModels[idx]] ?? dailyLimit) <= 0
                            ? 'text-red-400 font-medium'
                            : (quotas[cardModels[idx]] ?? dailyLimit) <= 3
                              ? 'text-amber-500'
                              : 'text-slate-400'
                        }`}>
                          {quotas[cardModels[idx]] ?? dailyLimit}/{dailyLimit}
                        </span>
                      </div>
                      <div className="flex items-center gap-1 shrink-0">
                        {result.loading ? (
                          <Loader2 className="h-3 w-3 animate-spin text-primary" />
                        ) : (
                          <>
                            {result.duration_ms > 0 && (
                              <span className="text-[10px] text-slate-300 flex items-center">
                                <Clock className="h-3 w-3 mr-0.5" />
                                {(result.duration_ms / 1000).toFixed(1)}s
                              </span>
                            )}
                            <button
                              onClick={() => handleRetrySingle(idx)}
                              disabled={predicting || (quotas[cardModels[idx]] ?? dailyLimit) <= 0}
                              className={`transition-colors ${(quotas[cardModels[idx]] ?? dailyLimit) <= 0 ? 'text-slate-200 cursor-not-allowed' : 'text-slate-300 hover:text-primary'}`}
                              title={(quotas[cardModels[idx]] ?? dailyLimit) <= 0 ? '今日次数已用完' : '重新推理'}
                            >
                              <RotateCcw className="h-3 w-3" />
                            </button>
                            <button
                              onClick={() => setExpandedIdx(expandedIdx === idx ? null : idx)}
                              className="text-slate-300 hover:text-primary transition-colors"
                              title={expandedIdx === idx ? '收起' : '展开'}
                            >
                              {expandedIdx === idx ? <Minimize2 className="h-3 w-3" /> : <Maximize2 className="h-3 w-3" />}
                            </button>
                          </>
                        )}
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className={`p-4 overflow-x-auto ${expandedIdx === idx ? 'max-h-[70vh] overflow-y-auto' : ''}`}>
                    {result.loading ? (
                      <div className="space-y-2">
                        <Skeleton className="h-3 w-full" />
                        <Skeleton className="h-3 w-5/6" />
                        <Skeleton className="h-3 w-4/6" />
                        <Skeleton className="h-3 w-3/6" />
                      </div>
                    ) : result.error ? (
                      <div className="flex items-start gap-2 text-red-400">
                        <AlertCircle className="h-4 w-4 mt-0.5 shrink-0" />
                        <p className="text-xs text-red-400">{result.error}</p>
                      </div>
                    ) : (
                      <div className="prose prose-sm max-w-none prose-p:my-1 prose-headings:my-2 prose-code:px-1 prose-code:bg-slate-100 prose-code:rounded">
                        <ReactMarkdown remarkPlugins={[remarkGfm]}>
                          {result.content}
                        </ReactMarkdown>
                      </div>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
    </div>
  )
}
