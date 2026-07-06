import { startTransition, useEffect, type ReactNode } from "react";
import { clsx } from "clsx";
import { preloadPage } from "../lib/pagePreload";
import { useAppStore, type Page, type NavItem, MAIN_NAV_ITEMS, SETTINGS_NAV_ITEM, NAV_ITEMS } from "../store";
import { Logo } from "./Logo";
import { useSettingList } from "../api/settings";

function selectPage(page: Page, setCurrentPage: (page: Page) => void) {
  preloadPage(page);
  startTransition(() => setCurrentPage(page));
}

function DesktopNavButton({
  item: { id, label, icon: Icon },
  active,
  onSelect,
}: {
  item: NavItem;
  active: boolean;
  onSelect: (page: Page) => void;
}) {
  return (
    <button
      type="button"
      onMouseEnter={() => preloadPage(id)}
      onFocus={() => preloadPage(id)}
      onClick={() => onSelect(id)}
      aria-current={active ? "page" : undefined}
      className={clsx(
        "flex w-full cursor-pointer items-center justify-start gap-3 rounded-2xl p-3",
        active
          ? "bg-default text-foreground font-medium"
          : "text-muted hover:bg-surface-tertiary hover:text-foreground"
      )}
    >
      <Icon className="h-5 w-5 flex-shrink-0" />
      <span>{label}</span>
    </button>
  );
}

function DesktopSidebar() {
  const currentPage = useAppStore((state) => state.currentPage);
  const setCurrentPage = useAppStore((state) => state.setCurrentPage);
  const handleSelect = (page: Page) => selectPage(page, setCurrentPage);

  return (
    <aside className="hidden h-full w-48 flex-none flex-col whitespace-nowrap px-4 py-6 md:flex">
      <div className="mb-6 flex items-center gap-2 py-2">
        <Logo />
        <span className="text-xl font-bold tracking-tight text-foreground">BESTSUB</span>
      </div>
      <nav className="flex h-full flex-col gap-1" aria-label="主导航">
        {MAIN_NAV_ITEMS.map((item) => (
          <DesktopNavButton
            key={item.id}
            item={item}
            active={currentPage === item.id}
            onSelect={handleSelect}
          />
        ))}
        <div className="mt-auto">
          <DesktopNavButton
            item={SETTINGS_NAV_ITEM}
            active={currentPage === SETTINGS_NAV_ITEM.id}
            onSelect={handleSelect}
          />
        </div>
      </nav>
    </aside>
  );
}

function MobileDockButton({
  item: { id, label, icon: Icon },
  active,
  onSelect,
}: {
  item: NavItem;
  active: boolean;
  onSelect: (page: Page) => void;
}) {
  return (
    <button
      type="button"
      onFocus={() => preloadPage(id)}
      onTouchStart={() => preloadPage(id)}
      onClick={() => onSelect(id)}
      aria-label={label}
      aria-current={active ? "page" : undefined}
      className={clsx(
        "flex h-12 w-12 flex-none cursor-pointer items-center justify-center rounded-2xl",
        active ? "bg-default text-foreground" : "text-muted"
      )}
    >
      <Icon className="h-5 w-5 flex-shrink-0" />
    </button>
  );
}

function MobileDock() {
  const currentPage = useAppStore((state) => state.currentPage);
  const setCurrentPage = useAppStore((state) => state.setCurrentPage);
  const handleSelect = (page: Page) => selectPage(page, setCurrentPage);

  return (
    <nav
      className="fixed bottom-[calc(0.5rem+env(safe-area-inset-bottom))] left-1/2 z-20 flex -translate-x-1/2 gap-1 rounded-3xl border border-border bg-background/90 p-1 shadow-lg shadow-foreground/10 backdrop-blur-md md:hidden"
      aria-label="移动导航"
    >
      {NAV_ITEMS.map((item) => (
        <MobileDockButton
          key={item.id}
          item={item}
          active={currentPage === item.id}
          onSelect={handleSelect}
        />
      ))}
    </nav>
  );
}

export function AppShell({ children }: { children: ReactNode }) {
  const { data: settings } = useSettingList();

  useEffect(() => {
    if (!settings) return;
    const themeAuto = settings.find((s) => s.key === "theme_auto")?.value;
    const theme = settings.find((s) => s.key === "theme")?.value;

    const root = document.documentElement;
    root.classList.remove("light", "dark");

    if (themeAuto !== "0") {
      const isDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
      root.classList.add(isDark ? "dark" : "light");
    } else {
      root.classList.add(theme === "dark" ? "dark" : "light");
    }
  }, [settings]);

  return (
    <div className="flex h-dvh overflow-hidden bg-background font-sans text-foreground">
      <DesktopSidebar />
      <main className="flex min-h-0 flex-1 flex-col pt-4 px-4">
        {children}
      </main>
      <MobileDock />
    </div>
  );
}
