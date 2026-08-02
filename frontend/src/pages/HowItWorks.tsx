import { Link } from "react-router-dom";
import { Section, Eyebrow } from "../components/ui";
import { FAQ } from "../components/FAQ";
import {
  IconTask,
  IconSearch,
  IconWallet,
  IconGauge,
  IconGate,
  IconCard,
  IconAgent,
  IconStore,
} from "../components/icons";

export function HowItWorks() {
  return (
    <div className="pb-24">
      <Section className="pt-16 pb-10">
        <Eyebrow>How it works</Eyebrow>
        <h1 className="mt-3 max-w-2xl font-display text-4xl font-semibold leading-[1.1] tracking-tight sm:text-[2.75rem]">
          Six steps from a task to a payment decision.
        </h1>
        <p className="mt-5 max-w-xl text-lg leading-relaxed text-ink-soft">
          Every purchase this system makes goes through the same sequence, in the same
          order, with no step skipped.
        </p>
      </Section>

      {/* The confusion-resolving diagram: who gets scored */}
      <Section className="mt-6">
        <WhoGetsScored />
      </Section>

      <Section className="mt-16">
        <div className="flex flex-col gap-3">
          {steps.map((step, i) => (
            <StepRow key={step.title} index={i + 1} {...step} />
          ))}
        </div>
      </Section>

      <Section className="mt-16">
        <FAQ />
      </Section>

      <Section className="mt-16">
        <div className="rounded-2xl border border-line bg-paper-raised p-8 text-center">
          <p className="text-lg text-ink-soft">
            Every one of these steps is live — real merchant search, a real reputation
            lookup, a real sandbox payment session.
          </p>
          <Link
            to="/demo"
            className="mt-5 inline-flex items-center gap-2 rounded-full bg-signal px-6 py-3 font-medium text-paper transition-opacity hover:opacity-90"
          >
            Run it yourself
            <ArrowIcon />
          </Link>
        </div>
      </Section>
    </div>
  );
}

function ArrowIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
      <path
        d="M3 8h9M8 3l5 5-5 5"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

const steps: { title: string; body: string; icon: React.ReactNode; side: "agent" | "merchant" | "neutral" }[] = [
  {
    title: "Task and budget come in",
    body: "A task description (\"buy a cheap razor\") and a spending cap in rupees. This is the only human input to the whole flow.",
    icon: <IconTask />,
    side: "neutral",
  },
  {
    title: "The agent searches merchants",
    body: "It queries every configured merchant's live product catalog over the Universal Commerce Protocol (UCP), in real time — not a cached or mocked list.",
    icon: <IconSearch />,
    side: "merchant",
  },
  {
    title: "The agent's own wallet reputation is checked",
    body: "Not the merchant's. Not the human's. The wallet belongs to the agent itself — its own persistent, on-chain economic identity, queried against Swava's reputation index.",
    icon: <IconWallet />,
    side: "agent",
  },
  {
    title: "Score converts to a spend limit",
    body: "A score of 70+ unlocks a ₹10,000 limit. A new or unknown wallet gets a neutral ₹500 limit — never zero, never an error. Below 30, the limit is ₹0.",
    icon: <IconGauge />,
    side: "agent",
  },
  {
    title: "The gate decides",
    body: "Auto-approve if the purchase is under the limit. Human review if it's over. Blocked outright if the score itself is too low — before any payment system is ever contacted.",
    icon: <IconGate />,
    side: "agent",
  },
  {
    title: "Payment is attempted",
    body: "Only past the gate. A real sandbox session is created, a card is issued, and checkout is attempted against the merchant — the last mile with real, documented findings of its own.",
    icon: <IconCard />,
    side: "merchant",
  },
];

