export type GlobalFilters = {
  window?: "1h" | "24h" | "7d" | "custom";
  from?: string;
  to?: string;
  countryCode?: string;
  region?: string;
  sector?: string;
  industry?: string;
  venue?: string;
  symbol?: string;
  locale?: string;
};

export function filtersToQuery(filters: GlobalFilters = {}) {
  const p = new URLSearchParams();
  Object.entries(filters).forEach(([k, v]) => v && p.set(k, String(v)));
  const q = p.toString();
  return q ? `?${q}` : "";
}
