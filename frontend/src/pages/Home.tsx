import { Link } from "react-router-dom";
import { Section, Eyebrow } from "../components/ui";
import { Pipeline } from "../components/Pipeline";
import { FAQ } from "../components/FAQ";
import { Reveal } from "../components/Reveal";
import heroSketch from "../assets/hero-sketch.png";
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
      <Section className="relative overflow-hidden pt-16 pb-4 sm:pt-24">
        <img
          src={heroSketch}
          alt=""
          aria-hidden="true"
          className="float-slow pointer-events-none absolute -right-10 top-0 hidden w-[420px] opacity-40 sm:block lg:-right-4 lg:w-[520px]"
        />
        <div className="relative max-w-2xl">
          <div className="rise-in">
            <Eyebrow>Swava — AI Procurement Manager</Eyebrow>
          </div>
          <h1
            className="rise-in mt-4 font-display text-4xl font-extrabold uppercase leading-[1.05] tracking-[0.02em] text-ink sm:text-6xl"
            style={{ animationDelay: "80ms" }}
          >
            Let agents spend company money —{" "}
            <span className="mark">not unsupervised</span>.
          </h1>
          <p
            className="rise-in mt-6 max-w-md text-lg leading-relaxed text-ink"
            style={{ animationDelay: "160ms" }}
          >
            Every purchase an AI agent makes passes through a Procurement
            Approval Engine, built from the agent's own trust score.
          </p>

          <div
            className="rise-in mt-9 flex flex-wrap items-center gap-4"
            style={{ animationDelay: "240ms" }}
          >
            <Link
              to="/demo"
              className="inline-flex items-center gap-2 rounded-[6px] bg-signal px-6 py-3 font-medium text-paper shadow-[0_1px_2px_rgba(0,0,0,0.05)] transition-transform hover:-translate-y-0.5 hover:opacity-90"
            >
              <ArrowIcon />
              Try the live demo
            </Link>
            <Link
              to="/how-it-works"
              className="inline-flex items-center gap-2 rounded-[6px] border border-line-strong px-6 py-3 font-medium text-ink transition-colors hover:bg-whisper"
            >
              See how it works
            </Link>
          </div>
        </div>
      </Section>

      {/* The one clear visual: agent -> score -> gate, compressed */}
      <Reveal>
        <Section className="mt-16">
          <div className="rounded-[12px] border border-line bg-paper-raised p-6 sm:p-10">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <Eyebrow>What this build demonstrates</Eyebrow>
              <span className="font-mono text-xs text-ink">task → search → score → gate → payment</span>
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
            <p className="mt-6 max-w-2xl text-sm leading-relaxed text-ink">
              Every stage is real — live merchant search, a live trust score, a live
              sandbox payment.
            </p>
          </div>
        </Section>
      </Reveal>

      <Reveal>
        <Section className="mt-20 grid gap-10 sm:grid-cols-4">
          <Stat n="3" label="Real merchants" detail="Searched live via UCP." />
          <Stat n="11,000+" label="Indexed agent scores" detail="Pre-hackathon reputation index." />
          <Stat n="3" label="Approval outcomes" detail="All real, all reachable." />
          <Stat n="1" label="Procurement policy" detail="Enforced live." />
        </Section>
      </Reveal>

      <Reveal>
        <Section className="mt-20">
          <FAQ />
        </Section>
      </Reveal>

      <Reveal>
        <Section className="mt-16">
          <div className="rounded-[12px] border border-line bg-mint p-6 sm:p-8">
            <Eyebrow>What's next</Eyebrow>
            <p className="mt-2 max-w-2xl text-sm leading-relaxed text-ink">
              Department-level budgets — separate spend pools per team — are the
              natural next constraint. Not built yet.
            </p>
          </div>
        </Section>
      </Reveal>

      <Reveal>
        <Section className="mt-16">
          <div className="rounded-[12px] border border-line-strong bg-signal px-8 py-10 text-paper sm:px-12">
            <h2 className="max-w-lg font-display text-2xl font-extrabold uppercase tracking-[0.02em] text-paper sm:text-3xl">
              Watch an agent get scored, gated, and{" "}
              <span className="mark text-ink">(almost)</span> pay.
            </h2>
            <Link
              to="/demo"
              className="mt-6 inline-flex items-center gap-2 rounded-[6px] bg-paper px-6 py-3 font-medium text-ink shadow-[0_1px_2px_rgba(0,0,0,0.05)] transition-transform hover:-translate-y-0.5 hover:opacity-90"
            >
              <ArrowIcon />
              Run the live demo
            </Link>
          </div>
        </Section>
      </Reveal>
    </div>
  );
}

function Stat({ n, label, detail }: { n: string; label: string; detail: string }) {
  return (
    <div>
      <div className="font-display text-4xl font-extrabold text-ink">{n}</div>
      <div className="mt-1 text-sm font-medium text-ink">{label}</div>
      <div className="mt-1 text-sm text-ink">{detail}</div>
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
