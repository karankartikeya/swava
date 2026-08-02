const common = {
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.6,
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
};

export function IconTask() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <path d="M6 4h9l4 4v12a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1Z" />
      <path d="M14 4v4h4" />
      <path d="M8 13h8M8 16.5h5" />
    </svg>
  );
}

export function IconSearch() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <circle cx="10.5" cy="10.5" r="6.5" />
      <path d="M20 20l-4.8-4.8" />
    </svg>
  );
}

export function IconWallet() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <path d="M3 8a2 2 0 0 1 2-2h11.5a1.5 1.5 0 0 1 1.5 1.5V8" />
      <path d="M3 8v10a2 2 0 0 0 2 2h13.5a1.5 1.5 0 0 0 1.5-1.5v-2" />
      <path d="M20 12.5h-3.2a2 2 0 0 0 0 4H20a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1Z" />
    </svg>
  );
}

export function IconGauge() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <path d="M3.5 16.5a8.5 8.5 0 1 1 17 0" strokeWidth={1.8} />
      <path d="M12 16.5l4-5.5" strokeWidth={1.8} />
      <circle cx="12" cy="16.5" r="1.3" fill="currentColor" stroke="none" />
      <path d="M3.5 16.5h1.8M18.7 16.5h1.8" strokeWidth={1.8} />
    </svg>
  );
}

export function IconGate() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <path d="M4 21V7l8-4 8 4v14" />
      <path d="M4 21h16" />
      <path d="M9 21v-6h6v6" />
    </svg>
  );
}

export function IconCard() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <rect x="3" y="5" width="18" height="14" rx="2" />
      <path d="M3 10h18" />
      <path d="M7 14.5h4" />
    </svg>
  );
}

export function IconCheck() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <path d="M4 12.5l5 5L20 6" />
    </svg>
  );
}

export function IconPause() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <rect x="6" y="4" width="4" height="16" rx="1" />
      <rect x="14" y="4" width="4" height="16" rx="1" />
    </svg>
  );
}

export function IconBlock() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <circle cx="12" cy="12" r="9" />
      <path d="M6.3 6.3l11.4 11.4" />
    </svg>
  );
}

export function IconAgent() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <rect x="4" y="7" width="16" height="12" rx="2.5" />
      <path d="M12 7V3" />
      <circle cx="12" cy="2" r="1" fill="currentColor" stroke="none" />
      <circle cx="9" cy="13" r="1.2" fill="currentColor" stroke="none" />
      <circle cx="15" cy="13" r="1.2" fill="currentColor" stroke="none" />
      <path d="M9 16.5h6" />
    </svg>
  );
}

export function IconStore() {
  return (
    <svg viewBox="0 0 24 24" {...common} width="100%" height="100%">
      <path d="M4 9l1-5h14l1 5" />
      <path d="M4 9a2 2 0 0 0 4 0 2 2 0 0 0 4 0 2 2 0 0 0 4 0 2 2 0 0 0 4 0" />
      <path d="M5 9v10h14V9" />
      <path d="M10 19v-6h4v6" />
    </svg>
  );
}
