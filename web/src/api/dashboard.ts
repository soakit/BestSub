import { useQuery } from "@tanstack/react-query";
import { apiRequest } from "./client";
import type { Subscription } from "./sub";

export type DashboardData = {
    total_nodes: number;
    subscriptions_total: number;
    tasks_total: number;
    shares_total: number;
    subscriptions: Subscription[];
    country_counts: Record<string, number>;
};

export function useDashboardSummary() {
    return useQuery({
        queryKey: ["dashboard", "summary"],
        queryFn: () => apiRequest<DashboardData>("/api/v1/dashboard/summary"),
    });
}
