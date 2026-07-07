import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { ProgressCircle, toast } from "@heroui/react";
import { Pencil, TrashBin, ArrowsRotateRight, Clock, Server, ChartLine } from "@gravity-ui/icons";
import { useDeleteSubscription, useRefreshSubscription, useRefreshStream, useGetSubscription, queryKey, type RefreshEvent, type Subscription } from "../../api/sub";
import { formatBytes, formatRelativeTime } from "../../lib/format";

function formatExpireDate(s: string): string {
    if (!s || s.startsWith("0001")) return "";
    const d = new Date(s);
    if (isNaN(d.getTime())) return "";
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

export function SubItem({
    sub,
    onEdit,
    onDelete,
    refreshSub,
}: {
    sub: Subscription;
    onEdit: (sub: Subscription) => void;
    onDelete: (sub: Subscription) => void;
    refreshSub: ReturnType<typeof useRefreshSubscription>;
}) {
    const qc = useQueryClient();
    const [progress, setProgress] = useState<{ current: number; total: number; status: string } | null>(null);
    const { refetch: refetchSub } = useGetSubscription(progress ? sub.id : "");

    // 监听 SSE，只处理当前订阅的刷新事件
    useRefreshStream((ev: RefreshEvent) => {
        if (ev.sub_id !== sub.id) return;
        switch (ev.type) {
            case "progress":
                setProgress({
                    current: (ev.payload?.index as number) + 1,
                    total: ev.payload?.total as number,
                    status: ev.payload?.status as string,
                });
                if (ev.payload?.status === "fail" && ev.payload?.error) {
                    toast.danger(sub.name, { description: ev.payload.error as string });
                }
                break;
            case "done":
                refetchSub().then(({ data: updated }) => {
                    if (updated) {
                        qc.setQueryData<Subscription[]>(queryKey, (old) =>
                            old?.map((s) => (s.id === updated.id ? updated : s)) ?? [updated]
                        );
                    }
                    setProgress(null);
                });
                break;
            case "failed":
                setProgress(null);
                toast.danger(sub.name, { description: ev.payload?.message as string });
                break;
        }
    });

    const hasTraffic = sub.traffic_total > 0;
    const trafficPercent = hasTraffic ? Math.min((sub.traffic_used / sub.traffic_total) * 100, 100) : 0;

    return (
        <>
            <div className="bg-surface rounded-2xl p-4 flex flex-col h-full">
                <div className="flex justify-between items-start gap-4 mb-4">
                    <h3 className="flex-1 min-w-0 text-foreground text-xl leading-snug line-clamp-1" title={sub.name}>{sub.name}</h3>
                    <div className="flex gap-0.5 bg-surface-secondary rounded-lg p-0.5 shrink-0">
                        <button onClick={() => onEdit(sub)} className="p-1.5 text-muted hover:text-accent hover:bg-surface rounded-md transition-all">
                            <Pencil className="size-3.5" />
                        </button>
                        {progress ? (
                            <div className="p-1.5 overflow-hidden flex items-center justify-center">
                                <ProgressCircle
                                    aria-label="刷新进度"
                                    className="size-3.5"
                                    color={progress.status === "fail" ? "danger" : "accent"}
                                    value={progress.current}
                                    minValue={0}
                                    maxValue={progress.total}
                                >
                                    <ProgressCircle.Track>
                                        <ProgressCircle.TrackCircle />
                                        <ProgressCircle.FillCircle />
                                    </ProgressCircle.Track>
                                </ProgressCircle>
                            </div>
                        ) : (
                            <button
                                onClick={() => refreshSub.mutate(sub.id)}
                                disabled={refreshSub.isPending}
                                className="p-1.5 text-muted hover:text-success hover:bg-surface rounded-md transition-all"
                            >
                                <ArrowsRotateRight className={`size-3.5 ${refreshSub.isPending ? "animate-spin" : ""}`} />
                            </button>
                        )}
                        <button onClick={() => onDelete(sub)} className="p-1.5 text-muted hover:text-danger hover:bg-surface rounded-md transition-all">
                            <TrashBin className="size-3.5" />
                        </button>
                    </div>
                </div>

                <div className="flex flex-col gap-2 flex-1">
                    {hasTraffic && (
                        <div className="bg-surface-secondary rounded-xl p-3 flex flex-col gap-2">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-1 text-accent"><ChartLine className="size-4" /><span className="text-xs font-medium text-muted">流量消耗</span></div>
                                <span className="text-xl text-foreground leading-none">{trafficPercent.toFixed(1)}%</span>
                            </div>
                            <div className="h-1.5 w-full bg-surface-tertiary rounded-full overflow-hidden">
                                <div className={`h-full rounded-full transition-all duration-700 ${trafficPercent > 90 ? "bg-danger" : trafficPercent > 75 ? "bg-warning" : "bg-accent"}`} style={{ width: `${trafficPercent}%` }} />
                            </div>
                            <div className="flex justify-between text-xs font-medium text-muted">
                                <span>{formatBytes(sub.traffic_used)} / {formatBytes(sub.traffic_total)}</span>
                                <span>{formatExpireDate(sub.expires_at) || "-"}</span>
                            </div>
                        </div>
                    )}

                    <div className={`flex gap-2 flex-1 w-full ${hasTraffic ? "" : "flex-col"}`}>
                        <div className={`bg-surface-secondary rounded-xl p-3 flex flex-col gap-2 ${hasTraffic ? "w-1/2" : "flex-1"}`}>
                            <div className="flex items-center gap-1 text-accent"><Server className="size-4" /><span className="text-xs font-medium text-muted">可用节点</span></div>
                            <div className="flex-1 flex items-end justify-end text-xl text-foreground leading-none">{sub.node_num}</div>
                        </div>
                        <div className={`bg-surface-secondary rounded-xl p-3 flex flex-col gap-2 ${hasTraffic ? "w-1/2" : "flex-1"}`}>
                            <div className="flex items-center gap-1 text-accent"><Clock className="size-4" /><span className="text-xs font-medium text-muted">更新时间</span></div>
                            <div className="flex-1 flex items-end justify-end text-xl text-foreground leading-none">{formatRelativeTime(sub.refreshed_at)}</div>
                        </div>
                    </div>
                </div>
            </div>
        </>
    );
}
