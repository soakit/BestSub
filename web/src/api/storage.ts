import { useQuery } from "@tanstack/react-query";
import { apiRequest } from "./client";

export type Storage = {
    id: string;
};

const queryKey = ["storage"];

export function useStorages() {
    return useQuery({
        queryKey,
        queryFn: () => apiRequest<Storage[]>("/api/v1/storage/list"),
    });
}
