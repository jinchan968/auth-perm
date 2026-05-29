import { apiClient } from './client'
import type { RecognizeImageRequest, GenerateImageRequest, MultimodalResponse } from '@/types/multimodal'

const BASE = '/multimodal'

export function recognizeImage(
  tenantId: string,
  data: RecognizeImageRequest
): Promise<MultimodalResponse> {
  return apiClient.post<MultimodalResponse>(
    `${BASE}/recognize?tenant_id=${tenantId}`,
    { prompt: data.prompt, images: data.images }
  )
}

export function generateImage(
  tenantId: string,
  data: GenerateImageRequest
): Promise<MultimodalResponse> {
  return apiClient.post<MultimodalResponse>(
    `${BASE}/generate?tenant_id=${tenantId}`,
    { prompt: data.prompt, style: data.style }
  )
}
