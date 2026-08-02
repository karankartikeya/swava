import { Eyebrow } from "./ui";

/** Shared across Home, How It Works, and Live Demo — anticipates the
 * sharpest objection: if a company already sets a spending mandate for its
 * agent, why does it also need a trust score. */
export function FAQ() {
  return (
    <div className="rounded-[12px] border border-line bg-paper-raised p-8 sm:p-10">
      <Eyebrow>Before you ask</Eyebrow>
      <h2 className="mt-2 max-w-xl font-display text-2xl font-extrabold leading-snug text-ink">
        Why would I want to limit my own agent?
      </h2>

      <div className="mt-6 flex flex-col gap-4 text-[15px] leading-relaxed text-ink">
        <p>
          Two questions get conflated. <strong className="text-ink">Can the agent spend</strong> —
          that's the procurement policy. <strong className="text-ink">Should it spend right now,
          unchecked</strong> — that's the trust score. If a manager reviews every purchase, the
          score is redundant. It matters once the agent acts alone.
        </p>

        <p>
          For a single founder on their own funds, watching closely — this mostly doesn't
          apply. Fair objection.
        </p>

        <p>
          It matters at real headcount: a finance team that didn't build the agent, several
          agents needing one shared trust signal, an auditor asking why a purchase was
          allowed. Same reason employee spending controls exist — nobody audits their own
          paycheck. Scores exist for whoever answers for someone else's spending.
        </p>

        <p>
          Even for one well-behaved agent: a dropping score warns of compromised or
          changed behavior that a budget cap alone can't catch — a policy checks the
          ceiling, never what changed underneath it.
        </p>
      </div>
    </div>
  );
}