function StepRow({
  index,
  title,
  body,
  icon,
  side,
}: {
  index: number;
  title: string;
  body: string;
  icon: React.ReactNode;
  side: "agent" | "merchant" | "neutral";
}) {
  const sideStyle =
    side === "agent"
      ? "border-signal/40 bg-signal-soft text-signal-ink"
      : side === "merchant"
        ? "border-line-strong bg-paper-raised text-ink-soft"
        : "border-line bg-paper text-ink-soft";

  return (
    <div className="grid grid-cols-[2.5rem_1fr] gap-5 sm:grid-cols-[3rem_420px_1fr] sm:items-center">
      <div className="font-mono text-sm text-ink-faint sm:self-start sm:pt-4">
        {String(index).padStart(2, "0")}
      </div>
      <div className={`col-span-2 flex items-start gap-4 rounded-2xl border p-5 sm:col-span-1 ${sideStyle}`}>
        <div className="mt-0.5 h-6 w-6 shrink-0">{icon}</div>
        <div>
          <h3 className="text-base font-semibold text-ink">{title}</h3>
          <p className="mt-1.5 text-sm leading-relaxed text-ink-soft">{body}</p>
        </div>
      </div>
      {side !== "neutral" && (
        <div className="col-span-2 hidden items-center gap-2 pl-2 sm:col-span-1 sm:flex">
          <svg width="20" height="10" viewBox="0 0 20 10" fill="none" aria-hidden="true" className="shrink-0 text-line-strong">
            <path d="M0 5h17M17 5l-4-4M17 5l-4 4" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          <span className="text-xs font-medium uppercase tracking-wide text-ink-faint">
            {side === "agent" ? "scores the agent's wallet" : "concerns the merchant, not scoring"}
          </span>
        </div>
      )}
    </div>
  );
}

function WhoGetsScored() {
  return (
    <div className="rounded-3xl border border-line bg-paper-raised p-8 sm:p-10">
      <Eyebrow>The one thing to get right</Eyebrow>
      <h2 className="mt-2 font-display text-2xl font-semibold text-ink">
        It's the buying agent that gets scored — never the merchant.
      </h2>
      <div className="mt-8 grid gap-6 sm:grid-cols-[1fr_auto_1fr]">
        <ActorCard
          icon={<IconAgent />}
          label="Buying agent"
          sub="has a wallet"
          detail="Its wallet's transaction history is the reputation signal. This is what gets a score, a spend limit, and a gate decision."
          highlighted
        />
        <div className="flex items-center justify-center py-2 sm:py-0">
          <div className="flex flex-col items-center gap-1 text-ink-faint">
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none" aria-hidden="true" className="rotate-90 sm:rotate-0">
              <path d="M3 12h16M13 6l6 6-6 6" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
            <span className="font-mono text-[10px] uppercase tracking-wide">pays</span>
          </div>
        </div>
        <ActorCard
          icon={<IconStore />}
          label="Merchant"
          sub="has a store"
          detail="Sells the product. Its catalog is searched live. Its reputation is not evaluated by this system at all."
        />
      </div>
    </div>
  );
}

function ActorCard({
  icon,
  label,
  sub,
  detail,
  highlighted = false,
}: {
  icon: React.ReactNode;
  label: string;
  sub: string;
  detail: string;
  highlighted?: boolean;
}) {
  return (
    <div
      className={`rounded-2xl border p-6 ${
        highlighted
          ? "border-signal bg-signal text-paper"
          : "border-line-strong bg-paper text-ink"
      }`}
    >
      <div className={`h-8 w-8 ${highlighted ? "text-paper" : "text-ink-faint"}`}>{icon}</div>
      <div className="mt-4 font-display text-xl font-semibold">{label}</div>
      <div
        className={`font-mono text-xs uppercase tracking-wide ${
          highlighted ? "text-paper/70" : "text-ink-faint"
        }`}
      >
        {sub}
      </div>
      <p className={`mt-3 text-sm leading-relaxed ${highlighted ? "text-paper/90" : "text-ink-soft"}`}>
        {detail}
      </p>
      {highlighted && (
        <div className="mt-4 inline-flex items-center gap-1.5 rounded-full bg-paper/15 px-3 py-1 text-xs font-medium">
          scored here
        </div>
      )}
    </div>
  );
}
