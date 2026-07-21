import { useImperativeHandle, useState } from "react";
import { Button, Modal, Form, TextField, Label, Skeleton, Switch, TextArea, useOverlayState } from "@heroui/react";
import { useConvertNode, useCreateNode, useUpdateNode, type Node, type NodeConfig } from "../../api/node";
import { useTags } from "../../api/tags";
import { TagSelector } from "../common/TagSelector";

export function NodeForm({ ref }: { ref?: React.Ref<(node?: Node) => void> }) {
    const state = useOverlayState();
    const [editing, setEditing] = useState<Node | null>(null);
    const [content, setContent] = useState("");
    const [converted, setConverted] = useState(""); // 保存后端返回并最终入库的 Mihomo JSON。
    const [tagNames, setTagNames] = useState<string[]>([]);
    const [landingOnly, setLandingOnly] = useState(0); // 是否仅作落地节点。
    const { data: allTags = [] } = useTags();
    const convertNode = useConvertNode();
    const createNode = useCreateNode();
    const updateNode = useUpdateNode();
    const preview = converted ? JSON.parse(converted) as Record<string, unknown> : null;

    useImperativeHandle(ref, () => (node?: Node) => {
        setEditing(node ?? null);
        setContent(node?.content ?? "");
        setConverted(node?.content ?? "");
        setTagNames(node?.tag_names ?? []);
        setLandingOnly(node?.landing_only ?? 0);
        convertNode.reset();
        state.open();
    });

    const handleSubmit = (e: React.SyntheticEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!converted) return;

        const payload: NodeConfig = {
            name: String(preview?.name),
            content: converted,
            tag_names: tagNames,
            landing_only: landingOnly,
        };
        if (editing) {
            updateNode.mutate({ ...payload, id: editing.id }, { onSuccess: () => state.setOpen(false) });
        } else {
            createNode.mutate(payload, { onSuccess: () => state.setOpen(false) });
        }
    };

    return (
        <Modal state={state}>
            <Modal.Backdrop variant="blur">
                <Modal.Container>
                    <Modal.Dialog>
                        <Modal.CloseTrigger />
                        <Modal.Header>
                            <Modal.Heading>{editing ? "编辑节点" : "添加节点"}</Modal.Heading>
                        </Modal.Header>
                        <Modal.Body>
                            <Form id="node-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={handleSubmit}>
                                <TextField isRequired>
                                    <Label>节点内容</Label>
                                    <TextArea
                                        value={content}
                                        disabled={convertNode.isPending}
                                        variant="secondary"
                                        onChange={(event) => {
                                            setContent(event.target.value);
                                            setConverted("");
                                            convertNode.reset();
                                        }}
                                        onBlur={() => {
                                            if (!content.trim() || converted || convertNode.isPending) return;
                                            convertNode.mutate(content.trim(), { onSuccess: setConverted });
                                        }}
                                    />
                                </TextField>

                                <div className="min-h-16 w-full rounded-2xl bg-surface-secondary px-4 py-3">
                                    {preview ? (
                                        <>
                                            <div className="truncate text-sm text-foreground">{String(preview.name)}</div>
                                            <div className="mt-1 truncate text-xs text-muted">
                                                {String(preview.type).toUpperCase()} · {String(preview.server)}:{String(preview.port)}
                                            </div>
                                        </>
                                    ) : (
                                        <div className="space-y-2">
                                            <Skeleton className="h-4 w-3/5 rounded-lg" />
                                            <Skeleton className="h-3 w-2/5 rounded-lg" />
                                        </div>
                                    )}
                                </div>

                                <div className="flex items-center">
                                    <span className="text-sm text-foreground">仅作落地节点</span>
                                    <div className="flex-1" />
                                    <Switch aria-label="仅作落地节点" isSelected={landingOnly === 1} onChange={(selected) => setLandingOnly(selected ? 1 : 0)}>
                                        <Switch.Content><Switch.Control><Switch.Thumb /></Switch.Control></Switch.Content>
                                    </Switch>
                                </div>

                                <TagSelector value={allTags.filter((tag) => tagNames.includes(tag.name))} onChange={(tags) => setTagNames(tags.map((tag) => tag.name))} />
                            </Form>
                        </Modal.Body>
                        <Modal.Footer>
                            <Button variant="ghost" slot="close">取消</Button>
                            <Button
                                type="submit"
                                form="node-form"
                                variant="primary"
                                isPending={createNode.isPending || updateNode.isPending}
                                isDisabled={!converted || convertNode.isPending || createNode.isPending || updateNode.isPending}
                            >
                                {editing ? "保存" : "添加"}
                            </Button>
                        </Modal.Footer>
                    </Modal.Dialog>
                </Modal.Container>
            </Modal.Backdrop>
        </Modal>
    );
}
