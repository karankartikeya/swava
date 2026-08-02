import type { ReactNode } from "react";
import type { Decision } from "../lib/api";
import { IconCheck, IconPause, IconBlock } from "./icons";

export function Section({
  children,
  className = "",
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <section className={`mx-auto max-w-[1200px] px-6 ${className}`}>{children}</section>
  );
}

export function Eyebrow({ children }: { children: ReactNode }) {
  return (
    <div className="inline-flex items-center gap-1.5 rounded-full bg-highlight px-3 py-1 font-body text-xs font-medium uppercase tracking-wide text-ink">
      {children}
    </div>
  );
}

const decisionMeta: Record<
  Decision,
  { label: string; icon: ReactNode; text: string; bg: string; border: string }
> = {
  auto_approve: {
    label: "Auto-approved",
    icon: <IconCheck />,
    text: "text-approve-ink",
    bg: "bg-approve-soft",
    border: "border-approve",
  },
  human_review: {
    label: "Manager approval required",
    icon: <IconPause />,
    text: "text-review-ink",
    bg: "bg-review-soft",
    border: "border-review",
  },
  blocked: {
    label: "Blocked",
    icon: <IconBlock />,
    text: "text-block-ink",
    bg: "bg-block-soft",
    border: "border-block",
  },
};

export function DecisionBadge({ decision }: { decision: Decision }) {
  const m = decisionMeta[decision];
  return (
    <span
      className={`inline-flex items-center gap-2 rounded-[6px] border px-3.5 py-1.5 text-sm font-medium ${m.text} ${m.bg} ${m.border}`}
    >
      <span className="h-4 w-4">{m.icon}</span>
      {m.label}
    </span>
  );
}

export function decisionMetaFor(decision: Decision) {
  return decisionMeta[decision];
}
