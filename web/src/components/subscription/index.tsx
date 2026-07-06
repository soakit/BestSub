import { useRef, useState } from "react";
import { AlertDialog, Button, Spinner } from "@heroui/react";
import { PageLayout } from "../PageLayout";
import { Plus, ArrowRotateRight } from "@gravity-ui/icons";
import { useSubscription, useDeleteSubscription, useRefreshSubscription, type Subscription } from "../../api/sub";
import { SubForm } from "./SubForm";
import { SubItem } from "./SubItem";

export default function Subscription() {
    const { data: subscriptions, isLoading } = useSubscription();
    const refreshSub = useRefreshSubscription();
    const deleteSub = useDeleteSubscription();
    const modalRef = useRef<(sub?: Subscription) => void>(null);
    const [deletingId, setDeletingId] = useState<string | null>(null);

    const refreshAll = async () => {
        if (!subscriptions?.length) return;
        for (const sub of subscriptions) {
            try { await refreshSub.mutateAsync(sub.id); } catch { }
        }
    };

    return (
        <PageLayout
            title="订阅"
            actions={
                <>
                    <Button isIconOnly variant="ghost" onPress={refreshAll} className="rounded-xl">
                        <ArrowRotateRight className="size-4 text-foreground/50" />
                    </Button>
                    <Button isIconOnly variant="ghost" onPress={() => modalRef.current?.()} className="rounded-xl">
                        <Plus className="size-4 text-foreground/50" />
                    </Button>
                </>
            }
        >
            <>
                {isLoading ? (
                    <div className="flex flex-1 items-center justify-center">
                        <Spinner size="sm" />
                    </div>
                ) : subscriptions?.length === 0 ? (
                    <div className="flex flex-1 items-center justify-center text-sm text-foreground/60">暂无订阅</div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {[...subscriptions!].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).map((sub) => (
                            <SubItem
                                key={sub.id}
                                sub={sub}
                                onEdit={(s) => modalRef.current?.(s)}
                                onDelete={(s) => setDeletingId(s.id)}
                                refreshSub={refreshSub}
                            />
                        ))}
                    </div>
                )}

                <SubForm ref={modalRef} />

                <AlertDialog.Backdrop isOpen={!!deletingId} onOpenChange={(open) => { if (!open) setDeletingId(null) }} variant="blur">
                    <AlertDialog.Container size="xs">
                        <AlertDialog.Dialog >
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
                                        if (deletingId) deleteSub.mutate(deletingId);
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
