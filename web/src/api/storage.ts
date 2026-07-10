import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "./client";

export type StorageType = "local" | "webdav" | "gist";

export type StorageConfig =
    | { name: string; type: "local"; params: Record<string, never> }
    | { name: string; type: "webdav"; params: { endpoint: string; username?: string; password?: string } }
    | { name: string; type: "gist"; params: { token: string; gist_id: string } };

export type Storage = StorageConfig & {
    id: string;
    status: string;
};

const queryKey = ["storage"];

export function useStorages() {
    return useQuery({
        queryKey,
        queryFn: () => apiRequest<Storage[]>("/api/v1/storage/list"),
    });
}

export function useCreateStorage() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (payload: StorageConfig) =>
            apiRequest<Storage>("/api/v1/storage/create", { method: "POST", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useUpdateStorage() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, ...payload }: StorageConfig & { id: string }) =>
            apiRequest<string>(`/api/v1/storage/update/${id}`, { method: "PUT", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useDeleteStorage() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) =>
            apiRequest<string>(`/api/v1/storage/del/${id}`, { method: "DELETE" }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useTestStorage() {
    return useMutation({
        mutationFn: (payload: StorageConfig) =>
            apiRequest<string>("/api/v1/storage/test", { method: "POST", body: payload }),
    });
}
