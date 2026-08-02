import { Section, Eyebrow } from "../components/ui";

interface Finding {
  title: string;
  body: string;
}

const findings: Finding[] = [
  {
    title: "Raw UCP checkout completion isn't self-servable for third-party agents",
    body: "search_catalog, get_product, and create_checkout only require a UCP agent-profile handshake — any agent can call them against any merchant. complete_checkout is different: it requires a merchant-side JWT that has to be issued through Shopify's own developer console, per merchant. There's no self-serve path to obtain one as an independent agent operator. This is a real, structural gap between \"the protocol supports agent checkout\" and \"an agent can actually complete a checkout it didn't build the storefront for.\"",
  },
  {
    title: "Manually completing checkout fails 3D Secure at the payment gateway",
    body: "Working around the JWT gap by submitting the issued card token directly to a merchant's regional payment gateway (Razorpay, in this build's case) gets further — but fails 3D Secure verification. The token Prava issues is built for programmatic handoff between systems that trust each other's context; it isn't shaped for a human-facing 3DS challenge, which expects a live cardholder session. Fighting this by hand confirmed the token format itself is the blocker, not a configuration issue.",
  },
  {
    title: "The documented supported path is production-only",
    body: "Prava's Browser Harness is the actual supported route for an agent to complete a checkout end to end — but it runs against production, uses real cards, and has no direct API. It's accessed through a CLI or a hosted-agent integration, not a request you can make from a backend. There is currently no sandbox equivalent, which means this specific gap can't be closed inside a hackathon's sandbox constraints — only worked around or documented, which is what this build does.",
  },
  {
    title: "Prava's payment identity and Swava's wallet identity aren't bound to each other",
    body: "A Prava account (a real card plus a passkey) and a Swava wallet (an on-chain address with reputation history) are two separate identity systems. Nothing today cryptographically proves that a given wallet and a given Prava account belong to the same agent operator. A real deployment needs an onboarding step — some signed attestation or registration flow — that binds the two before the trust gate's score can be trusted to mean anything about who's actually paying. This demo assumes that binding for illustration. Enforcing it is future work, not a solved problem.",
  },
];

export function Findings() {
  return (
    <div className="pb-24">
      <Section className="pt-16 pb-10">
        <Eyebrow>Findings</Eyebrow>
        <h1 className="mt-3 max-w-2xl font-display text-4xl font-semibold leading-[1.1] tracking-tight sm:text-[2.75rem]">
          What this build actually found.
        </h1>
        <p className="mt-5 max-w-xl text-lg leading-relaxed text-ink-soft">
          Four real, technical discoveries — not framed as failures. Each one is a
          concrete gap between what the protocols document and what an independent
          agent operator can actually do today.
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
    <div className="rounded-2xl border border-line bg-paper-raised p-6 sm:p-8">
      <div className="flex items-start gap-4">
        <span className="mt-1 font-mono text-sm text-ink-faint">{String(index).padStart(2, "0")}</span>
        <div>
          <h2 className="font-display text-xl font-semibold text-ink">{finding.title}</h2>
          <p className="mt-3 max-w-3xl text-[15px] leading-relaxed text-ink-soft">{finding.body}</p>
        </div>
      </div>
    </div>
  );
}

function Disclosure() {
  return (
    <div className="rounded-2xl border border-line-strong bg-paper px-7 py-8">
      <div className="font-mono text-xs uppercase tracking-[0.14em] text-ink-faint">Disclosure</div>
      <p className="mt-3 max-w-3xl text-[15px] leading-relaxed text-ink-soft">
        Swava's reputation API, the on-chain indexer, and ~11,000+ already-indexed
        wallet scores existed before this hackathon. Everything else — the agent
        decision loop, the risk policy that converts a score into a spend limit, the
        Prava sandbox integration, the merchant checkout flow over UCP, the HTTP layer
        connecting all of it, and this site — was built during it.
      </p>
    </div>
  );
}
