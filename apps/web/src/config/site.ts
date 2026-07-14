/**
 * Central place for brand and site-wide values. The product name is a
 * working title — change it here and it updates everywhere.
 */
export const siteConfig = {
  name: "Relay",
  tagline: "One link. Total clarity.",
  description:
    "Relay is the focused link workspace for creating clean short URLs, changing destinations without breaking links, and seeing every visit — without the enterprise bloat.",
  /** Display-only short domain used in product mockups. */
  shortHost: "rly.to",
  url: "http://localhost:3000",
  repoUrl: "https://github.com/tapadar13/url-shortener",
  nav: [
    { label: "Product", href: "#product" },
    { label: "How it flows", href: "#how-it-works" },
    { label: "Built different", href: "#architecture" },
  ],
} as const

export type SiteConfig = typeof siteConfig
