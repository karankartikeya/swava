import { useEffect, useState } from "react";
import { api } from "../lib/api";
import type { DecideResult, DemoWallet, PurchaseResult } from "../lib/api";
import { Section, Eyebrow, DecisionBadge, decisionMetaFor } from "../components/ui";
import { IconSearch, IconCheck, IconCard } from "../components/icons";

type Phase = "idle" | "deciding" | "decided" | "purchasing" | "purchased" | "error";

// The Go backend writes plain-ASCII "Rs." in free-text reasoning strings
// (keeps CLI/log output ASCII-safe) — swap in the ₹ glyph used everywhere
// else in this UI purely for display.
function rupeeify(text: string): string {
  return text.replace(/Rs\.(\d)/g, "₹$1");
}

export function Demo() {
  const [wallets, setWallets] = useState<DemoWallet[]>([]);
  const [task, setTask] = useState("buy a cheap razor");
  const [budget, setBudget] = useState(500);
  const [walletAddress, setWalletAddress] = useState("");

  const [phase, setPhase] = useState<Phase>("idle");
  const [decideResult, setDecideResult] = useState<DecideResult | null>(null);
  const [purchaseResult, setPurchaseResult] = useState<PurchaseResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .wallets()
      .then((w) => {
        setWallets(w);
        if (w.length > 0) setWalletAddress(w[0].address);
      })
      .catch(() => setError("Could not reach the backend at /api. Is cmd/server running?"));
  }, []);

  async function runDemo() {
    setError(null);
    setDecideResult(null);
    setPurchaseResult(null);
    setPhase("deciding");
    try {
      const decide = await api.decide(task, budget);
      setDecideResult(decide);
      setPhase("decided");

      if (!decide.chosen_product) return;

      setPhase("purchasing");
      const purchase = await api.purchase(walletAddress, decide.chosen_product);
      setPurchaseResult(purchase);
      setPhase("purchased");
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
      setPhase("error");
    }
  }

  const running = phase === "deciding" || phase === "purchasing";

  return (
    <div className="pb-24">
      <Section className="pt-16 pb-10">
        <Eyebrow>Live demo</Eyebrow>
        <h1 className="mt-3 max-w-2xl font-display text-4xl font-semibold leading-[1.1] tracking-tight sm:text-[2.75rem]">
          Give the agent a task and a budget.
        </h1>
        <p className="mt-5 max-w-xl text-lg leading-relaxed text-ink-soft">
          Every call below hits the real backend — real merchant search, a real Swava
          score, a real Prava sandbox session.
        </p>
      </Section>

      <Section>
        <DemoForm
          task={task}
          setTask={setTask}
          budget={budget}
          setBudget={setBudget}
          wallets={wallets}
          walletAddress={walletAddress}
          setWalletAddress={setWalletAddress}
          onRun={runDemo}
          running={running}
        />
      </Section>

      {error && (
        <Section className="mt-8">
          <div className="rounded-2xl border border-block bg-block-soft p-5 text-sm text-block-ink">
            {error}
          </div>
        </Section>
      )}

      {(decideResult || phase === "deciding") && (
        <Section className="mt-12">
          <AgentTrail result={decideResult} loading={phase === "deciding"} />
        </Section>
      )}

      {decideResult?.chosen_product && (
        <Section className="mt-12">
          <TrustGateCenterpiece
            walletAddress={walletAddress}
            wallets={wallets}
            product={decideResult.chosen_product}
            purchase={purchaseResult}
            loading={phase === "purchasing"}
            walletRole={decideResult.wallet_role}
          />
        </Section>
      )}

      {purchaseResult && (
        <Section className="mt-12">
          <PaymentOutcome result={purchaseResult} />
        </Section>
      )}
    </div>
  );
}

