import { useCallback, useMemo, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { listApplications } from "../api/applications";
import { Application, ApplicationStatus } from "../types";
import { useAuth } from "../providers/AuthProvider";
import { useRealtime } from "../hooks/useRealtime";
import { toast } from "sonner";
import { RequestTable } from "../components/requests/RequestTable";
import { CreateRequestForm } from "../components/requests/CreateRequestForm";
import { RequestDetailDrawer } from "../components/requests/RequestDetailDrawer";
import { Header, PageLayout, Sidebar } from "../components/Layout";

const PAGE_SIZE_OPTIONS = [10, 20, 50];

export default function DashboardPage() {
  const { token } = useAuth();
  const qc = useQueryClient();

  const [filters, setFilters] = useState<{
    country: string;
    status: ApplicationStatus | "";
    from_date: string;
    to_date: string;
    limit: number;
  }>({
    country: "",
    status: "",
    from_date: "",
    to_date: "",
    limit: 20,
  });
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<Application | null>(null);
  const [highlightId, setHighlightId] = useState<string | null>(null);

  const offset = (page - 1) * filters.limit;

  const queryKey = useMemo(
    () => ["applications", filters, page],
    [filters, page]
  );

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () => listApplications({ ...filters, offset }),
  });

  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / filters.limit));

  const handleRealtime = useCallback(
    (msg: any) => {
      if (msg?.application_id) {
        setHighlightId(msg.application_id);
        qc.invalidateQueries({ queryKey: ["applications"] });
        toast.success(`Solicitud ${msg.application_id} actualizada`, {
          id: `update-${msg.application_id}`,
          duration: 2400,
          position: "bottom-right",
        });
        setTimeout(() => setHighlightId(null), 1600);
      }
    },
    [qc]
  );

  const connected = useRealtime(token, handleRealtime);

  const handleFilterSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    setFilters({
      country: (fd.get("country") as string) || "",
      status: ((fd.get("status") as string) || "") as ApplicationStatus | "",
      from_date: (fd.get("from_date") as string) || "",
      to_date: (fd.get("to_date") as string) || "",
      limit: Number(fd.get("limit")) || 20,
    });
    setPage(1); // volver a primera página al cambiar filtros
  };

  return (
    <PageLayout sidebar={<Sidebar />} header={<Header connected={connected} />}>
      <div className="grid gap-4 lg:grid-cols-[360px_1fr]">
        <div className="space-y-4">
          <CreateRequestForm />

          <div className="card p-4">
            <div className="mb-3">
              <p className="text-xs uppercase text-muted">Filtros</p>
              <h3 className="text-lg font-semibold">Listado</h3>
            </div>
            <form className="grid grid-cols-1 gap-3" onSubmit={handleFilterSubmit}>
              <div>
                <label className="label">País</label>
                <input name="country" className="input" defaultValue={filters.country} placeholder="MX" />
              </div>
              <div>
                <label className="label">Estado</label>
                <select name="status" className="input" defaultValue={filters.status}>
                  <option value="">Todos</option>
                  <option value="PENDING">PENDING</option>
                  <option value="VALIDATING">VALIDATING</option>
                  <option value="APPROVED">APPROVED</option>
                  <option value="DENIED">DENIED</option>
                </select>
              </div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="label">From</label>
                  <input type="date" name="from_date" className="input" defaultValue={filters.from_date} />
                </div>
                <div>
                  <label className="label">To</label>
                  <input type="date" name="to_date" className="input" defaultValue={filters.to_date} />
                </div>
              </div>
              <div>
                <label className="label">Por página</label>
                <select name="limit" className="input" defaultValue={filters.limit}>
                  {PAGE_SIZE_OPTIONS.map((n) => (
                    <option key={n} value={n}>{n} resultados</option>
                  ))}
                </select>
              </div>
              <button className="btn btn-ghost" type="submit">Aplicar filtros</button>
            </form>
          </div>
        </div>

        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs uppercase text-muted">Solicitudes</p>
              <h2 className="text-xl font-semibold">Resumen</h2>
            </div>
            {total > 0 && (
              <span className="text-sm text-muted">
                {offset + 1}–{Math.min(offset + filters.limit, total)} de {total}
              </span>
            )}
          </div>

          <RequestTable
            data={data?.applications || []}
            isLoading={isLoading}
            onSelect={(app) => setSelected(app)}
            highlightId={highlightId}
          />

          {totalPages > 1 && (
            <div className="flex items-center justify-between rounded-lg border border-border bg-card px-4 py-2">
              <button
                className="btn btn-ghost text-sm"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                ← Anterior
              </button>
              <span className="text-sm text-muted">
                Página {page} de {totalPages}
              </span>
              <button
                className="btn btn-ghost text-sm"
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
              >
                Siguiente →
              </button>
            </div>
          )}
        </div>
      </div>

      <RequestDetailDrawer app={selected} onClose={() => setSelected(null)} />
    </PageLayout>
  );
}
