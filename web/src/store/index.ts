import { type ElementType } from 'react';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { Signal, Server, Pulse, Link as LinkIcon, Cloud, Gear } from '@gravity-ui/icons';

export type Page = 'subscriptions' | 'nodes' | 'testing' | 'sharing' | 'storage' | 'settings';

export type NavItem = { id: Page; label: string; icon: ElementType };

export const MAIN_NAV_ITEMS: NavItem[] = [
  { id: 'subscriptions', label: '订阅', icon: Signal },
  { id: 'nodes', label: '节点', icon: Server },
  { id: 'testing', label: '检测', icon: Pulse },
  { id: 'sharing', label: '分享', icon: LinkIcon },
  { id: 'storage', label: '储存', icon: Cloud },
];

export const SETTINGS_NAV_ITEM: NavItem = { id: 'settings', label: '设置', icon: Gear };

export const NAV_ITEMS: NavItem[] = [...MAIN_NAV_ITEMS, SETTINGS_NAV_ITEM];

interface AppState {
  currentPage: Page;
  setCurrentPage: (page: Page) => void;
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      currentPage: 'subscriptions',
      setCurrentPage: (page) => set({ currentPage: page }),
    }),
    { name: 'bestsub-app' }
  )
);
