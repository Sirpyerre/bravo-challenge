export type ApplicationStatus = "PENDING" | "VALIDATING" | "APPROVED" | "DENIED";

export interface Application {
  id: string;
  user_id?: string;
  country: string;
  full_name: string;
  identity_document: string;
  monthly_income: number;
  requested_amount: number;
  status: ApplicationStatus;
  risk_level?: string | null;
  notes?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ListApplicationsResponse {
  applications: Application[];
  total: number;
}
