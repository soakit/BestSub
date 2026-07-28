import { useRef, useState } from "react";
import { AlertDialog, Button, Spinner } from "@heroui/react";
import { PageLayout } from "../PageLayout";
import { Plus } from "@gravity-ui/icons";
import { type Task, useDeleteTask, useTasks } from "../../api/task";
import { TaskForm } from "./TaskForm";
import { TaskItem } from "./TaskItem";

export default function TaskPage() {
    const { data: tasks, isLoading } = useTasks();
    const deleteTask = useDeleteTask();
    const modalRef = useRef<(task?: Task) => void>(null); // 打开任务编辑弹窗的方法。
    const [deletingId, setDeletingId] = useState<string | null>(null);

    if (isLoading) {
        return <PageLayout title="任务"><div className="flex min-h-[28rem] items-center justify-center"><Spinner size="sm" /></div></PageLayout>;
    }

    return (
        <PageLayout
            title="任务"
            actions={
                <Button isIconOnly variant="ghost" onPress={() => modalRef.current?.()} className="rounded-xl">
                    <Plus className="size-4 text-foreground/50" />
                </Button>
            }
        >
            <>
                {tasks?.length === 0 ? (
                    <div className="flex flex-1 items-center justify-center text-sm text-foreground/60">暂无任务</div>
                ) : (
                    <div className="grid grid-cols-[repeat(auto-fill,minmax(14rem,1fr))] gap-4">
                        {[...tasks!].sort((a, b) => new Date(b.create_at).getTime() - new Date(a.create_at).getTime()).map((task) => (
                            <TaskItem
                                key={task.id}
                                task={task}
                                onEdit={(item) => modalRef.current?.(item)}
                                onDelete={(t) => setDeletingId(t.id)}
                            />
                        ))}
                    </div>
                )}

                <TaskForm ref={modalRef} tasks={tasks ?? []} />

                <AlertDialog.Backdrop isOpen={!!deletingId} onOpenChange={(open) => { if (!open) setDeletingId(null) }} variant="blur">
                    <AlertDialog.Container size="xs">
                        <AlertDialog.Dialog>
                            <AlertDialog.CloseTrigger />
                            <AlertDialog.Header>
                                <AlertDialog.Icon status="danger" />
                                <AlertDialog.Heading>确定要删除吗？</AlertDialog.Heading>
                            </AlertDialog.Header>
                            <AlertDialog.Footer>
                                <Button slot="close" variant="tertiary">取消</Button>
                                <Button
                                    slot="close"
                                    variant="danger"
                                    onPress={() => {
                                        if (deletingId) deleteTask.mutate(deletingId);
                                        setDeletingId(null);
                                    }}
                                >
                                    删除
                                </Button>
                            </AlertDialog.Footer>
                        </AlertDialog.Dialog>
                    </AlertDialog.Container>
                </AlertDialog.Backdrop>
            </>
        </PageLayout>
    );
}
