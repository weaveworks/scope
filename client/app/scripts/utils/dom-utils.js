/**
 * Cleans up a value to be used as an element ID.
 *
 * Encodes invalid characters to be valid in XHTML and makes it
 * so that it can be reversed by {@link decodeIdAttribute}.
 */
export function encodeIdAttribute(id) {
  return id.replace(/[<>&;]/gm, m => `__u${m.charCodeAt(0)}__`);
}

/**
 * Reverts {@link encodeIdAttribute}.
 */
export function decodeIdAttribute(id) {
  return id.replace(/__u(\d+)__/gm, (m, d) => String.fromCharCode(d));
}
