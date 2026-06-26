import type { ReactNode } from "react";

export function PageLayout({ title, actions, children }: { title: string; actions?: ReactNode; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-2 h-full">
      <div className="flex justify-between items-center">
        <h1 className="text-lg font-bold tracking-tight text-foreground">{title}</h1>
        {actions && <div className="flex items-center gap-2">{actions}</div>}
      </div>
      {children}
    </div>
  );
}
