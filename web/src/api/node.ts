import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "./client";

export type NodeConfig = {
    name: string;
    tag_names: string[];
    content: string;
};

export type NodeInfo = {
    delay: number;
    download_speed: number; // 下载速度，单位 kb/s。
    country_code: string;
    traffic_multiplier: number;
};

export type Node = {
    id: string;
    created_at: string;
} & NodeConfig & NodeInfo;

const queryKey = ["nodes"];

export function useNodes() {
    return useQuery({
        queryKey,
        queryFn: () => apiRequest<Node[]>("/api/v1/node/list"),
    });
}

export function useCreateNode() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (payload: NodeConfig) =>
            apiRequest<Node>("/api/v1/node/create", { method: "POST", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useUpdateNode() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, ...payload }: NodeConfig & { id: string }) =>
            apiRequest<string>(`/api/v1/node/update/${id}`, { method: "PUT", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useDeleteNode() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) =>
            apiRequest<string>(`/api/v1/node/del/${id}`, { method: "DELETE" }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}
