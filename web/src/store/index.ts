import { type ElementType } from 'react';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { Signal, Server, Pulse, Link as LinkIcon, Gear } from '@gravity-ui/icons';

export type Page = 'subscription' | 'node' | 'task' | 'sharing' | 'setting';

export type NavItem = { id: Page; label: string; icon: ElementType };

export const MAIN_NAV_ITEMS: NavItem[] = [
  { id: 'subscription', label: '订阅', icon: Signal },
  { id: 'node', label: '节点', icon: Server },
  { id: 'task', label: '任务', icon: Pulse },
  { id: 'sharing', label: '分享', icon: LinkIcon },
];

export const SETTINGS_NAV_ITEM: NavItem = { id: 'setting', label: '设置', icon: Gear };

export const NAV_ITEMS: NavItem[] = [...MAIN_NAV_ITEMS, SETTINGS_NAV_ITEM];

interface AppState {
  currentPage: Page;
  setCurrentPage: (page: Page) => void;
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      currentPage: 'subscription',
      setCurrentPage: (page) => set({ currentPage: page }),
    }),
    { name: 'bestsub-app' }
  )
);
