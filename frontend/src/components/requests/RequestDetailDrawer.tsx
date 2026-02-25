import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Application, ApplicationStatus } from "../../types";
import { updateStatus } from "../../api/applications";
import { toast } from "sonner";

const statuses: ApplicationStatus[] = ["PENDING", "VALIDATING", "APPROVED", "DENIED"];

export function RequestDetailDrawer({ app, onClose }: { app: Application | null; onClose: () => void }) {
  const qc = useQueryClient();
  const [copied, setCopied] = useState(false);

  const handleCopyId = () => {
    if (!app) return;
    if (navigator.clipboard) {
      navigator.clipboard.writeText(app.id);
    } else {
      const el = document.createElement("textarea");
      el.value = app.id;
      el.style.position = "fixed";
      el.style.opacity = "0";
      document.body.appendChild(el);
      el.select();
      document.execCommand("copy");
      document.body.removeChild(el);
    }
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  const { mutateAsync, isPending } = useMutation({
    mutationFn: async (payload: { status: ApplicationStatus; notes?: string | null }) => {
      if (!app) return;
      return updateStatus(app.id, payload.status, payload.notes);
    },
    onSuccess: () => {
      toast.success("Estado actualizado", {
        id: app ? `update-${app.id}` : undefined,
        duration: 2400,
        position: "bottom-right",
      });
      qc.invalidateQueries({ queryKey: ["applications"] });
    },
    onError: (err: any) => toast.error(err?.response?.data?.message || "Error al actualizar"),
  });

  if (!app) return null;

  const handleUpdate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const status = fd.get("status") as ApplicationStatus;
    const notes = (fd.get("notes") as string) || null;
    await mutateAsync({ status, notes });
  };

  return (
    <div className="drawer">
      <div className="flex items-center justify-between border-b border-border px-4 py-3">
        <div>
          <p className="text-xs uppercase text-muted">Solicitud</p>
          <h3 className="text-lg font-semibold">{app.full_name}</h3>
        </div>
        <button className="btn btn-ghost" onClick={onClose}>Cerrar</button>
      </div>

      <div className="mx-4 mt-3 flex w-[calc(100%-2rem)] items-center justify-between gap-2 rounded-md border border-border bg-card px-3 py-2">
        <span className="select-all break-all font-mono text-xs text-muted">{app.id}</span>
        <button
          onClick={handleCopyId}
          title="Copiar ID"
          className="shrink-0 text-[10px] uppercase tracking-wide text-muted transition hover:text-foreground"
        >
          {copied ? "✓ Copiado" : "Copiar"}
        </button>
      </div>

      <div className="space-y-4 p-4">
        <div className="grid grid-cols-2 gap-3 text-sm text-muted">
          <div><span className="label">País</span><div>{app.country}</div></div>
          <div><span className="label">Estado</span><div><span className={`badge ${app.status}`}>{app.status}</span></div></div>
          <div><span className="label">Monto</span><div>${app.requested_amount.toLocaleString()}</div></div>
          <div><span className="label">Ingreso</span><div>${app.monthly_income.toLocaleString()}</div></div>
          <div><span className="label">Documento</span><div>{app.identity_document}</div></div>
          <div><span className="label">Creado</span><div>{new Date(app.created_at).toLocaleString()}</div></div>
        </div>

        <form className="space-y-3" onSubmit={handleUpdate}>
          <div>
            <label className="label">Nuevo estado</label>
            <select name="status" className="input" defaultValue={app.status} key={app.id}>
              {statuses.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="label">Notas</label>
            <textarea name="notes" className="input" rows={3} defaultValue={app.notes || ""} />
          </div>
          <button className="btn btn-primary w-full" type="submit" disabled={isPending}>
            {isPending ? "Guardando..." : "Guardar cambios"}
          </button>
        </form>
      </div>
    </div>
  );
}
