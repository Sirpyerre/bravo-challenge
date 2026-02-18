import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "sonner";
import App from "./App";
import { AuthProvider } from "./providers/AuthProvider";
import "./index.css";

const queryClient = new QueryClient();

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </AuthProvider>
      <Toaster
        richColors
        closeButton
        position="top-right"
        offset={72}
        gap={8}
        visibleToasts={3}
        toastOptions={{
          style: { maxWidth: "360px" },
        }}
      />
    </QueryClientProvider>
  </React.StrictMode>
);
