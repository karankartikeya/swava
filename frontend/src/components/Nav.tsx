import { NavLink } from "react-router-dom";

const links = [
  { to: "/", label: "Home" },
  { to: "/how-it-works", label: "How It Works" },
  { to: "/demo", label: "Live Demo" },
  { to: "/findings", label: "Findings" },
];

export function Nav() {
  return (
    <header className="sticky top-4 z-20 mx-auto max-w-4xl px-4">
      <div className="flex items-center justify-between gap-4 rounded-[16px] border border-line bg-paper px-4 py-2.5">
        <NavLink to="/" className="flex items-center gap-2.5">
          <LogoMark />
          <span className="font-body text-lg font-bold tracking-tight text-ink">
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
                `rounded-[6px] px-3.5 py-1.5 text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-mint text-ink"
                    : "text-ink hover:bg-whisper"
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

function LogoMark() {
  return (
    <span
      className="flex h-8 w-8 shrink-0 items-center justify-center rounded-[6px] bg-highlight font-body text-sm font-bold text-ink"
      aria-hidden="true"
    >
      sw
    </span>
  );
}
