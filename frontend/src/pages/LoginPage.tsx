import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../providers/AuthProvider";
import { toast } from "sonner";

export default function LoginPage() {
  const { login, register } = useAuth();
  const [isRegister, setIsRegister] = useState(false);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const email = ((form.get("email") ?? "") as string).trim();
    const password = ((form.get("password") ?? "") as string).trim();
    const country = isRegister ? ((form.get("country") as string) || "MX") : "MX";
    const role = isRegister ? ((form.get("role") as string) || "USER") : "USER";

    if (!email || !password) {
      toast.error("Email y password son requeridos");
      return;
    }

    setLoading(true);
    try {
      if (isRegister) {
        await register(email, password, country, role);
        toast.success("Registro exitoso");
      } else {
        await login(email, password);
      }
      navigate("/app");
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Error de autenticación");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-bg via-[#0f172a] to-bg p-4">
      <div className="w-full max-w-md space-y-4 rounded-2xl border border-border bg-card p-6 shadow-card">
        <div>
          <p className="text-xs uppercase text-muted">Bravo</p>
          <h1 className="text-2xl font-bold">{isRegister ? "Crear cuenta" : "Iniciar sesión"}</h1>
          <p className="text-sm text-muted">JWT + sesión persistente + rutas protegidas</p>
        </div>

        <form className="space-y-3" onSubmit={handleSubmit}>
          <div className="space-y-1">
            <label className="label">Email</label>
            <input className="input" name="email" type="email" required placeholder="dev@bravo.test" />
          </div>
          <div className="space-y-1">
            <label className="label">Password</label>
            <input className="input" name="password" type="password" required placeholder="••••••" />
          </div>
          {isRegister && (
            <>
              <div className="space-y-1">
                <label className="label">País</label>
                <select className="input" name="country" defaultValue="MX">
                  <option value="MX">🇲🇽 México (MX)</option>
                  <option value="BR">🇧🇷 Brasil (BR)</option>
                </select>
              </div>
              <div className="space-y-1">
                <label className="label">Rol</label>
                <select className="input" name="role" defaultValue="USER">
                  <option value="USER">Usuario</option>
                  <option value="AGENT">Agente</option>
                </select>
              </div>
            </>
          )}

          <button className="btn btn-primary w-full" type="submit" disabled={loading}>
            {loading ? "Procesando..." : isRegister ? "Registrarme" : "Entrar"}
          </button>
        </form>

        <button
          className="btn btn-ghost w-full"
          onClick={() => setIsRegister((p) => !p)}
          type="button"
        >
          {isRegister ? "Ya tengo cuenta" : "Crear nueva cuenta"}
        </button>
      </div>
    </div>
  );
}
