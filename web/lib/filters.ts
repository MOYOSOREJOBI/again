export type GlobalFilters = {
  time_window?: "now" | "1h" | "24h" | "7d";
  country?: string;
  region?: string;
  sector?: string;
  industry?: string;
  venue?: string;
  asset_class?: string;
  severity?: string;
  trust_state?: string;
};

export function filtersToQuery(filters: GlobalFilters = {}) {
  const p = new URLSearchParams();
  Object.entries(filters).forEach(([k, v]) => v && p.set(k, String(v)));
  const q = p.toString();
  return q ? `?${q}` : "";
}
