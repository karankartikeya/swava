import type { ReactNode } from "react";
import { Nav } from "./Nav";

export function Layout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <Nav />
      <main className="flex-1 pt-10">{children}</main>
      <Footer />
    </div>
  );
}

function Footer() {
  return (
    <footer className="border-t border-line">
      <div className="mx-auto flex max-w-4xl flex-col gap-1 px-6 py-8 text-sm text-ink sm:flex-row sm:items-center sm:justify-between">
        <span>Swava — an AI Procurement Manager for company-funded agents.</span>
        <span className="font-mono text-xs text-ink">Built for a hackathon. Sandbox environment.</span>
      </div>
    </footer>
  );
}
