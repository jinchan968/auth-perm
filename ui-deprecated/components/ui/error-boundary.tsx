'use client'

import React, { Component, ErrorInfo, ReactNode } from 'react'
import { Button } from '@/components/ui/button'
import { AlertTriangle, RefreshCw } from 'lucide-react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  }

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  private handleReload = () => {
    window.location.reload()
  }

  public render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div className="flex flex-col items-center justify-center min-h-[400px] p-8">
          <div className="max-w-md w-full text-center">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-red-100 flex items-center justify-center">
              <AlertTriangle className="w-8 h-8 text-red-600" />
            </div>
            
            <h2 className="text-xl font-semibold text-gray-900 mb-2">
              页面出错了
            </h2>
            
            <p className="text-gray-600 mb-6">
              抱歉，页面遇到了意外错误。这可能是暂时性的问题。
            </p>
            
            <div className="flex gap-3 justify-center">
              <Button onClick={this.handleReload}>
                <RefreshCw className="w-4 h-4 mr-2" />
                刷新页面
              </Button>
              
              <Button
                variant="outline"
                onClick={() => window.history.back()}
              >
                返回上一页
              </Button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
