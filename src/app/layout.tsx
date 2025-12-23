import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Jaiu MVP",
  description: "MVP web para integração com backend REST"
};

export default function RootLayout({
  children
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="pt-BR">
      <body>
        <div className="min-h-screen bg-slate-950 text-slate-100">
          {children}
        </div>
      </body>
    </html>
  );
}
