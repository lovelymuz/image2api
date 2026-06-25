// Credit display helpers. Credits and model prices are stored server-side as
// integer 积分 (points) — the single unit across the whole app. These helpers
// only format for display; they do NOT convert units.
/** Round to an integer 积分 value. */
export function points(value) {
  return Math.round(Number(value || 0))
}

/** "<n> 积分" label with thousands separators. */
export function pointsLabel(value) {
  return points(value).toLocaleString('en-US') + ' 积分'
}
