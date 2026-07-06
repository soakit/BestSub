import { lazy, Suspense, useDeferredValue, useEffect, useRef, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { checkAuth } from "./api/auth";
import { apiUnauthorizedEvent } from "./api/client";
import { hideBootLoading } from "./lib/bootLoading";
import { pageImports, preloadRemainingPages } from "./lib/pagePreload";
import { useAppStore } from "./store";
import { AppShell } from "./components/AppShell";
import Login from "./components/Login";

const Subscription = lazy(pageImports.subscription);
const Node = lazy(pageImports.node);
const Testing = lazy(pageImports.testing);
const Sharing = lazy(pageImports.sharing);
const Storage = lazy(pageImports.storage);
const Setting = lazy(pageImports.setting);

// 放在 Suspense 内部，只有当前 lazy 页面提交后才会隐藏 HTML 启动画面。
function BootLoadingGate({ children }: { children: ReactNode }) {
  useEffect(() => {
    hideBootLoading();
  }, []);

  return children;
}

export default function App() {
  const currentPage = useAppStore(state => state.currentPage);
  const visiblePage = useDeferredValue(currentPage);
  const preloadedRemainingPages = useRef(false);
  const { data: authenticated, isPending, refetch } = useQuery({
    queryKey: ["auth", "check"],
    queryFn: checkAuth,
    retry: false,
  });

  useEffect(() => {
    const handleUnauthorized = () => {
      void refetch();
    };
    window.addEventListener(apiUnauthorizedEvent, handleUnauthorized);
    return () => window.removeEventListener(apiUnauthorizedEvent, handleUnauthorized);
  }, [refetch]);

  useEffect(() => {
    if (authenticated && !preloadedRemainingPages.current) {
      preloadedRemainingPages.current = true;
      preloadRemainingPages(currentPage);
    }
  }, [authenticated, currentPage]);

  if (isPending) {
    return null;
  }

  if (!authenticated) {
    return (
      <BootLoadingGate>
        <Login onSuccess={() => { void refetch(); }} />
      </BootLoadingGate>
    );
  }

  return (
    <AppShell>
      <Suspense fallback={null}>
        <BootLoadingGate>
          {visiblePage === 'subscription' && <Subscription />}
          {visiblePage === 'node' && <Node />}
          {visiblePage === 'testing' && <Testing />}
          {visiblePage === 'sharing' && <Sharing />}
          {visiblePage === 'storage' && <Storage />}
          {visiblePage === 'setting' && <Setting />}
        </BootLoadingGate>
      </Suspense>
    </AppShell>
  );
}
