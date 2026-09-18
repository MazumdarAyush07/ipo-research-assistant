export interface IPO {
  id: number;
  company_name: string;
  exchange_type: string;
  sector: string;
  issue_price: number;
  issue_size: number;
  open_date: string;
  close_date: string;
  listing_date: string;
  final_score?: number | null;
}

export interface Score {
  TotalScore: number;
  Recommendation: string;
  FinancialsScore: number;
  FinancialsReason: string;
  ValuationScore: number;
  ValuationReason: string;
  PromoterScore: number;
  PromoterReason: string;
  IndustryScore: number;
  IndustryReason: string;
  RiskScore: number;
  RiskReason: string;
  SubscriptionScore: number;
  SubscriptionReason: string;
  GmpScore: number;
  GmpReason: string;
}

export interface Financials {
  Year: string;
  Revenue: number;
  Pat: number;
  Ebitda: number;
  Debt: number;
  NetWorth: number;
}

export interface Peer {
  ticker: string;
  name: string;
  pe: number;
  pb: number;
  market_cap: number;
}

export interface GMP {
  gmp_price: number;
  est_listing: number;
  premium_percent: number;
  updated_at: string;
}

export interface Subscriptions {
  qib_demand: number;
  nii_demand: number;
  retail_demand: number;
  total_demand: number;
  updated_at: string;
}

export interface AIAnalysis {
  red_flags: string;
  management_assumptions: string;
  unique_risks: string;
}

export interface ParsingAuditStats {
  total_ipos: number;
  total_financials: number;
  missing_parsing: number;
}

const getApiBaseUrl = () => {
  if (typeof window === "undefined") {
    // Server-side rendering (SSR) - running inside Docker container
    return process.env.NEXT_SERVER_API_URL || "http://backend:8080/api";
  }
  // Client-side rendering (CSR) - running in browser
  return process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";
};

const API_BASE_URL = getApiBaseUrl();

