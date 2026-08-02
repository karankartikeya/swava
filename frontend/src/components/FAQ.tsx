import { Eyebrow } from "./ui";

/** Shared across Home, How It Works, and Live Demo — anticipates the
 * sharpest objection: if a company already sets a spending mandate for its
 * agent, why does it also need a trust score. */
export function FAQ() {
  return (
    <div className="rounded-3xl border border-line bg-paper-raised p-8 sm:p-10">
      <Eyebrow>Before you ask</Eyebrow>
      <h2 className="mt-2 max-w-xl font-display text-2xl font-semibold leading-snug text-ink">
        Why would I want to limit my own agent?
      </h2>

      <div className="mt-6 flex flex-col gap-4 text-[15px] leading-relaxed text-ink-soft">
        <p>
          Two different questions get conflated here. <strong className="text-ink">Does the
          agent have a budget to spend</strong> — that's the procurement policy, set once by
          whoever owns the agent. <strong className="text-ink">Should it be trusted to spend
          right now, without a manager personally checking</strong> — that's what the trust
          score answers. If a manager reviews every purchase, the score is redundant; the
          manager is the check. It matters the moment a company wants the agent acting
          on its own, without someone watching every transaction.
        </p>

        <p>
          For a single founder running one agent on their own funds, watching closely — this
          mostly doesn't apply. That's a fair objection, and it's fine to say so.
        </p>

        <p>
          It matters at any real headcount: a finance team that didn't build the agent it's
          now letting spend company money, a company running several agents across
          departments that needs one shared signal to calibrate trust across all of them, an
          auditor asking why a given purchase was allowed. Same reason employee spending
          controls and credit checks exist at all — nobody needs oversight to spend their own
          paycheck. Scores exist for the person who has to answer for someone else's spending.
        </p>

        <p>
          One thing still holds even for a single well-behaved agent: if it gets compromised,
          or just starts behaving differently after a long clean run, a dropping score is an
          early warning a budget cap alone can't give you — a procurement policy only checks
          the ceiling, never whether the behavior underneath it changed.
        </p>
      </div>
    </div>
  );
}
