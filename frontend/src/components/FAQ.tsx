import { Eyebrow } from "./ui";

/** Shared across How It Works and Live Demo — anticipates the sharpest
 * objection a self-funded, single-operator visitor has: why limit an agent
 * that's only ever spending your own money under your own eye. */
export function FAQ() {
  return (
    <div className="rounded-3xl border border-line bg-paper-raised p-8 sm:p-10">
      <Eyebrow>Before you ask</Eyebrow>
      <h2 className="mt-2 max-w-xl font-display text-2xl font-semibold leading-snug text-ink">
        Why would I want to limit my own agent?
      </h2>

      <div className="mt-6 flex flex-col gap-4 text-[15px] leading-relaxed text-ink-soft">
        <p>
          Two different questions get conflated here. <strong className="text-ink">Does the agent
          have money to spend</strong> — that's Prava's job, you already set the mandate and the
          ceiling. <strong className="text-ink">Should it be trusted to spend it right now,
          without you personally checking</strong> — that's what the score answers. If a human
          reviews every transaction, the score is redundant; the human is the check. It matters
          the moment you want the agent acting while you're not watching.
        </p>

        <p>
          For one person running their own agent, on their own funds, watching closely — this
          mostly doesn't apply. That's a fair objection, and it's fine to say so.
        </p>

        <p>
          It matters somewhere else: an insurer underwriting an agent it's never dealt with, a
          marketplace receiving a purchase request from an unfamiliar wallet, a company running
          agents it didn't personally build and needs one shared signal to calibrate trust
          across. Same reason human credit scores exist — nobody needs a FICO score to spend
          their own paycheck. Scores exist for the stranger deciding whether to trust you.
        </p>

        <p>
          One thing still holds even in the self-funded case: if the agent gets compromised, or
          just starts behaving differently after a long clean run, a dropping score is an early
          warning a spend limit alone can't give you — a mandate only checks the ceiling, never
          whether the behavior underneath it changed.
        </p>
      </div>
    </div>
  );
}
