import { useEffect, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest, apiUnauthorizedEvent } from "./client";
import type { Storage } from "./storage";

export type TaskParams = {
    url: string;
    timeout_ms: number;
    attempts: number;
    max_kb: number; // 最大读取量，单位 kb。
    max_duration_ms: number;
};

export type TaskPass = {
    limit: number;
    min_delay: number;
    max_delay: number;
    min_download_speed: number; // 最小下载速度，单位 kb/s。
    max_download_speed: number; // 最大下载速度，单位 kb/s。
    include_country_codes: string[];
    exclude_country_codes: string[];
};

export type TaskOrder = "none" | "delay" | "speed";

export const taskSaveFormats = [ // 任务结果支持的保存格式，与后端转换目标保持一致。
    "QuantumultX",
    "Surge",
    "Loon",
    "SurgeMac",
    "Mihomo",
    "URI",
    "V2Ray",
    "ShadowRocket",
    "Surfboard",
    "singbox",
    "Egern",
] as const;

export type TaskSaveFormat = (typeof taskSaveFormats)[number];

export type TaskStep = {
    type: "delay" | "speed" | "country";
    params?: Partial<TaskParams>;
    concurrency?: number;
    node_pool_delete: number;
    skip_existing: number; // 对应检测结果已存在时是否跳过本步骤探测。
    pass: Partial<TaskPass>;
    order: TaskOrder;
};

export type StorageConfig = {
    storage_enable: number;
    storage_id: string;
    storage?: Storage;
    save_format: TaskSaveFormat;
    save_path: string;
    node_rename_expression: string;
};

export type TaskConfig = {
    name: string;
    auto_run: number;
    cron_expr: string;
    steps: TaskStep[];
    subscriptions: TaskSubscription[];
    nodes: TaskNode[];
    tags: TaskTag[];
    result_tasks: TaskInputResult[];
    all_input_enable: number;
    custom_landing_node_enable: number;
    landing_subscriptions: TaskSubscription[];
    landing_nodes: TaskNode[];
} & StorageConfig;

export type TaskSubscription = {
    id: string;
};

export type TaskNode = {
    id: string;
};

export type TaskTag = {
    id: number;
};

export type TaskInputResult = {
    id: string;
};

export type Task = {
    id: string;
    create_at: string;
    finished_at: string;
} & TaskConfig;

export type TaskProgress = {
    taskId: string;
    step: number;
    done: number;
    total: number;
};

export type TaskProgressEvent = TaskProgress & {
    type: "progress" | "done";
};

export const queryKey = ["task"];

export function useTasks() {
    return useQuery({
        queryKey,
        queryFn: () => apiRequest<Task[]>("/api/v1/task/list"),
    });
}

export function useGetTask(id: string) {
    return useQuery({
        queryKey: ["task", id],
        queryFn: () => apiRequest<Task>(`/api/v1/task/get/${id}`),
        enabled: !!id,
    });
}

export function useTaskResultCount(id: string) {
    return useQuery({
        queryKey: ["task", id, "result"],
        queryFn: () => apiRequest<number>(`/api/v1/task/result/${id}`),
        enabled: !!id,
    });
}

export function useCreateTask() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (payload: TaskConfig) =>
            apiRequest<Task>("/api/v1/task/create", { method: "POST", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useUpdateTask() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, ...payload }: TaskConfig & { id: string }) =>
            apiRequest<string>(`/api/v1/task/update/${id}`, { method: "PUT", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useDeleteTask() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) =>
            apiRequest<string>(`/api/v1/task/del/${id}`, { method: "DELETE" }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useRunTask() {
    return useMutation({
        mutationFn: (id: string) =>
            apiRequest<void>(`/api/v1/task/run/${id}`, { method: "POST" }),
    });
}

export function useStopTask() {
    return useMutation({
        mutationFn: (id: string) =>
            apiRequest<void>(`/api/v1/task/stop/${id}`, { method: "POST" }),
    });
}

type Listener = (ev: TaskProgressEvent) => void;

let es: EventSource | null = null;
let listeners: Set<Listener> = new Set();
let connecting = false;

function connect() {
    if (es || connecting) return;
    connecting = true;
    es = new EventSource("/api/v1/task/stream", {
        withCredentials: true,
    });

    const onProgress = (e: MessageEvent) => {
        try {
            listeners.forEach((fn) => fn({ ...(JSON.parse(e.data) as TaskProgress), type: "progress" }));
        } catch { }
    };
    const onDone = (e: MessageEvent) => {
        try {
            listeners.forEach((fn) => fn({ ...(JSON.parse(e.data) as TaskProgress), type: "done" }));
        } catch { }
    };

    es.addEventListener("progress", onProgress);
    es.addEventListener("done", onDone);

    es.onerror = () => {
        es?.close();
        es = null;
        connecting = false;
        setTimeout(() => { if (listeners.size > 0) connect(); }, 3000);
    };

    connecting = false;
}

function disconnect() {
    es?.close();
    es = null;
    connecting = false;
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

export function useTaskProgressStream(onEvent: (ev: TaskProgressEvent) => void) {
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
