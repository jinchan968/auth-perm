import { apiClient } from './client'
import type {
  GenerateImageRequest,
  ImageGenerateRequest,
  ImageGenerateResponse,
  MultimodalResponse,
  RecognizeImageRequest,
} from '@/types/multimodal'

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

export function generateActualImage(
  tenantId: string,
  data: ImageGenerateRequest
): Promise<ImageGenerateResponse> {
  return apiClient.post<ImageGenerateResponse>(
    `${BASE}/image-generate?tenant_id=${tenantId}`,
    { prompt: data.prompt, size: data.size, quality: data.quality }
  )
}
