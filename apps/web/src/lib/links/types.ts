export interface ShortLink {
  id: string
  url: string
  shortCode: string
  shortUrl: string
  createdAt: string
  updatedAt: string
  expiresAt?: string
}

export interface ShortLinkListPage {
  items: LinkStats[]
  nextCursor?: string
}

export interface LinkStats extends ShortLink {
  accessCount: number
  lastAccessedAt?: string
}

export interface DailyClicks {
  date: string
  clicks: number
}

export interface LinkAnalytics {
  shortCode: string
  from: string
  to: string
  totalClicks: number
  daily: DailyClicks[]
}

export interface CreateLinkInput {
  url: string
  shortCode?: string
  expiresAt?: string
}
