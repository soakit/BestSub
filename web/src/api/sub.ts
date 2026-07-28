import { useEffect, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest, apiUnauthorizedEvent } from "./client";

export type Tag = {
    id: number;
    name: string;
};

export type SubscriptionStatus = {
    node_num: number;
    traffic_total: number;
    traffic_used: number;
    expires_at: string;
    refreshed_at: string;
};

export type SubscriptionConfig = {
    url_type: number;
    url: string[];
    enable: number;
    name: string;
    tag_names: string[];
    header: Record<string, string>;
    auto_update: number;
    cron_expr: string;
    proxy_mode: number;
    protocol_filter_mode: number;
    protocol_filter: string[];
};

export type Subscription = {
    id: string;
    created_at: string;
} & SubscriptionConfig & SubscriptionStatus;

export const queryKey = ["subscription"];

export function useSubscription() {
    return useQuery({
        queryKey,
        queryFn: () => apiRequest<Subscription[]>("/api/v1/sub/list"),
    });
}

export function useGetSubscription(id: string) {
    return useQuery({
        queryKey: ["subscription", id],
        queryFn: () => apiRequest<Subscription>(`/api/v1/sub/get/${id}`),
        enabled: !!id,
    });
}

export function useCreateSubscription() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (payload: SubscriptionConfig) =>
            apiRequest<Subscription>("/api/v1/sub/create", { method: "POST", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useUpdateSubscription() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, ...payload }: SubscriptionConfig & { id: string }) =>
            apiRequest<string>(`/api/v1/sub/update/${id}`, { method: "PUT", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useDeleteSubscription() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) =>
            apiRequest<string>(`/api/v1/sub/del/${id}`, { method: "DELETE" }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useRefreshSubscription() {
    return useMutation({
        mutationFn: (id: string) =>
            apiRequest<void>(`/api/v1/sub/refresh/${id}`, { method: "POST" }),
    });
}

export type RefreshEvent = {
    sub_id: string;
    type: "progress" | "done" | "failed";
    payload?: Record<string, unknown>;
};

type Listener = (ev: RefreshEvent) => void;

let es: EventSource | null = null;
const listeners = new Set<Listener>();

function connect() {
    if (es) return;
    es = new EventSource("/api/v1/sub/refresh", {
        withCredentials: true,
    });

    const onEvent = (e: MessageEvent) => {
        try {
            const ev = JSON.parse(e.data) as RefreshEvent;
            listeners.forEach((fn) => fn(ev));
        } catch { /* heartbeat 或非法数据 */ }
    };

    es.addEventListener("progress", onEvent);
    es.addEventListener("done", onEvent);
    es.addEventListener("failed", onEvent);

    es.onerror = () => {
        disconnect();
        setTimeout(() => { if (listeners.size > 0) connect(); }, 3000);
    };
}

function disconnect() {
    es?.close();
    es = null;
}

if (typeof document !== "undefined") {
    document.addEventListener("visibilitychange", () => {
        if (document.hidden) disconnect();
        else if (listeners.size > 0) connect();
    });
    window.addEventListener(apiUnauthorizedEvent, () => {
        disconnect();
        listeners.clear();
    });
}

export function useRefreshStream(onEvent: (ev: RefreshEvent) => void) {
    const cbRef = useRef(onEvent);
    cbRef.current = onEvent;

    useEffect(() => {
        const listener: Listener = (ev) => cbRef.current(ev);
        listeners.add(listener);
        if (!document.hidden) connect();
        return () => {
            listeners.delete(listener);
            if (listeners.size === 0) disconnect();
        };
    }, []);
}
