import { useCallback, useMemo, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { listApplications } from "../api/applications";
import { Application } from "../types";
import { useAuth } from "../providers/AuthProvider";
import { useRealtime } from "../hooks/useRealtime";
import { toast } from "sonner";
import { RequestTable } from "../components/requests/RequestTable";
import { CreateRequestForm } from "../components/requests/CreateRequestForm";
import { RequestDetailDrawer } from "../components/requests/RequestDetailDrawer";
import { Header, PageLayout, Sidebar } from "../components/Layout";

export default function DashboardPage() {
  const { token } = useAuth();
  const qc = useQueryClient();
  const [filters, setFilters] = useState({ country: "", status: "", from_date: "", to_date: "", limit: 20, offset: 0 });
  const [selected, setSelected] = useState<Application | null>(null);
  const [highlightId, setHighlightId] = useState<string | null>(null);

  const queryKey = useMemo(() => ["applications", filters], [filters]);

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () => listApplications(filters),
  });

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
      status: (fd.get("status") as string) || "",
      from_date: (fd.get("from_date") as string) || "",
      to_date: (fd.get("to_date") as string) || "",
      limit: Number(fd.get("limit")) || 20,
      offset: Number(fd.get("offset")) || 0,
    });
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
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="label">Límite</label>
                  <input type="number" name="limit" className="input" defaultValue={filters.limit} />
                </div>
                <div>
                  <label className="label">Offset</label>
                  <input type="number" name="offset" className="input" defaultValue={filters.offset} />
                </div>
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
          </div>

          <RequestTable
            data={data?.applications || []}
            isLoading={isLoading}
            onSelect={(app) => setSelected(app)}
            highlightId={highlightId}
          />
        </div>
      </div>

      <RequestDetailDrawer app={selected} onClose={() => setSelected(null)} />
    </PageLayout>
  );
}
