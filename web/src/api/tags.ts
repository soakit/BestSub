import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "./client";
import type { Tag } from "./sub";

export type { Tag };

const queryKey = ["tags"];

export function useTags() {
    return useQuery({
        queryKey,
        queryFn: () => apiRequest<Tag[]>("/api/v1/tag/list"),
    });
}

export function useCreateTag() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (name: string) =>
            apiRequest<Tag>("/api/v1/tag/create", { method: "POST", body: { name } }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useDeleteTag() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: number) =>
            apiRequest<string>(`/api/v1/tag/del/${id}`, { method: "DELETE" }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}
