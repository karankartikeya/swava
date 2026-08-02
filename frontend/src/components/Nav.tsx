import { NavLink } from "react-router-dom";

const links = [
  { to: "/", label: "Home" },
  { to: "/how-it-works", label: "How It Works" },
  { to: "/demo", label: "Live Demo" },
  { to: "/findings", label: "Findings" },
];

export function Nav() {
  return (
    <header className="border-b border-line">
      <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-5">
        <NavLink to="/" className="flex items-center gap-2.5">
          <MarkIcon />
          <span className="font-display text-lg font-semibold tracking-tight text-ink">
            Swava
          </span>
        </NavLink>
        <nav className="flex items-center gap-1">
          {links.map((l) => (
            <NavLink
              key={l.to}
              to={l.to}
              end={l.to === "/"}
              className={({ isActive }) =>
                `rounded-full px-4 py-2 text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-signal-soft text-signal-ink"
                    : "text-ink-soft hover:text-ink"
                }`
              }
            >
              {l.label}
            </NavLink>
          ))}
        </nav>
      </div>
    </header>
  );
}

function MarkIcon() {
  return (
    <svg width="22" height="22" viewBox="0 0 22 22" fill="none" aria-hidden="true">
      <rect x="1" y="1" width="20" height="20" rx="5" stroke="var(--signal)" strokeWidth="1.6" />
      <circle cx="7" cy="11" r="2.1" fill="var(--signal)" />
      <path d="M9.6 11H15.5" stroke="var(--signal)" strokeWidth="1.6" strokeLinecap="round" />
      <path d="M13.2 8.4L15.8 11L13.2 13.6" stroke="var(--signal)" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