export const adminApi = {
  triggerAllDownloads: async () => {
    const res = await fetch(`${API_BASE_URL}/admin/downloads/trigger`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to trigger downloads');
    return res.json();
  },

  getParsingAudit: async (): Promise<ParsingAuditStats> => {
    const res = await fetch(`${API_BASE_URL}/admin/parsing-audit`);
    if (!res.ok) throw new Error('Failed to fetch parsing audit');
    return res.json();
  }
};

export async function fetchIPOs(): Promise<IPO[]> {
  let rawIPOs = [];
  try {
    const res = await fetch(`${API_BASE_URL}/ipos`, { next: { revalidate: 60 } });
    if (res.ok) {
      const data = await res.json();
      rawIPOs = data.data || [];
    }
  } catch (error) {
    console.warn("Failed to fetch IPOs (backend might be down during build):", error);
  }
  
  const getString = (val: any): string => {
    if (typeof val === 'string') return val;
    if (val && typeof val === 'object' && 'String' in val) return val.String || "";
    return "";
  };

  return rawIPOs.map((raw: any) => {
    return {
      id: raw.id || raw.ID,
      company_name: getString(raw.Name) || getString(raw.company_name) || "",
      ticker_symbol: getString(raw.TickerSymbol) || getString(raw.ticker_symbol) || "",
      exchange_type: getString(raw.ExchangeType) || getString(raw.exchange_type) || "",
      sector: getString(raw.Sector) || getString(raw.sector) || "",
      issue_price: parseFloat(getString(raw.PriceBandHigh) || getString(raw.PriceBandLow) || getString(raw.issue_price) || "0"),
      issue_size: raw.issue_size || 0,
      open_date: getString(raw.open_date) || (raw.OpenDate?.Valid ? raw.OpenDate.Time : ""),
      close_date: getString(raw.close_date) || (raw.CloseDate?.Valid ? raw.CloseDate.Time : ""),
      listing_date: getString(raw.listing_date) || (raw.ListingDate?.Valid ? raw.ListingDate.Time : ""),
      final_score: raw.final_score ? parseFloat(getString(raw.final_score) || (raw.FinalScore?.Valid ? raw.FinalScore.String : "0")) : (raw.FinalScore?.Valid ? parseFloat(raw.FinalScore.String) : null)
    };
  });
}

export async function fetchIPODetails(id: string): Promise<{
  score: Score | null;
  financials: Financials[];
  peers: Peer[];
  ipoPe: number | null;
  gmp: GMP | null;
  subscriptions: Subscriptions | null;
  analysis: AIAnalysis | null;
}> {
  const [
    scoreRes,
    financialsRes,
    peersRes,
    gmpRes,
    subsRes,
    analysisRes
  ] = await Promise.all([
    fetch(`${API_BASE_URL}/ipos/${id}/score`, { next: { revalidate: 60 } }).catch(() => null),
    fetch(`${API_BASE_URL}/ipos/${id}/financials`, { next: { revalidate: 3600 } }).catch(() => null),
    fetch(`${API_BASE_URL}/ipos/${id}/peers`, { next: { revalidate: 3600 } }).catch(() => null),
    fetch(`${API_BASE_URL}/ipos/${id}/gmp`, { next: { revalidate: 60 } }).catch(() => null),
    fetch(`${API_BASE_URL}/ipos/${id}/subscriptions`, { next: { revalidate: 60 } }).catch(() => null),
    fetch(`${API_BASE_URL}/ipos/${id}/analysis`, { next: { revalidate: 3600 } }).catch(() => null),
  ]);

  const scoreData = scoreRes?.ok ? await scoreRes.json().catch(() => null) : null;
  const financialsData = financialsRes?.ok ? await financialsRes.json().catch(() => null) : null;
  const peersData = peersRes?.ok ? await peersRes.json().catch(() => null) : null;
  const gmpData = gmpRes?.ok ? await gmpRes.json().catch(() => null) : null;
  const subsData = subsRes?.ok ? await subsRes.json().catch(() => null) : null;
  const analysisData = analysisRes?.ok ? await analysisRes.json().catch(() => null) : null;

  const getString = (val: any): string => {
    if (typeof val === 'string') return val;
    if (val && typeof val === 'object' && 'String' in val) return val.String || "";
    return "";
  };

  const mapScore = (raw: any): Score | null => {
    if (!raw) return null;
    return {
      TotalScore: parseFloat(getString(raw.FinalScore) || getString(raw.TotalScore) || raw.FinalScore?.toString() || raw.TotalScore?.toString() || "0"),
      Recommendation: getString(raw.Recommendation) || "",
      FinancialsScore: parseFloat(getString(raw.FinancialsScore) || raw.FinancialsScore?.toString() || "0"),
      FinancialsReason: getString(raw.FinancialsReason) || "",
      ValuationScore: parseFloat(getString(raw.ValuationScore) || raw.ValuationScore?.toString() || "0"),
      ValuationReason: getString(raw.ValuationReason) || "",
      PromoterScore: parseFloat(getString(raw.PromoterScore) || raw.PromoterScore?.toString() || "0"),
      PromoterReason: getString(raw.PromoterReason) || "",
      IndustryScore: parseFloat(getString(raw.IndustryScore) || raw.IndustryScore?.toString() || "0"),
      IndustryReason: getString(raw.IndustryReason) || "",
      RiskScore: parseFloat(getString(raw.RiskScore) || raw.RiskScore?.toString() || "0"),
      RiskReason: getString(raw.RiskReason) || "",
      SubscriptionScore: parseFloat(getString(raw.SubscriptionScore) || raw.SubscriptionScore?.toString() || "0"),
      SubscriptionReason: getString(raw.SubscriptionReason) || "",
      GmpScore: parseFloat(getString(raw.GmpScore) || raw.GmpScore?.toString() || "0"),
      GmpReason: getString(raw.GmpReason) || "",
    };
  };

  const mapFinancial = (raw: any): Financials => {
    return {
      Year: raw.Year?.toString() || "",
      Revenue: parseFloat(getString(raw.Revenue) || getString(raw.revenue) || "0"),
      Pat: parseFloat(getString(raw.Pat) || getString(raw.pat) || "0"),
      Ebitda: parseFloat(getString(raw.Ebitda) || getString(raw.ebitda) || "0"),
      Debt: parseFloat(getString(raw.TotalDebt) || getString(raw.total_debt) || "0"),
      NetWorth: parseFloat(getString(raw.Equity) || getString(raw.equity) || "0"),
    };
  };

  const mapPeer = (raw: any): Peer => {
    return {
      ticker: getString(raw.TickerSymbol) || getString(raw.ticker) || "",
      name: getString(raw.Name) || getString(raw.name) || "",
      pe: raw.PeRatio || raw.pe || 0,
      pb: raw.PbRatio || raw.pb || 0,
      market_cap: raw.MarketCap || raw.market_cap || 0,
    };
  };

  const mapAnalysis = (raw: any): AIAnalysis | null => {
    if (!raw) return null;
    return {
      red_flags: getString(raw.RedFlags) || getString(raw.red_flags) || "",
      management_assumptions: getString(raw.ManagementAssumptions) || getString(raw.management_assumptions) || "",
      unique_risks: getString(raw.KeyRisks) || getString(raw.key_risks) || "",
    };
  };

  const mapGmp = (raw: any): GMP | null => {
    if (!raw) return null;
    return {
      gmp_price: parseFloat(getString(raw.GmpAmount) || getString(raw.gmp_price) || "0"),
      est_listing: 0,
      premium_percent: parseFloat(getString(raw.PremiumPercent) || getString(raw.premium_percent) || "0"),
      updated_at: raw.RecordedAt?.Time || getString(raw.updated_at) || "",
    };
  };

  const mapSubscriptions = (rawArray: any): Subscriptions | null => {
    if (!rawArray || !Array.isArray(rawArray)) return null;
    
    const subs: Subscriptions = {
      qib_demand: 0,
      nii_demand: 0,
      retail_demand: 0,
      total_demand: 0,
      updated_at: "",
    };

    rawArray.forEach(raw => {
      const val = parseFloat(getString(raw.TimesSubscribed) || "0");
      const cat = raw.Category?.toLowerCase();
      if (cat === "qib") subs.qib_demand = val;
      if (cat === "nii") subs.nii_demand = val;
      if (cat === "retail") subs.retail_demand = val;
      if (cat === "total") subs.total_demand = val;
      if (!subs.updated_at) subs.updated_at = raw.RecordedAt?.Time || "";
    });

    return subs;
  };

  return {
    score: mapScore(scoreData?.data),
    financials: (financialsData?.data || []).map(mapFinancial),
    peers: (peersData?.data?.peers || peersData?.data || []).map(mapPeer),
    ipoPe: peersData?.data?.ipo_pe || null,
    gmp: mapGmp(gmpData?.data),
    subscriptions: mapSubscriptions(subsData?.data),
    analysis: mapAnalysis(analysisData?.data),
  };
}
