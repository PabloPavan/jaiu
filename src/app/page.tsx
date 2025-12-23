import Link from "next/link";
import { ApiStatusCard } from "@/components/api-status-card";

export default function HomePage() {
  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-10 px-6 py-12">
      <section className="flex flex-col gap-4">
        <p className="text-sm uppercase tracking-[0.2em] text-slate-400">
          jaiu frontend
        </p>
        <h1 className="text-4xl font-semibold text-white sm:text-5xl">
          Estrutura base do MVP com Next.js, TypeScript e TailwindCSS
        </h1>
        <p className="max-w-2xl text-base text-slate-300">
          Pronto para conectar ao backend via REST, que será mantido em outro
          repositório. Ajuste o baseURL em
          <span className="font-semibold text-white"> .env </span> e comece a
          consumir endpoints com segurança e tipagem.
        </p>
        <div className="flex flex-wrap gap-3">
          <Link
            className="rounded-full bg-white px-6 py-2 text-sm font-semibold text-slate-900 hover:bg-slate-200"
            href="/"
          >
            Dashboard
          </Link>
          <Link
            className="rounded-full border border-slate-700 px-6 py-2 text-sm font-semibold text-white hover:border-slate-500"
            href="/"
          >
            Documentação
          </Link>
        </div>
      </section>

      <section className="grid gap-6 md:grid-cols-2">
        <div className="rounded-2xl border border-slate-800 bg-slate-900/50 p-6">
          <h2 className="text-xl font-semibold">Próximos passos</h2>
          <ul className="mt-4 space-y-3 text-sm text-slate-300">
            <li>• Configurar rotas privadas e autenticação.</li>
            <li>• Adicionar layouts para área administrativa.</li>
            <li>• Criar componentes de formulário com validação.</li>
            <li>• Conectar endpoints com hooks assíncronos.</li>
          </ul>
        </div>
        <ApiStatusCard />
      </section>
    </main>
  );
}
