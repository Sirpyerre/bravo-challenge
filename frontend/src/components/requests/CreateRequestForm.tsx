import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createApplication } from "../../api/applications";
import { toast } from "sonner";

export function CreateRequestForm() {
  const qc = useQueryClient();
  const initialForm = {
    country: "MX",
    full_name: "",
    identity_document: "",
    monthly_income: "",
    requested_amount: "",
  };
  const [form, setForm] = useState(initialForm);

  const monthly = Number(form.monthly_income);
  const requested = Number(form.requested_amount);

  const isInvalid =
    !form.country.trim() ||
    !form.full_name.trim() ||
    !form.identity_document.trim() ||
    Number.isNaN(monthly) ||
    Number.isNaN(requested) ||
    monthly <= 0 ||
    requested <= 0;

  const { mutateAsync, isPending } = useMutation({
    mutationFn: async () => {
      return createApplication({
        country: form.country,
        full_name: form.full_name,
        identity_document: form.identity_document,
        monthly_income: Number(form.monthly_income),
        requested_amount: Number(form.requested_amount),
      });
    },
    onSuccess: (app) => {
      toast.success(`Solicitud ${app.id} creada`, {
        id: `create-${app.id}`,
        duration: 2600,
      });
      qc.invalidateQueries({ queryKey: ["applications"] });
      setForm(initialForm);
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Error al crear");
    },
  });

  return (
    <div className="card p-4">
      <div className="mb-3">
        <p className="text-xs uppercase text-muted">Nueva solicitud</p>
        <h3 className="text-lg font-semibold">Crear</h3>
      </div>
      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
        <div>
          <label className="label">País</label>
          <input
            className="input"
            value={form.country}
            onChange={(e) => setForm({ ...form, country: e.target.value })}
          />
        </div>
        <div>
          <label className="label">Nombre completo</label>
          <input
            className="input"
            value={form.full_name}
            onChange={(e) => setForm({ ...form, full_name: e.target.value })}
            placeholder="Juan Pérez"
          />
        </div>
        <div>
          <label className="label">Documento</label>
          <input
            className="input"
            value={form.identity_document}
            onChange={(e) => setForm({ ...form, identity_document: e.target.value })}
            placeholder="CURP / CPF"
          />
        </div>
        <div>
          <label className="label">Ingreso mensual</label>
          <input
            className="input"
            type="number"
            value={form.monthly_income}
            onChange={(e) => setForm({ ...form, monthly_income: e.target.value })}
          />
        </div>
        <div>
          <label className="label">Monto solicitado</label>
          <input
            className="input"
            type="number"
            value={form.requested_amount}
            onChange={(e) => setForm({ ...form, requested_amount: e.target.value })}
          />
        </div>
      </div>
      <button
        className="btn btn-primary mt-4 w-full"
        disabled={isInvalid || isPending}
        onClick={() => mutateAsync()}
      >
        {isPending ? "Creando..." : "Crear"}
      </button>
    </div>
  );
}
