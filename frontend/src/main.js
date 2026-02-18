const statusEl = document.getElementById("status");
const authInfo = document.getElementById("auth-info");
const baseUrlInput = document.getElementById("base-url");
const listContainer = document.getElementById("list");
const detailContainer = document.getElementById("detail");
const eventsContainer = document.getElementById("events");

let token = null;
let socket = null;

const $ = (id) => document.getElementById(id);

function apiBase() {
  return baseUrlInput.value.replace(/\/$/, "");
}

function setStatus(text) {
  statusEl.textContent = text;
}

function setAuthInfo(text) {
  authInfo.textContent = text;
}

function headers(extra = {}) {
  return {
    "Content-Type": "application/json",
    Authorization: token ? `Bearer ${token}` : undefined,
    ...extra,
  };
}

async function register(email, password, country) {
  const res = await fetch(`${apiBase()}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, country }),
  });
  return res;
}

async function login(email, password) {
  const res = await fetch(`${apiBase()}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  return res;
}

function connectSocket() {
  if (!token) return;
  if (socket) socket.close();
  const wsUrl = apiBase().replace(/^http/, "ws") + `/ws?token=${token}`;
  socket = new WebSocket(wsUrl);
  socket.onopen = () => setStatus("Conectado (WS)");
  socket.onclose = () => setStatus("Desconectado");
  socket.onerror = () => setStatus("WS error");
  socket.onmessage = (evt) => {
    const msg = (() => {
      try {
        return JSON.parse(evt.data);
      } catch {
        return { raw: evt.data };
      }
    })();
    const el = document.createElement("div");
    el.className = "event-item";
    el.textContent = `${new Date().toLocaleTimeString()} :: ${JSON.stringify(msg)}`;
    eventsContainer.prepend(el);
  };
}

function renderList(apps = []) {
  if (!apps.length) {
    listContainer.innerHTML = "<p>No hay solicitudes.</p>";
    return;
  }
  const rows = apps
    .map(
      (a) => `<tr data-id="${a.id}">
        <td>${a.country}</td>
        <td>${a.full_name}</td>
        <td><span class="badge ${a.status}">${a.status}</span></td>
        <td>${a.requested_amount}</td>
        <td>${new Date(a.created_at).toLocaleString()}</td>
      </tr>`
    )
    .join("");
  listContainer.innerHTML = `<table><thead><tr><th>País</th><th>Nombre</th><th>Estado</th><th>Monto</th><th>Creado</th></tr></thead><tbody>${rows}</tbody></table>`;
  listContainer.querySelectorAll("tr[data-id]").forEach((row) => {
    row.addEventListener("click", () => loadDetail(row.dataset.id));
  });
}

function renderDetail(app) {
  if (!app) {
    detailContainer.textContent = "No encontrado";
    return;
  }
  detailContainer.innerHTML = `
    <div><strong>${app.full_name}</strong> (${app.country})</div>
    <div>Estado: <span class="badge ${app.status}">${app.status}</span></div>
    <div>Monto: ${app.requested_amount}</div>
    <div>Ingreso: ${app.monthly_income}</div>
    <div>Documento: ${app.identity_document}</div>
    <div>Creado: ${new Date(app.created_at).toLocaleString()}</div>
    <div>Notas: ${app.notes || "-"}</div>
  `;
  const idInput = document.querySelector('#status-form input[name="id"]');
  if (idInput) idInput.value = app.id;
}

async function listApplications(params = {}) {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== "") qs.append(k, v);
  });
  const res = await fetch(`${apiBase()}/api/v1/applications?${qs.toString()}`, {
    headers: headers(),
  });
  if (!res.ok) throw new Error(`List failed: ${res.status}`);
  const data = await res.json();
  renderList(data.applications || []);
}

async function loadDetail(id) {
  const res = await fetch(`${apiBase()}/api/v1/applications/${id}`, {
    headers: headers(),
  });
  if (!res.ok) {
    detailContainer.textContent = "No encontrado";
    return;
  }
  const data = await res.json();
  renderDetail(data);
}

function randomKey() {
  return crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36).slice(2);
}

// Event wiring

$("register-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const res = await register(fd.get("email"), fd.get("password"), fd.get("country") || "MX");
  if (res.ok) {
    const { token: t } = await res.json();
    token = t;
    setAuthInfo(`Token: ${t}`);
    connectSocket();
  } else {
    setAuthInfo(`Registro fallo: ${res.status}`);
  }
});

$("login-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const res = await login(fd.get("email"), fd.get("password"));
  if (res.ok) {
    const { token: t } = await res.json();
    token = t;
    setAuthInfo(`Token: ${t}`);
    connectSocket();
    listApplications();
  } else {
    setAuthInfo(`Login fallo: ${res.status}`);
  }
});

$("create-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  if (!token) return alert("Inicia sesión primero");
  const fd = new FormData(e.target);
  const payload = Object.fromEntries(fd.entries());
  payload.monthly_income = parseInt(payload.monthly_income, 10);
  payload.requested_amount = parseInt(payload.requested_amount, 10);
  const res = await fetch(`${apiBase()}/api/v1/applications`, {
    method: "POST",
    headers: headers({ "Idempotency-Key": `front-${randomKey()}` }),
    body: JSON.stringify(payload),
  });
  if (res.ok) {
    await listApplications();
  } else {
    alert(`Error al crear: ${res.status}`);
  }
});

$("filters").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const params = Object.fromEntries(fd.entries());
  try {
    await listApplications(params);
  } catch (err) {
    alert(err.message);
  }
});

$("status-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  if (!token) return alert("Inicia sesión primero");
  const fd = new FormData(e.target);
  const id = fd.get("id");
  const res = await fetch(`${apiBase()}/api/v1/applications/${id}`, {
    method: "PUT",
    headers: headers({ "Idempotency-Key": `status-${randomKey()}` }),
    body: JSON.stringify({ status: fd.get("status"), notes: fd.get("notes") || null }),
  });
  if (res.ok) {
    await loadDetail(id);
    await listApplications();
  } else {
    alert(`Error al actualizar: ${res.status}`);
  }
});

// Initial dummy state
renderList([]);
setStatus("Desconectado");

// Try to list without auth to show empty state if token is set later
// user will trigger list after login
