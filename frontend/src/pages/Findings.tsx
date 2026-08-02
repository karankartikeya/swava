import { Section, Eyebrow } from "../components/ui";

interface Finding {
  title: string;
  body: string;
}

const findings: Finding[] = [
  {
    title: "Raw UCP checkout completion isn't self-servable for third-party agents",
    body: "search_catalog, get_product, and create_checkout only need a UCP agent-profile handshake. complete_checkout needs a merchant-side JWT issued through Shopify's developer console, per merchant — no self-serve path exists. A real gap between \"the protocol supports agent checkout\" and \"an agent can complete one it didn't build the storefront for.\"",
  },
  {
    title: "Manually completing checkout fails 3D Secure at the payment gateway",
    body: "Submitting Prava's issued token directly to a merchant's payment gateway (Razorpay, here) gets further — but fails 3D Secure, which expects a live cardholder session. The token is built for system-to-system handoff, not a human-facing challenge. The token format is the blocker, not a config issue.",
  },
  {
    title: "The documented supported path is production-only",
    body: "Prava's Browser Harness is the real supported route for an agent to complete checkout — but it's production-only, real cards, no direct API, CLI or hosted-agent only. No sandbox equivalent exists, so this gap can't be closed inside a hackathon sandbox — only documented.",
  },
  {
    title: "Prava's payment identity and Swava's wallet identity aren't bound to each other",
    body: "A Prava account (card + passkey) and a Swava wallet (on-chain reputation) are separate systems. Nothing cryptographically proves they belong to the same operator. A real deployment needs a signed binding step before the score means anything about who's actually paying. This demo assumes it. Enforcing it is future work.",
  },
];

export function Findings() {
  return (
    <div className="pb-24">
      <Section className="pt-16 pb-10">
        <Eyebrow>Findings</Eyebrow>
        <h1 className="mt-3 max-w-2xl font-display text-4xl font-extrabold uppercase leading-[1.1] tracking-[0.02em] text-ink sm:text-[2.75rem]">
          What this build actually found.
        </h1>
        <p className="mt-5 max-w-xl text-lg leading-relaxed text-ink">
          Four real gaps between what the protocols document and what an
          independent agent can actually do.
        </p>
      </Section>

      <Section>
        <div className="flex flex-col gap-4">
          {findings.map((f, i) => (
            <FindingCard key={f.title} index={i + 1} finding={f} />
          ))}
        </div>
      </Section>

      <Section className="mt-16">
        <Disclosure />
      </Section>
    </div>
  );
}

function FindingCard({ index, finding }: { index: number; finding: Finding }) {
  return (
    <div className="rounded-[12px] border border-line bg-paper-raised p-6 sm:p-8">
      <div className="flex items-start gap-4">
        <span className="mt-1 font-mono text-sm text-ink">{String(index).padStart(2, "0")}</span>
        <div>
          <h2 className="font-display text-xl font-extrabold text-ink">{finding.title}</h2>
          <p className="mt-3 max-w-3xl text-[15px] leading-relaxed text-ink">{finding.body}</p>
        </div>
      </div>
    </div>
  );
}

function Disclosure() {
  return (
    <div className="rounded-[12px] border border-line-strong bg-whisper px-7 py-8">
      <div className="font-mono text-xs uppercase tracking-[0.14em] text-ink">Disclosure</div>
      <p className="mt-3 max-w-3xl text-[15px] leading-relaxed text-ink">
        The reputation API, indexer, and ~11,000+ indexed scores pre-date this
        hackathon. Everything else — decision loop, risk policy, Prava integration,
        UCP checkout flow, HTTP layer, this site — was built during it.
      </p>
    </div>
  );
}
