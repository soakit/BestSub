import { useEffect, useState, type FormEvent } from "react";
import { AlertDialog, Button, Form, Label, Modal, TextArea, TextField, useOverlayState } from "@heroui/react";
import { Pencil, Plus, TrashBin } from "@gravity-ui/icons";
import { useCreateRenameTemplate, useDeleteRenameTemplate, useRenamePreview, useRenameTemplates } from "../../api/rename";

export function Rename() {
    const { data: templates = [] } = useRenameTemplates();
    const createTemplate = useCreateRenameTemplate();
    const deleteTemplate = useDeleteRenameTemplate();
    const renamePreview = useRenamePreview();
    const editorState = useOverlayState({ defaultOpen: false });
    const [expression, setExpression] = useState("");
    const [deletingId, setDeletingId] = useState<number | null>(null);
    const normalizedExpression = expression.trim();
    const templateExists = templates.some((template) => template.expression === normalizedExpression);
    const previewReady = renamePreview.variables === normalizedExpression && renamePreview.data !== undefined;

    useEffect(() => {
        // 输入变化时清除旧结果并延迟预览，避免显示过期名称。
        renamePreview.reset();
        if (!normalizedExpression) return;
        const timeout = window.setTimeout(() => renamePreview.mutate(normalizedExpression), 400);
        return () => window.clearTimeout(timeout);
    }, [normalizedExpression]);

    const closeEditor = () => {
        editorState.close();
        setExpression("");
        renamePreview.reset();
    };

    const openCreate = () => {
        setExpression("");
        renamePreview.reset();
        editorState.open();
    };

    const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!normalizedExpression || templateExists || !previewReady) return;
        createTemplate.mutate(normalizedExpression, { onSuccess: closeEditor });
    };

    return (
        <div className="settings-category">
            <div className="text-foreground/85 mt-1 mb-2.5 flex items-center gap-2 px-1 text-base font-semibold tracking-tight">
                <Pencil className="size-4 shrink-0" />
                <span className="flex-1">重命名模板</span>
                <Button isIconOnly size="sm" variant="ghost" onPress={openCreate}>
                    <Plus className="size-4" />
                </Button>
            </div>
            <div className="bg-surface grid grid-cols-1 overflow-hidden rounded-xl">
                {templates.length === 0 ? (
                    <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                        <span className="min-w-0 grow truncate text-sm text-muted">暂无重命名模板</span>
                        <Button isIconOnly size="sm" variant="ghost" isDisabled>
                            <TrashBin className="size-4" />
                        </Button>
                    </div>
                ) : (
                    [...templates].sort((a, b) => b.id - a.id).map((template) => (
                        <div key={template.id} className="flex min-h-11 items-center gap-3 px-4 py-2">
                            <span className="text-foreground min-w-0 grow truncate text-sm">{template.preview}</span>
                            <Button
                                isIconOnly
                                size="sm"
                                variant="ghost"
                                className="text-muted hover:text-danger"
                                onPress={() => setDeletingId(template.id)}
                            >
                                <TrashBin className="size-4" />
                            </Button>
                        </div>
                    ))
                )}
            </div>

            <Modal state={editorState}>
                <Modal.Backdrop>
                    <Modal.Container size="lg">
                        <Modal.Dialog>
                            <Modal.CloseTrigger />
                            <Modal.Header>
                                <Modal.Heading>添加重命名模板</Modal.Heading>
                            </Modal.Header>
                            <Modal.Body>
                                <Form id="rename-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={handleSubmit}>
                                    <div className="flex flex-col gap-1.5">
                                        <span className="text-foreground text-sm font-medium">预览</span>
                                        <div
                                            aria-live="polite"
                                            className={`flex h-10 items-center overflow-x-auto whitespace-nowrap rounded-xl bg-surface-secondary px-3 text-sm ${renamePreview.isError ? "text-danger" : previewReady ? "text-foreground" : "text-muted"}`}
                                        >
                                            {!normalizedExpression
                                                ? "输入表达式后预览"
                                                : renamePreview.isPending
                                                    ? "预览中"
                                                    : renamePreview.isError
                                                        ? renamePreview.error.message
                                                        : previewReady
                                                            ? renamePreview.data.result
                                                            : "等待预览"}
                                        </div>
                                    </div>
                                    <TextField isRequired value={expression} onChange={setExpression}>
                                        <Label>表达式</Label>
                                        <TextArea
                                            rows={5}
                                            className="font-mono"
                                            spellCheck={false}
                                            placeholder="{{.Country.Alpha2}} {{.Index}}"
                                            variant="secondary"
                                        />
                                    </TextField>
                                </Form>
                            </Modal.Body>
                            <Modal.Footer>
                                <Button variant="ghost" onPress={closeEditor}>取消</Button>
                                <Button
                                    type="submit"
                                    form="rename-form"
                                    variant="primary"
                                    isDisabled={!normalizedExpression || templateExists || !previewReady || createTemplate.isPending}
                                >
                                    添加
                                </Button>
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
                                isDisabled={deleteTemplate.isPending}
                                onPress={() => {
                                    if (deletingId) deleteTemplate.mutate(deletingId);
                                    setDeletingId(null);
                                }}
                            >
                                删除
                            </Button>
                        </AlertDialog.Footer>
                    </AlertDialog.Dialog>
                </AlertDialog.Container>
            </AlertDialog.Backdrop>
        </div>
    );
}
