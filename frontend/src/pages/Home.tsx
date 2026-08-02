import { Link } from "react-router-dom";
import { Section, Eyebrow } from "../components/ui";
import { Pipeline } from "../components/Pipeline";
import { FAQ } from "../components/FAQ";
import {
  IconTask,
  IconSearch,
  IconWallet,
  IconGauge,
  IconGate,
  IconCard,
} from "../components/icons";

export function Home() {
  return (
    <div className="pb-24">
      <Section className="pt-16 pb-4 sm:pt-24">
        <Eyebrow>Swava — AI Procurement Manager</Eyebrow>
        <h1 className="mt-4 max-w-3xl font-display text-4xl font-semibold leading-[1.08] tracking-tight sm:text-6xl">
          Let AI agents spend company money — without letting them spend it unsupervised.
        </h1>
        <p className="mt-6 max-w-xl text-lg leading-relaxed text-ink-soft">
          Swava lets companies safely let AI agents spend company money. Every
          purchase an agent tries to make goes through a Procurement Approval
          Engine — built from the agent's own trust score — before it ever
          reaches payment.
        </p>

        <div className="mt-9 flex flex-wrap items-center gap-4">
          <Link
            to="/demo"
            className="inline-flex items-center gap-2 rounded-full bg-signal px-6 py-3 font-medium text-paper transition-opacity hover:opacity-90"
          >
            Try the live demo
            <ArrowIcon />
          </Link>
          <Link
            to="/how-it-works"
            className="inline-flex items-center gap-2 rounded-full border border-line-strong px-6 py-3 font-medium text-ink transition-colors hover:border-ink-faint"
          >
            See how it works
          </Link>
        </div>
      </Section>

      {/* The one clear visual: agent -> score -> gate, compressed */}
      <Section className="mt-16">
        <div className="rounded-3xl border border-line bg-paper-raised p-6 sm:p-10">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <Eyebrow>What this build demonstrates</Eyebrow>
            <span className="font-mono text-xs text-ink-faint">task → search → score → gate → payment</span>
          </div>
          <div className="mt-6 overflow-x-auto">
            <Pipeline
              compact
              stages={[
                { key: "task", label: "Task + budget", icon: <IconTask />, side: "neutral", state: "done" },
                { key: "search", label: "Search merchants", icon: <IconSearch />, side: "merchant", state: "done" },
                { key: "wallet", label: "Score the AI agent", icon: <IconWallet />, side: "agent", state: "done" },
                { key: "limit", label: "Apply procurement policy", icon: <IconGauge />, side: "agent", state: "done" },
                { key: "gate", label: "Approval engine decides", icon: <IconGate />, side: "agent", state: "active" },
                { key: "pay", label: "Payment attempt", icon: <IconCard />, side: "merchant", state: "idle" },
              ]}
            />
          </div>
          <p className="mt-6 max-w-2xl text-sm leading-relaxed text-ink-soft">
            Every stage above is real: a live agent searches three real Shopify
            merchants over UCP, checks a real agent trust score against ~11,000+
            indexed scores, applies a real per-agent procurement policy, and
            attempts a real payment through Prava's sandbox.
          </p>
        </div>
      </Section>

      <Section className="mt-20 grid gap-10 sm:grid-cols-4">
        <Stat n="3" label="Real merchants" detail="Searched live via UCP, not mocked." />
        <Stat n="11,000+" label="Indexed agent scores" detail="Swava's reputation index, pre-hackathon." />
        <Stat n="3" label="Approval outcomes" detail="Auto-approve, manager approval, blocked — all real, all reachable." />
        <Stat n="1" label="Procurement policy" detail="A real per-agent category cap, enforced live." />
      </Section>

      <Section className="mt-20">
        <FAQ />
      </Section>

      <Section className="mt-16">
        <div className="rounded-2xl border border-line bg-paper-raised p-6 sm:p-8">
          <Eyebrow>What's next</Eyebrow>
          <p className="mt-2 max-w-2xl text-sm leading-relaxed text-ink-soft">
            Department-level budgets — separate spend pools per team, not just per
            agent — are the natural next constraint on top of this. Not built yet;
            noted here as the direction, not a promise.
          </p>
        </div>
      </Section>

      <Section className="mt-16">
        <div className="rounded-3xl border border-line-strong bg-signal px-8 py-10 text-paper sm:px-12">
          <h2 className="max-w-lg font-display text-2xl font-semibold sm:text-3xl">
            Watch an AI agent get scored, gated, and (almost) pay for something.
          </h2>
          <Link
            to="/demo"
            className="mt-6 inline-flex items-center gap-2 rounded-full bg-paper px-6 py-3 font-medium text-signal-ink transition-opacity hover:opacity-90"
          >
            Run the live demo
            <ArrowIcon />
          </Link>
        </div>
      </Section>
    </div>
  );
}

function Stat({ n, label, detail }: { n: string; label: string; detail: string }) {
  return (
    <div>
      <div className="font-display text-4xl font-semibold text-ink">{n}</div>
      <div className="mt-1 text-sm font-medium text-ink">{label}</div>
      <div className="mt-1 text-sm text-ink-faint">{detail}</div>
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
