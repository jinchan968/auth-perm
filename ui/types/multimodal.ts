export interface RecognizeImageRequest {
  prompt?: string
  images: string[] // base64 data URLs: "data:image/png;base64,..."
}

export interface GenerateImageRequest {
  prompt: string
  style?: string
}

export interface ImageGenerateRequest {
  prompt: string
  size?: 'auto' | '1024x1024' | '1536x1024' | '1024x1536'
  quality?: 'auto' | 'low' | 'medium' | 'high'
}

export interface MultimodalResponse {
  content: string
  duration_ms: number
  model_id: string
}

export interface ImageGenerateResponse {
  image_data_url: string
  b64_json: string
  revised_prompt?: string
  duration_ms: number
  model_id: string
}
