import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "@heroui/react";
import { Pencil, TrashBin, Clock, CircleCheck, Play, Stop } from "@gravity-ui/icons";
import { queryKey, useRunTask, useStopTask, useTaskProgressStream, useTaskResultCount, type Task, type TaskProgress } from "../../api/task";

export function TaskItem({ task, onEdit, onDelete }: { task: Task; onEdit: (task: Task) => void; onDelete: (task: Task) => void }) {
    const qc = useQueryClient();
    const runTask = useRunTask();
    const stopTask = useStopTask();
    const resultCount = useTaskResultCount(task.id);
    const [progress, setProgress] = useState<TaskProgress | null>(null);

    useTaskProgressStream((ev) => {
        if (ev.taskId !== task.id) return;
        if (ev.type === "progress") {
            setProgress(ev);
            return;
        }
        setProgress(null);
        qc.invalidateQueries({ queryKey: ["task", task.id, "result"] });
        qc.invalidateQueries({ queryKey });
    });

    return (
        <div className="bg-surface rounded-2xl p-4 flex flex-col h-full">
            <div className="flex justify-between items-start gap-4 mb-3">
                <h3 className="flex-1 min-w-0 text-foreground text-xl leading-snug line-clamp-1" title={task.name}>{task.name || "未命名任务"}</h3>
                <div className="flex gap-0.5 bg-surface-secondary rounded-lg p-0.5 shrink-0">
                    {progress || runTask.isPending ? (
                        <button
                            onClick={() => stopTask.mutate(task.id, { onSettled: () => setProgress(null) })}
                            disabled={runTask.isPending || stopTask.isPending}
                            className="p-1.5 text-muted hover:text-danger hover:bg-surface rounded-md transition-all disabled:opacity-50"
                        >
                            <Stop className="size-3.5" />
                        </button>
                    ) : (
                        <button
                            onClick={() => {
                                setProgress({ taskId: task.id, step: 0, done: 0, total: 0 });
                                runTask.mutate(task.id, {
                                    onError: (err) => {
                                        setProgress(null);
                                        toast.danger(task.name || "未命名任务", { description: err instanceof Error ? err.message : "启动失败" });
                                    },
                                });
                            }}
                            className="p-1.5 text-muted hover:text-success hover:bg-surface rounded-md transition-all"
                        >
                            <Play className="size-3.5" />
                        </button>
                    )}
                    <button onClick={() => onEdit(task)} className="p-1.5 text-muted hover:text-accent hover:bg-surface rounded-md transition-all">
                        <Pencil className="size-3.5" />
                    </button>
                    <button onClick={() => onDelete(task)} className="p-1.5 text-muted hover:text-danger hover:bg-surface rounded-md transition-all">
                        <TrashBin className="size-3.5" />
                    </button>
                </div>
            </div>

            <div className="flex flex-1 flex-col gap-2">
                <div className="flex flex-col gap-4 rounded-xl bg-surface-secondary p-3">
                    <div className="flex items-center gap-1 text-accent"><CircleCheck className="size-4" /><span className="text-xs text-muted">结果</span></div>
                    <div className="flex flex-1 items-end justify-end text-xl leading-none text-foreground">{resultCount.data ?? "-"}</div>
                </div>
                <div className="flex flex-col gap-4 rounded-xl bg-surface-secondary p-3">
                    <div className="flex items-center gap-1 text-accent"><Clock className="size-4" /><span className="text-xs text-muted">运行</span></div>
                    <div className="flex flex-1 items-end justify-end text-right text-xl leading-none text-foreground">
                        {progress && progress.step > 0 ? (
                            <div className="flex w-full items-end justify-between gap-3">
                                <span className="min-w-0 truncate text-left">{task.steps[progress.step - 1].type === "speed" ? "测速" : task.steps[progress.step - 1].type === "country" ? "落地" : "延迟"} {progress.step}/{task.steps.length}</span>
                                <span className="shrink-0">{progress.done}/{progress.total}</span>
                            </div>
                        ) : progress
                            ? "启动中"
                            : task.finished_at && task.finished_at !== "0001-01-01T00:00:00Z" && task.finished_at !== null
                                ? new Date(task.finished_at).toLocaleString(undefined, { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" })
                                : "从未运行"}
                    </div>
                </div>
            </div>
        </div>
    );
}
