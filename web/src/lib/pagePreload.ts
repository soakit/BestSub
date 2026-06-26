import type { Page } from "../store";

export const pageImports = {
  subscriptions: () => import("../components/Subscriptions"),
  nodes: () => import("../components/Nodes"),
  testing: () => import("../components/Testing"),
  sharing: () => import("../components/Sharing"),
  storage: () => import("../components/Storage"),
  settings: () => import("../components/Settings"),
};

const preloadedPages = new Set<Page>();

export function preloadPage(page: Page) {
  if (!preloadedPages.has(page)) {
    preloadedPages.add(page);
    void pageImports[page]();
  }
}

export function preloadRemainingPages(currentPage: Page) {
  const pages = (Object.keys(pageImports) as Page[]).filter((page) => page !== currentPage);

  // 空闲时分批加载剩余页面，避免登录后一次性解析多个 chunk 影响当前页面交互。
  const preloadNext = () => {
    const page = pages.shift();
    if (page) {
      preloadPage(page);
      window.setTimeout(preloadNext, 200);
    }
  };

  if (window.requestIdleCallback) {
    window.requestIdleCallback(preloadNext);
  } else {
    window.setTimeout(preloadNext, 500);
  }
}
