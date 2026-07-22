import { useState } from "react";
import { AlertDialog, Button, Modal, Spinner, useOverlayState } from "@heroui/react";
import { Plus } from "@gravity-ui/icons";
import { type Share, useDeleteShare, useShares } from "../../api/share";
import { PageLayout } from "../PageLayout";
import { ShareForm } from "./ShareForm";
import { ShareItem } from "./ShareItem";

export default function SharePage() {
    const { data: shares, isLoading } = useShares();
    const deleteShare = useDeleteShare();
    const editorState = useOverlayState();
    const [editing, setEditing] = useState<Share | null>(null); // 当前正在编辑的分享，空值表示新建。
    const [editorKey, setEditorKey] = useState(0); // 每次打开时重建表单，避免保留上次输入。
    const [deletingId, setDeletingId] = useState<string | null>(null); // 等待确认删除的分享 ID。

    if (isLoading) {
        return <PageLayout title="分享"><div className="flex min-h-[28rem] items-center justify-center"><Spinner size="sm" /></div></PageLayout>;
    }

    const openEditor = (share?: Share) => {
        setEditing(share ?? null);
        setEditorKey((key) => key + 1);
        editorState.open();
    };

    return (
        <PageLayout
            title="分享"
            actions={
                <Button isIconOnly variant="ghost" onPress={() => openEditor()} className="rounded-xl">
                    <Plus className="size-4 text-foreground/50" />
                </Button>
            }
        >
            <>
                {shares?.length === 0 ? (
                    <div className="flex flex-1 items-center justify-center text-sm text-foreground/60">暂无分享</div>
                ) : (
                    <div className="grid grid-cols-[repeat(auto-fill,minmax(14rem,1fr))] gap-4">
                        {[...shares!].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).map((share) => (
                            <ShareItem
                                key={share.id}
                                share={share}
                                onEdit={openEditor}
                                onDelete={(item) => setDeletingId(item.id)}
                            />
                        ))}
                    </div>
                )}

                <Modal state={editorState} data-scrollbar="none">
                    <Modal.Backdrop variant="blur">
                        <Modal.Container>
                            <Modal.Dialog>
                                <Modal.CloseTrigger />
                                <Modal.Header>
                                    <Modal.Heading>{editing ? "编辑分享" : "添加分享"}</Modal.Heading>
                                </Modal.Header>
                                <Modal.Body>
                                    <ShareForm
                                        key={`${editing?.id ?? "create"}-${editorKey}`}
                                        share={editing ?? undefined}
                                        onClose={editorState.close}
                                    />
                                </Modal.Body>
                                <Modal.Footer>
                                    <Button variant="ghost" slot="close">取消</Button>
                                    <Button type="submit" form="share-form" variant="primary">{editing ? "保存" : "添加"}</Button>
                                </Modal.Footer>
                            </Modal.Dialog>
                        </Modal.Container>
                    </Modal.Backdrop>
                </Modal>

                <AlertDialog.Backdrop isOpen={!!deletingId} onOpenChange={(open) => { if (!open) setDeletingId(null); }} variant="blur">
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
                                        if (deletingId) deleteShare.mutate(deletingId);
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
