export interface RecognizeImageRequest {
  prompt?: string
  images: string[] // base64 data URLs: "data:image/png;base64,..."
}

export interface GenerateImageRequest {
  prompt: string
  style?: string
}

export interface MultimodalResponse {
  content: string
  duration_ms: number
  model_id: string
}
