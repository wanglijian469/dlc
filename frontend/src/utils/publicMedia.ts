export const PUBLIC_MEDIA_VERSION = "7";

/**
 * Public media responses are deliberately browser-cacheable. Add the
 * watermark renderer version to local media URLs so a renderer change is
 * visible immediately instead of reusing yesterday's image response.
 */
export function publicMediaURL(src?: string) {
  const value = src?.trim() || "";
  if (!/(?:^|\/)api\/media\/\d+(?:[?#]|$)/.test(value)) return value;

  const hashIndex = value.indexOf("#");
  const base = hashIndex >= 0 ? value.slice(0, hashIndex) : value;
  const hash = hashIndex >= 0 ? value.slice(hashIndex) : "";
  if (/[?&]wm=/.test(base)) {
    return `${base.replace(/([?&]wm=)[^&]*/, `$1${PUBLIC_MEDIA_VERSION}`)}${hash}`;
  }
  return `${base}${base.includes("?") ? "&" : "?"}wm=${PUBLIC_MEDIA_VERSION}${hash}`;
}
