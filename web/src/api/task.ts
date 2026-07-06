import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "./client";
import type { Node } from "./node";
import type { Storage } from "./storage";
import type { Subscription, Tag } from "./sub";

export type TaskParams = {
    url: string;
    timeout_ms: number;
    attempts: number;
    max_bytes: number;
    max_duration_ms: number;
};

export type TaskPass = {
    limit: number;
    min_delay: number;
    max_delay: number;
    min_download_speed: number;
    max_download_speed: number;
    include_country_codes: string[];
    exclude_country_codes: string[];
};

export type TaskOrder = "none" | "delay" | "speed";

export type TaskStep = {
    type: "delay" | "download" | "country";
    params?: Partial<TaskParams>;
    concurrency?: number;
    pass: Partial<TaskPass>;
    order: TaskOrder;
};

export type StorageConfig = {
    storage_enable: number;
    storage_id: string;
    storage?: Storage;
    save_path: string;
    node_rename_expression: string;
};

export type TaskConfig = {
    name: string;
    auto_run: number;
    cron_expr: string;
    steps: TaskStep[];
    subscriptions: Subscription[];
    nodes: Node[];
    tags: Tag[];
    result_tasks: TaskRef[];
    custom_landing_node_enable: number;
    landing_subscriptions: Subscription[];
    landing_nodes: Node[];
} & StorageConfig;

export type TaskRef = {
    id: string;
    name: string;
};

export type Task = {
    id: string;
} & TaskConfig;

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
        queryFn: () => apiRequest<Task>(`/api/v1/task/${id}`),
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
            apiRequest<string>(`/api/v1/task/${id}`, { method: "PUT", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useDeleteTask() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) =>
            apiRequest<string>(`/api/v1/task/${id}`, { method: "DELETE" }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}
