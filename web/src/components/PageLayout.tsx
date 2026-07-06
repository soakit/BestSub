import type { ReactNode } from "react";

export function PageLayout({ title, actions, children, className }: { title: string; actions?: ReactNode; children: ReactNode; className?: string }) {
  return (
    <div className="flex flex-col gap-2 h-full">
      <div className="flex-none flex justify-between items-center min-h-10">
        <h1 className="text-lg font-bold tracking-tight">{title}</h1>
        {actions && <div className="flex items-center gap-2">{actions}</div>}
      </div>
      <div className={"w-full flex-1 overflow-y-auto min-h-0 rounded-t-2xl bg-surface-primary pb-24 md:pb-3" + (className ? " " + className : "")}>
        {children}
      </div>
    </div>
  );
}
