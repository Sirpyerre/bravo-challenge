import { Application } from "../../types";
import { clsx } from "clsx";

interface Props {
  data: Application[];
  isLoading: boolean;
  onSelect: (app: Application) => void;
  highlightId?: string | null;
}

export function RequestTable({ data, isLoading, onSelect, highlightId }: Props) {
  if (isLoading) return <div className="text-sm text-muted">Cargando solicitudes...</div>;
  if (!data?.length) return <div className="text-sm text-muted">Sin solicitudes.</div>;

  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <table className="table">
        <thead className="bg-card">
          <tr>
            <th>País</th>
            <th>Nombre</th>
            <th>Estado</th>
            <th>Monto</th>
            <th>Ingreso</th>
            <th>Creado</th>
          </tr>
        </thead>
        <tbody>
          {data.map((app) => (
            <tr
              key={app.id}
              onClick={() => onSelect(app)}
              className={clsx("cursor-pointer transition", {
                "animate-flash": highlightId === app.id,
              })}
            >
              <td className="font-medium">{app.country}</td>
              <td>{app.full_name}</td>
              <td>
                <span className={clsx("badge", app.status)}>{app.status}</span>
              </td>
              <td>${app.requested_amount.toLocaleString()}</td>
              <td>${app.monthly_income.toLocaleString()}</td>
              <td>{new Date(app.created_at).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
