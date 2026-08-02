import type { ReactNode } from "react";

export interface PipelineStage {
  key: string;
  label: string;
  detail?: string;
  icon: ReactNode;
  /** Which side of the transaction this stage concerns — used to color-code
   * "agent" stages distinctly from "merchant" stages, resolving the most
   * common point of confusion: it's the buying agent's wallet that gets
   * scored, never the merchant's. */
  side?: "agent" | "merchant" | "neutral";
  state?: "idle" | "active" | "done" | "skipped";
}

const sideColor: Record<NonNullable<PipelineStage["side"]>, string> = {
  agent: "border-signal bg-mint text-ink",
  merchant: "border-line-strong bg-paper-raised text-ink",
  neutral: "border-line-strong bg-paper-raised text-ink",
};

export function Pipeline({
  stages,
  compact = false,
}: {
  stages: PipelineStage[];
  compact?: boolean;
}) {
  return (
    <div
      className={`flex ${compact ? "gap-2" : "gap-3"} overflow-x-auto pb-2`}
      role="list"
      aria-label="Purchase decision pipeline"
    >
      {stages.map((stage, i) => (
        <div key={stage.key} className="flex items-center" role="listitem">
          <div
            className={`flex shrink-0 flex-col items-center gap-2 rounded-[12px] border px-4 py-3.5 transition-colors ${
              stage.state === "active"
                ? "border-signal bg-signal text-paper shadow-[0_0_0_4px_var(--signal-soft)]"
                : stage.state === "done"
                  ? sideColor[stage.side ?? "neutral"]
                  : stage.state === "skipped"
                    ? "border-line bg-paper text-ink opacity-50"
                    : "border-line bg-paper text-ink"
            } ${compact ? "min-w-[92px]" : "min-w-[132px]"}`}
          >
            <div className={compact ? "h-5 w-5" : "h-6 w-6"}>{stage.icon}</div>
            <div className="text-center">
              <div className={`font-medium ${compact ? "text-[11px]" : "text-xs"}`}>
                {stage.label}
              </div>
              {stage.detail && !compact && (
                <div className="mt-0.5 text-[11px] opacity-75">{stage.detail}</div>
              )}
            </div>
          </div>
          {i < stages.length - 1 && (
            <svg
              width={compact ? "16" : "22"}
              height="10"
              viewBox="0 0 22 10"
              fill="none"
              className="shrink-0 text-line-strong"
              aria-hidden="true"
            >
              <path
                d="M0 5H19M19 5L14 1M19 5L14 9"
                stroke="currentColor"
                strokeWidth="1.4"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          )}
        </div>
      ))}
    </div>
  );
}
