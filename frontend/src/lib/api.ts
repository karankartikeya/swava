// Types mirror cmd/server's JSON wire format exactly — see
// internal/decide.Result, internal/purchase.{Product,TrustGateResult,SessionResult}.

export interface DemoWallet {
  address: string;
  label: string;
  agent_role: string;
  description: string;
}

export interface ProcurementPolicy {
  agent_label: string;
  category_cap_rupees: number;
  blocked_keywords?: string[];
}

export interface Transaction {
  tx_hash: string;
  from: string;
  to: string;
  value: string;
  timestamp: string;
  is_token_transfer: boolean;
  chain: string;
}

export interface AgentProfile {
  address: string;
  agent_role: string;
  known: boolean;
  raw_score: number;
  tier: string;
  policy?: ProcurementPolicy;
  transactions: Transaction[];
}

export interface Product {
  description: string;
  variant_id: string;
  unit_price_decimal: string;
  unit_price_rupees: number;
  currency: string;
  merchant_name: string;
  merchant_url: string;
  merchant_mcp_url: string;
  merchant_country: string;
}

export interface DecideStep {
  merchant_name: string;
  search_term: string;
  task_match: boolean;
  configured: boolean;
  reachable: boolean;
  error?: string;
  found_title?: string;
  found_price_rupees?: number;
  confirmed: boolean;
  under_budget: boolean;
  rejected: boolean;
  reject_reason?: string;
}

export interface DecideResult {
  task: string;
  budget_rupees: number;
  steps: DecideStep[];
  chosen_product?: Product;
  reasoning: string;
  wallet_role: string;
}

export type Decision = "auto_approve" | "human_review" | "blocked";

export interface TrustGateResult {
  wallet_address: string;
  known: boolean;
  raw_score: number;
  tier: string;
  normalized_score: number;
  spend_limit_rupees: number;
  decision: Decision;
  reason: string;
  policy?: ProcurementPolicy;
}

export interface PurchaseResult {
  trust_gate: TrustGateResult;
  prava_session_id?: string;
  prava_order_id?: string;
  iframe_url?: string;
  expires_at?: string;
  wallet_role: string;
  sandbox_note?: string;
}

export const BASE = import.meta.env.VITE_API_BASE ?? "/api";

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? `request failed: ${res.status}`);
  }
  return res.json();
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? `request failed: ${res.status}`);
  }
  return res.json();
}

export const api = {
  wallets: () => get<DemoWallet[]>("/wallets"),
  decide: (task: string, budgetRupees: number) =>
    post<DecideResult>("/decide", { task, budget_rupees: budgetRupees }),
  trustGate: (walletAddress: string, amountRupees: number) =>
    post<TrustGateResult>("/trust-gate", {
      wallet_address: walletAddress,
      amount_rupees: amountRupees,
    }),
  purchase: (walletAddress: string, product: Product) =>
    post<PurchaseResult>("/purchase", {
      wallet_address: walletAddress,
      product,
    }),
  agentProfile: (address: string) =>
    get<AgentProfile>(`/agent-profile?address=${encodeURIComponent(address)}`),
};
