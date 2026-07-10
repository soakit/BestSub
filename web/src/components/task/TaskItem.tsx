import { Pencil, TrashBin, Clock, Server, Cloud, ListCheck } from "@gravity-ui/icons";
import type { Task } from "../../api/task";

export function TaskItem({ task, onEdit, onDelete }: { task: Task; onEdit: (task: Task) => void; onDelete: (task: Task) => void }) {
    return (
        <div className="bg-surface rounded-2xl p-4 flex flex-col h-full">
            <div className="flex justify-between items-start gap-4 mb-3">
                <h3 className="min-w-0 flex-1 line-clamp-1 text-xl leading-snug text-foreground" title={task.name}>{task.name || "未命名任务"}</h3>
                <div className="flex gap-0.5 bg-surface-secondary rounded-lg p-0.5 shrink-0">
                    <button onClick={() => onEdit(task)} className="p-1.5 text-muted hover:text-accent hover:bg-surface rounded-md transition-all">
                        <Pencil className="size-3.5" />
                    </button>
                    <button onClick={() => onDelete(task)} className="p-1.5 text-muted hover:text-danger hover:bg-surface rounded-md transition-all">
                        <TrashBin className="size-3.5" />
                    </button>
                </div>
            </div>

            <div className="grid flex-1 grid-cols-2 gap-2">
                <div className="flex flex-col gap-2 rounded-xl bg-surface-secondary p-3">
                    <div className="flex items-center gap-1 text-accent"><Server className="size-4" /><span className="text-xs text-muted">输入</span></div>
                    <div className="flex flex-1 items-end justify-end text-xl leading-none text-foreground">{task.subscriptions.length + task.nodes.length + task.tags.length + task.result_tasks.length}</div>
                </div>
                <div className="flex flex-col gap-2 rounded-xl bg-surface-secondary p-3">
                    <div className="flex items-center gap-1 text-accent"><ListCheck className="size-4" /><span className="text-xs text-muted">步骤</span></div>
                    <div className="flex flex-1 items-end justify-end text-xl leading-none text-foreground">{task.steps.length}</div>
                </div>
                <div className="flex flex-col gap-2 rounded-xl bg-surface-secondary p-3">
                    <div className="flex items-center gap-1 text-accent"><Cloud className="size-4" /><span className="text-xs text-muted">储存</span></div>
                    <div className="flex flex-1 items-end justify-end text-xl leading-none text-foreground">{task.storage_enable === 1 ? "开" : "-"}</div>
                </div>
                <div className="flex flex-col gap-2 rounded-xl bg-surface-secondary p-3">
                    <div className="flex items-center gap-1 text-accent"><Clock className="size-4" /><span className="text-xs text-muted">运行</span></div>
                    <div className="flex flex-1 items-end justify-end text-sm leading-none text-foreground">
                        {task.finished_at && task.finished_at !== "0001-01-01T00:00:00Z" && task.finished_at !== null
                            ? new Date(task.finished_at).toLocaleString()
                            : "从未运行"}
                    </div>
                </div>
            </div>
        </div>
    );
}
