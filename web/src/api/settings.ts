import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "./client";

export type Setting = {
    key: string;
    value: string;
};

export type IfaceInfo = {
    name: string;
    ip: string;
};

const queryKey = ["setting"];

export function useSettingList() {
    return useQuery({
        queryKey,
        queryFn: () => apiRequest<Setting[]>("/api/v1/setting"),
    });
}

export function useInterfaces() {
    return useQuery({
        queryKey: ["interfaces"],
        queryFn: () => apiRequest<IfaceInfo[]>("/api/v1/setting/interfaces"),
    });
}

export function useSetSetting() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (payload: Setting) =>
            apiRequest<null>("/api/v1/setting", { method: "POST", body: payload }),
        onSuccess: () => qc.invalidateQueries({ queryKey }),
    });
}
