import { client } from "./client";
import { Application, ListApplicationsResponse, ApplicationStatus } from "../types";

export interface ListFilters {
  country?: string;
  status?: ApplicationStatus | "";
  from_date?: string;
  to_date?: string;
  limit?: number;
  offset?: number;
}

export async function listApplications(filters: ListFilters) {
  const params: Record<string, string | number> = {};
  Object.entries(filters).forEach(([k, v]) => {
    if (v !== undefined && v !== "") params[k] = v as any;
  });
  const res = await client.get<ListApplicationsResponse>("/api/v1/applications", { params });
  return res.data;
}

export async function createApplication(payload: {
  country: string;
  full_name: string;
  identity_document: string;
  monthly_income: number;
  requested_amount: number;
}) {
  const res = await client.post<Application>("/api/v1/applications", payload, {
    headers: { "Idempotency-Key": crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36) },
  });
  return res.data;
}

export async function getApplication(id: string) {
  const res = await client.get<Application>(`/api/v1/applications/${id}`);
  return res.data;
}

export async function updateStatus(id: string, status: ApplicationStatus, notes?: string | null) {
  const res = await client.put(`/api/v1/applications/${id}`, { status, notes: notes || null }, {
    headers: { "Idempotency-Key": crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36) },
  });
  return res.data;
}