function DemoForm({
  task,
  setTask,
  budget,
  setBudget,
  wallets,
  walletAddress,
  setWalletAddress,
  onRun,
  running,
}: {
  task: string;
  setTask: (v: string) => void;
  budget: number;
  setBudget: (v: number) => void;
  wallets: DemoWallet[];
  walletAddress: string;
  setWalletAddress: (v: string) => void;
  onRun: () => void;
  running: boolean;
}) {
  return (
    <div className="rounded-3xl border border-line bg-paper-raised p-6 sm:p-8">
      <div className="grid gap-6 sm:grid-cols-2">
        <label className="flex flex-col gap-2">
          <span className="text-sm font-medium text-ink">Task</span>
          <input
            className="rounded-xl border border-line-strong bg-paper px-4 py-3 text-sm text-ink outline-none focus:border-signal"
            value={task}
            onChange={(e) => setTask(e.target.value)}
            placeholder="buy a cheap razor"
          />
        </label>
        <label className="flex flex-col gap-2">
          <span className="text-sm font-medium text-ink">Budget (₹)</span>
          <input
            type="number"
            min={1}
            className="tabular rounded-xl border border-line-strong bg-paper px-4 py-3 text-sm text-ink outline-none focus:border-signal"
            value={budget}
            onChange={(e) => setBudget(Number(e.target.value))}
          />
        </label>
      </div>

      <div className="mt-6">
        <span className="text-sm font-medium text-ink">Wallet</span>
        <div className="mt-2 grid gap-3 sm:grid-cols-3">
          {wallets.map((w) => (
            <button
              key={w.address}
              type="button"
              onClick={() => setWalletAddress(w.address)}
              className={`rounded-xl border p-4 text-left transition-colors ${
                walletAddress === w.address
                  ? "border-signal bg-signal-soft"
                  : "border-line-strong bg-paper hover:border-ink-faint"
              }`}
            >
              <div className="text-sm font-semibold text-ink">{w.label}</div>
              <div className="mt-1 font-mono text-[11px] text-ink-faint">
                {w.address.slice(0, 8)}…{w.address.slice(-4)}
              </div>
              <div className="mt-1.5 text-xs leading-relaxed text-ink-soft">{w.description}</div>
            </button>
          ))}
        </div>
      </div>

      <button
        type="button"
        onClick={onRun}
        disabled={running || !walletAddress}
        className="mt-7 inline-flex items-center gap-2 rounded-full bg-signal px-6 py-3 font-medium text-paper transition-opacity hover:opacity-90 disabled:opacity-50"
      >
        {running ? "Running…" : "Run the agent"}
      </button>
    </div>
  );
}

function AgentTrail({ result, loading }: { result: DecideResult | null; loading: boolean }) {
  return (
    <div>
      <Eyebrow>Agent trail</Eyebrow>
      <h2 className="mt-2 font-display text-2xl font-semibold text-ink">
        What it searched, and why it picked what it picked
      </h2>

      <div className="mt-6 flex flex-col gap-3">
        {loading && !result && (
          <div className="flex items-center gap-3 rounded-xl border border-line bg-paper-raised px-5 py-4 text-sm text-ink-soft">
            <Spinner /> Searching merchants live…
          </div>
        )}
        {result?.steps.map((step) => (
          <div
            key={step.merchant_name}
            className={`rounded-xl border px-5 py-4 ${
              step.rejected
                ? "border-line bg-paper opacity-70"
                : step.configured && step.reachable && !step.rejected && result.chosen_product?.merchant_name === step.merchant_name
                  ? "border-signal bg-signal-soft"
                  : "border-line bg-paper-raised"
            }`}
          >
            <div className="flex flex-wrap items-center gap-2">
              <span className="h-4 w-4 text-ink-faint">
                <IconSearch />
              </span>
              <span className="font-medium text-ink">{step.merchant_name}</span>
              <span className="font-mono text-xs text-ink-faint">
                search_catalog("{step.search_term}")
              </span>
              {step.task_match && (
                <span className="rounded-full bg-signal-soft px-2 py-0.5 text-[11px] font-medium text-signal-ink">
                  task match
                </span>
              )}
            </div>

            <div className="mt-2 text-sm text-ink-soft">
              {!step.configured && "Merchant not configured — skipped."}
              {step.configured && !step.reachable && `Unreachable: ${step.error}`}
              {step.configured && step.reachable && step.found_title && (
                <>
                  Found <span className="font-medium text-ink">{step.found_title}</span> —{" "}
                  <span className="tabular font-medium text-ink">
                    ₹{step.found_price_rupees?.toLocaleString("en-IN")}
                  </span>
                  {step.confirmed && " (confirmed live match)"}
                  {step.rejected && (
                    <span className="text-block-ink"> — {rupeeify(step.reject_reason ?? "")}</span>
                  )}
                </>
              )}
              {step.configured && step.reachable && !step.found_title && "No results."}
            </div>
          </div>
        ))}

        {result && (
          <div className="mt-2 rounded-xl border border-signal bg-signal px-5 py-4 text-sm text-paper">
            <span className="h-4 w-4 inline-block align-text-bottom mr-2">
              <IconCheck />
            </span>
            {rupeeify(result.reasoning)}
          </div>
        )}
      </div>
    </div>
  );
}

