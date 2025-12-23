import { getApiBaseUrl } from "@/lib/api";

export function ApiStatusCard() {
  const apiBaseUrl = getApiBaseUrl();

  return (
    <div className="rounded-2xl border border-slate-800 bg-gradient-to-br from-slate-900 via-slate-950 to-slate-900 p-6">
      <h2 className="text-xl font-semibold">Integração REST</h2>
      <p className="mt-3 text-sm text-slate-300">
        Todas as chamadas do frontend devem passar por
        <span className="font-semibold text-white"> src/lib/api.ts</span>.
        Assim fica fácil centralizar headers e autenticação.
      </p>
      <div className="mt-6 rounded-xl border border-slate-800 bg-slate-950 p-4">
        <p className="text-xs uppercase tracking-[0.2em] text-slate-500">
          Base URL atual
        </p>
        <p className="mt-2 break-all text-sm font-semibold text-emerald-300">
          {apiBaseUrl}
        </p>
      </div>
    </div>
  );
}
