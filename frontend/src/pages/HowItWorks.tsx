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
        <h1 className="mt-3 max-w-2xl font-display text-4xl font-extrabold uppercase leading-[1.1] tracking-[0.02em] text-ink sm:text-[2.75rem]">
          Six steps to a procurement decision.
        </h1>
        <p className="mt-5 max-w-xl text-lg leading-relaxed text-ink">
          Every purchase goes through the same engine, in the same order.
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
        <div className="rounded-[12px] border border-line bg-teal p-8 text-center">
          <p className="text-lg text-ink">
            Every step above is live.
          </p>
          <Link
            to="/demo"
            className="mt-5 inline-flex items-center gap-2 rounded-[6px] bg-signal px-6 py-3 font-medium text-paper transition-opacity hover:opacity-90"
          >
            <ArrowIcon />
            Run it yourself
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
    body: "The only human input: a task and a spending cap.",
    icon: <IconTask />,
    side: "neutral",
  },
  {
    title: "The agent searches merchants",
    body: "Live product search over the Universal Commerce Protocol — never cached.",
    icon: <IconSearch />,
    side: "merchant",
  },
  {
    title: "The AI agent's own trust score is checked",
    body: "Not the merchant's, not the employee's — the agent's own on-chain identity.",
    icon: <IconWallet />,
    side: "agent",
  },
  {
    title: "Score and procurement policy set a limit",
    body: "70+ unlocks ₹10,000. Unrated gets a neutral ₹500. Below 30, it's ₹0. Company policy applies on top and always wins.",
    icon: <IconGauge />,
    side: "agent",
  },
  {
    title: "The Procurement Approval Engine decides",
    body: "Auto-approve under the limit. Manager approval over it. Blocked before payment if the score or policy fails.",
    icon: <IconGate />,
    side: "agent",
  },
  {
    title: "Payment is attempted",
    body: "Only past the engine — a real sandbox session, a real card, a real checkout attempt.",
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
      ? "border-signal bg-mint text-ink"
      : side === "merchant"
        ? "border-line-strong bg-paper text-ink"
        : "border-line bg-paper text-ink";

  return (
    <div className="grid grid-cols-[2.5rem_1fr] gap-5 sm:grid-cols-[3rem_420px_1fr] sm:items-center">
      <div className="font-mono text-sm text-ink sm:self-start sm:pt-4">
        {String(index).padStart(2, "0")}
      </div>
      <div className={`col-span-2 flex items-start gap-4 rounded-[12px] border p-5 sm:col-span-1 ${sideStyle}`}>
        <div className="mt-0.5 h-6 w-6 shrink-0">{icon}</div>
        <div>
          <h3 className="text-base font-bold text-ink">{title}</h3>
          <p className="mt-1.5 text-sm leading-relaxed text-ink">{body}</p>
        </div>
      </div>
      {side !== "neutral" && (
        <div className="col-span-2 hidden items-center gap-2 pl-2 sm:col-span-1 sm:flex">
          <svg width="20" height="10" viewBox="0 0 20 10" fill="none" aria-hidden="true" className="shrink-0 text-line-strong">
            <path d="M0 5h17M17 5l-4-4M17 5l-4 4" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          <span className="font-mono text-xs font-medium uppercase tracking-wide text-ink">
            {side === "agent" ? "scores the AI agent" : "concerns the merchant, not scoring"}
          </span>
        </div>
      )}
    </div>
  );
}

function WhoGetsScored() {
  return (
    <div className="rounded-[12px] border border-line bg-paper-raised p-8 sm:p-10">
      <Eyebrow>The one thing to get right</Eyebrow>
      <h2 className="mt-2 font-display text-2xl font-extrabold text-ink">
        It's the buying AI agent that gets scored — never the merchant.
      </h2>
      <div className="mt-8 grid gap-6 sm:grid-cols-[1fr_auto_1fr]">
        <ActorCard
          icon={<IconAgent />}
          label="AI agent"
          sub="has an identity"
          detail="Its own history is the trust signal — score, policy, decision."
          highlighted
        />
        <div className="flex items-center justify-center py-2 sm:py-0">
          <div className="flex flex-col items-center gap-1 text-ink">
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
          detail="Sells the product. Never scored — see below."
        />
      </div>

      <div className="mt-6 rounded-[12px] border border-line-strong bg-whisper px-6 py-5">
        <div className="flex items-center gap-2 font-mono text-xs uppercase tracking-wide text-ink">
          Merchant trust
          <span className="rounded-full bg-highlight px-2 py-0.5 text-[10px] normal-case tracking-normal text-ink">
            not yet available
          </span>
        </div>
        <p className="mt-2 max-w-2xl text-sm leading-relaxed text-ink">
          No merchant score exists yet, real or placeholder. See{" "}
          <Link to="/findings" className="font-medium text-ink underline decoration-signal underline-offset-4">
            Findings
          </Link>
          .
        </p>
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
      className={`rounded-[12px] border p-6 ${
        highlighted
          ? "border-signal bg-signal text-paper"
          : "border-line-strong bg-paper text-ink"
      }`}
    >
      <div className={`h-8 w-8 ${highlighted ? "text-paper" : "text-ink"}`}>{icon}</div>
      <div className="mt-4 font-display text-xl font-extrabold">{label}</div>
      <div
        className={`font-mono text-xs uppercase tracking-wide ${
          highlighted ? "text-paper/70" : "text-ink"
        }`}
      >
        {sub}
      </div>
      <p className={`mt-3 text-sm leading-relaxed ${highlighted ? "text-paper/90" : "text-ink"}`}>
        {detail}
      </p>
      {highlighted && (
        <div className="mt-4 inline-flex items-center gap-1.5 rounded-full bg-highlight px-3 py-1 text-xs font-medium text-ink">
          scored here
        </div>
      )}
    </div>
  );
}
