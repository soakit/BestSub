import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "./client";

export type ShareFilter = {
    min_delay?: number;
    max_delay?: number;
    min_download_speed?: number; // 最小下载速度，单位 kb/s。
    max_download_speed?: number; // 最大下载速度，单位 kb/s。
    include_country_codes?: string[];
    exclude_country_codes?: string[];
};

export type ShareInput = {
    subscriptions: { id: string }[];
    nodes: { id: string }[];
    tags: { id: number }[];
    result_tasks: { id: string }[];
};

export type ShareConfig = {
    name: string;
    filter: ShareFilter;
    node_rename_expression: string;
} & ShareInput;

export type Share = {
    id: string;
    token: string;
    created_at: string;
    node_count: number;
} & ShareConfig;

const queryKey = ["share"];

export function useShares() {
    return useQuery({
        queryKey,
        queryFn: () => apiRequest<Share[]>("/api/v1/share/list"),
    });
}

export function useGetShare(id: string) {
    return useQuery({
        queryKey: ["share", id],
        queryFn: () => apiRequest<Share>(`/api/v1/share/get/${id}`),
        enabled: !!id,
    });
}

export function useCreateShare() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (payload: ShareConfig) =>
            apiRequest<Share>("/api/v1/share/create", { method: "POST", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useUpdateShare() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, ...payload }: ShareConfig & { id: string }) =>
            apiRequest<string>(`/api/v1/share/update/${id}`, { method: "PUT", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useDeleteShare() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) =>
            apiRequest<string>(`/api/v1/share/del/${id}`, { method: "DELETE" }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}
