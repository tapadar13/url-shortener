/**
 * Central place for brand and site-wide values. The product name is a
 * working title — change it here and it updates everywhere.
 */
export const siteConfig = {
  name: "Relay",
  tagline: "Short links. Clearer signals.",
  description:
    "Relay is a focused link-management platform: create compact short links, manage where they lead, and understand every visit from one workspace.",
  /** Display-only short domain used in product mockups. */
  shortHost: "rly.to",
  url: "http://localhost:3000",
  repoUrl: "https://github.com/tapadar13/url-shortener",
  nav: [
    { label: "Product", href: "#product" },
    { label: "How it works", href: "#how-it-works" },
    { label: "Architecture", href: "#architecture" },
  ],
} as const

export type SiteConfig = typeof siteConfig