function TrustGateCenterpiece({
  walletAddress,
  wallets,
  product,
  purchase,
  loading,
  walletRole,
}: {
  walletAddress: string;
  wallets: DemoWallet[];
  product: DecideResult["chosen_product"];
  purchase: PurchaseResult | null;
  loading: boolean;
  walletRole: string;
}) {
  const wallet = wallets.find((w) => w.address === walletAddress);
  const gate = purchase?.trust_gate;

  return (
    <div>
      <Eyebrow>Trust gate</Eyebrow>
      <h2 className="mt-2 font-display text-2xl font-semibold text-ink">
        The decision that gates the payment
      </h2>
      <p className="mt-2 max-w-2xl text-sm leading-relaxed text-ink-soft">{walletRole}</p>

      <div className="mt-6 overflow-hidden rounded-3xl border border-line-strong bg-paper-raised">
        {loading && !gate ? (
          <div className="flex items-center gap-3 px-8 py-10 text-sm text-ink-soft">
            <Spinner /> Checking wallet reputation…
          </div>
        ) : gate ? (
          <GateReadout gate={gate} wallet={wallet} product={product} />
        ) : (
          <div className="px-8 py-10 text-sm text-ink-faint">Waiting on the decision…</div>
        )}
      </div>
    </div>
  );
}

function GateReadout({
  gate,
  wallet,
  product,
}: {
  gate: NonNullable<PurchaseResult["trust_gate"]>;
  wallet?: DemoWallet;
  product: DecideResult["chosen_product"];
}) {
  const meta = decisionMetaFor(gate.decision);
  return (
    <div>
      <div className={`px-8 py-6 ${meta.bg} border-b ${meta.border}`}>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <DecisionBadge decision={gate.decision} />
          <span className={`text-sm ${meta.text}`}>{gate.reason}</span>
        </div>
      </div>
      <div className="grid gap-px bg-line sm:grid-cols-3">
        <Field label="Wallet" value={wallet?.label ?? "—"} sub={`${gate.wallet_address.slice(0, 10)}…`} />
        <Field
          label="Score"
          value={`${gate.normalized_score}/100`}
          sub={gate.known ? `raw ${gate.raw_score}/1000, tier ${gate.tier}` : "unknown wallet — neutral default"}
        />
        <Field label="Spend limit" value={`₹${gate.spend_limit_rupees.toLocaleString("en-IN")}`} />
        <Field
          label="Purchase amount"
          value={`₹${product?.unit_price_rupees.toLocaleString("en-IN") ?? "—"}`}
          sub={product?.description}
        />
        <Field label="Decision" value={gate.decision.replace("_", " ")} />
        <Field label="Reason" value={gate.reason} />
      </div>
    </div>
  );
}

function Field({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div className="bg-paper-raised px-6 py-5">
      <div className="font-mono text-[11px] uppercase tracking-wide text-ink-faint">{label}</div>
      <div className="tabular mt-1 text-base font-medium text-ink">{value}</div>
      {sub && <div className="mt-0.5 text-xs text-ink-faint">{sub}</div>}
    </div>
  );
}

function PaymentOutcome({ result }: { result: PurchaseResult }) {
  const blocked = result.trust_gate.decision === "blocked";

  return (
    <div>
      <Eyebrow>Payment outcome</Eyebrow>
      <h2 className="mt-2 font-display text-2xl font-semibold text-ink">
        {blocked ? "Prava was never called" : "A real sandbox session was created"}
      </h2>

      {blocked ? (
        <div className="mt-6 rounded-2xl border border-block bg-block-soft p-6">
          <p className="text-sm leading-relaxed text-block-ink">
            The wallet was blocked at the trust gate, before any payment provider was
            contacted. No session was created, no card was issued — the flow stops
            entirely on our side. This is the point of the gate: a bad wallet never
            reaches money.
          </p>
        </div>
      ) : (
        <div className="mt-6 flex flex-col gap-4">
          <div className="rounded-2xl border border-line-strong bg-paper-raised p-6">
            <div className="flex items-center gap-2 text-sm font-medium text-ink">
              <span className="h-4 w-4 text-signal">
                <IconCard />
              </span>
              Real Prava sandbox session
            </div>
            <div className="mt-4 grid gap-4 sm:grid-cols-2">
              <Field label="Session ID" value={result.prava_session_id ?? "—"} />
              <Field label="Order ID" value={result.prava_order_id ?? "—"} />
            </div>
            {result.iframe_url && (
              <a
                href={result.iframe_url}
                target="_blank"
                rel="noreferrer"
                className="mt-4 inline-block break-all font-mono text-xs text-signal underline decoration-signal/40 underline-offset-4"
              >
                {result.iframe_url}
              </a>
            )}
          </div>

          {result.sandbox_note && (
            <div className="rounded-2xl border border-review bg-review-soft p-6">
              <div className="text-sm font-semibold text-review-ink">
                Sandbox stops here — this is a finding, not a failure
              </div>
              <p className="mt-2 text-sm leading-relaxed text-review-ink">
                {result.sandbox_note}
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function Spinner() {
  return (
    <span
      className="h-4 w-4 shrink-0 animate-spin rounded-full border-2 border-line-strong border-t-signal"
      aria-hidden="true"
    />
  );
}
