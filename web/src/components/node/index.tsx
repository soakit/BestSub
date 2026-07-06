import { useRef, useState } from "react";
import { AlertDialog, Button, Spinner } from "@heroui/react";
import { PageLayout } from "../PageLayout";
import { Plus } from "@gravity-ui/icons";
import { useNodes, useDeleteNode, type Node } from "../../api/node";
import { NodeForm } from "./NodeForm";
import { NodeItem } from "./NodeItem";

export default function Node() {
    const { data: nodes, isLoading } = useNodes();
    const deleteNode = useDeleteNode();
    const modalRef = useRef<(node?: Node) => void>(null);
    const [deletingId, setDeletingId] = useState<string | null>(null);

    return (
        <PageLayout
            title="节点"
            actions={
                <Button isIconOnly variant="ghost" onPress={() => modalRef.current?.()} className="rounded-xl">
                    <Plus className="size-4 text-foreground/50" />
                </Button>
            }
        >
            <>
                {isLoading ? (
                    <div className="flex flex-1 items-center justify-center">
                        <Spinner size="sm" />
                    </div>
                ) : nodes?.length === 0 ? (
                    <div className="flex flex-1 items-center justify-center text-sm text-foreground/60">暂无节点</div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {[...nodes!].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).map((node) => (
                            <NodeItem
                                key={node.id}
                                node={node}
                                onEdit={(n) => modalRef.current?.(n)}
                                onDelete={(n) => setDeletingId(n.id)}
                            />
                        ))}
                    </div>
                )}

                <NodeForm ref={modalRef} />

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
                                        if (deletingId) deleteNode.mutate(deletingId);
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
