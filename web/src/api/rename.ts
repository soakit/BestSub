import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "./client";

export type RenameTemplate = {
    id: number;
    preview: string;
    expression: string;
};

const queryKey = ["rename"];

export function useRenameTemplates() {
    return useQuery({
        queryKey,
        queryFn: () => apiRequest<RenameTemplate[]>("/api/v1/rename/list"),
    });
}

export function useCreateRenameTemplate() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (expression: string) =>
            apiRequest<RenameTemplate>("/api/v1/rename/create", { method: "POST", body: { expression } }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useDeleteRenameTemplate() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: number) =>
            apiRequest<string>(`/api/v1/rename/del/${id}`, { method: "DELETE" }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}

export function useRenamePreview() {
    return useMutation({
        mutationFn: (expression: string) =>
            apiRequest<{ result: string }>("/api/v1/rename/preview", { method: "POST", body: { expression } }),
    });
}
