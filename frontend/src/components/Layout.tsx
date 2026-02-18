import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../providers/AuthProvider";
import { Signal, LogOut, Menu, Search } from "lucide-react";
import { useState } from "react";

export function Header({ connected }: { connected: boolean }) {
  return (
    <header className="flex items-center justify-between border-b border-border bg-card/60 px-4 py-3 backdrop-blur">
      <div className="flex items-center gap-3 text-sm text-muted">
        <div className={`h-3 w-3 rounded-full ${connected ? "bg-accent" : "bg-danger"}`} />
        <span>{connected ? "Realtime activo" : "Sin conexión"}</span>
      </div>
      <div className="flex items-center gap-2 rounded-lg border border-border bg-[#0d1526] px-3 py-2 text-sm text-muted">
        <Search size={16} />
        <input className="bg-transparent outline-none text-white" placeholder="Buscar (WIP)" disabled />
      </div>
    </header>
  );
}

export function Sidebar() {
  const { logout, userEmail } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);

  const navItem = (to: string, label: string) => (
    <Link
      to={to}
      className={`block rounded-lg px-3 py-2 text-sm font-semibold transition hover:bg-accent/10 ${
        location.pathname === to ? "bg-accent/15 text-white" : "text-muted"
      }`}
      onClick={() => setOpen(false)}
    >
      {label}
    </Link>
  );

  return (
    <aside className={`border-r border-border bg-card p-4 transition md:translate-x-0 ${open ? "translate-x-0" : "-translate-x-full md:block"}`}>
      <div className="flex items-center justify-between pb-4">
        <div>
          <div className="text-xs uppercase text-muted">Usuario</div>
          <div className="text-sm font-semibold">{userEmail || "-"}</div>
        </div>
        <button className="md:hidden btn btn-ghost px-2 py-1" onClick={() => setOpen(!open)}>
          <Menu size={16} />
        </button>
      </div>
      <div className="space-y-2">
        {navItem("/app", "Solicitudes")}
      </div>
      <div className="mt-6 flex items-center gap-2 rounded-lg border border-border px-3 py-2 text-xs text-muted">
        <Signal size={14} className="text-accent" />
        Bravo Dashboard
      </div>
      <button
        className="mt-6 flex w-full items-center justify-center gap-2 rounded-lg border border-border px-3 py-2 text-sm font-semibold text-muted transition hover:text-white"
        onClick={() => {
          logout();
          navigate("/login");
        }}
      >
        <LogOut size={16} />
        Cerrar sesión
      </button>
    </aside>
  );
}

export function PageLayout({ sidebar, header, children }: { sidebar: React.ReactNode; header: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-bg text-white">
      <div className="grid min-h-screen grid-cols-[260px_1fr] md:grid-cols-[260px_1fr]">
        {sidebar}
        <div className="flex min-h-screen flex-col">
          {header}
          <main className="flex-1 bg-gradient-to-br from-bg via-[#0f172a] to-bg p-4 md:p-6">
            {children}
          </main>
        </div>
      </div>
    </div>
  );
}
