"use client"

/**
 * In-page navigation link that scrolls to the target section without
 * writing the hash into the URL. The href is kept as a fallback for
 * non-JS environments and for open-in-new-tab.
 */
export function AnchorLink({
  href,
  onClick,
  ...props
}: React.ComponentProps<"a">) {
  const handleClick = (event: React.MouseEvent<HTMLAnchorElement>) => {
    onClick?.(event)
    if (
      href?.startsWith("#") &&
      !event.defaultPrevented &&
      !event.metaKey &&
      !event.ctrlKey
    ) {
      event.preventDefault()
      document.querySelector(href)?.scrollIntoView()
    }
  }

  return <a href={href} onClick={handleClick} {...props} />
}
