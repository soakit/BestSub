import { useImperativeHandle, useState } from "react";
import { Button, Modal, Form, TextField, Label, Input, TextArea, useOverlayState } from "@heroui/react";
import { useCreateNode, useUpdateNode, type Node, type NodeConfig } from "../../api/node";
import { useTags } from "../../api/tags";
import { TagSelector } from "../common/TagSelector";

export function NodeForm({ ref }: { ref?: React.Ref<(node?: Node) => void> }) {
    const state = useOverlayState();
    const [editing, setEditing] = useState<Node | null>(null);
    const [tagNames, setTagNames] = useState<string[]>([]);
    const { data: allTags = [] } = useTags();
    const createNode = useCreateNode();
    const updateNode = useUpdateNode();

    useImperativeHandle(ref, () => (node?: Node) => {
        setEditing(node ?? null);
        setTagNames(node?.tag_names ?? []);
        state.open();
    });

    const handleSubmit = (e: React.SyntheticEvent<HTMLFormElement>) => {
        e.preventDefault();
        const fd = new FormData(e.currentTarget);
        const content = String(fd.get("content")).trim();
        if (!content) return;

        const payload: NodeConfig = {
            name: String(fd.get("name")).trim(),
            content,
            tag_names: tagNames,
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
                                <TextField defaultValue={editing?.name ?? ""}>
                                    <Label>节点名称</Label>
                                    <Input name="name" placeholder="我的节点" variant="secondary" />
                                </TextField>

                                <TextField isRequired defaultValue={editing?.content ?? ""}>
                                    <Label>节点内容</Label>
                                    <TextArea name="content" placeholder="vmess://... 或节点配置内容" variant="secondary" />
                                </TextField>

                                <TagSelector value={allTags.filter((tag) => tagNames.includes(tag.name))} onChange={(tags) => setTagNames(tags.map((tag) => tag.name))} />
                            </Form>
                        </Modal.Body>
                        <Modal.Footer>
                            <Button variant="ghost" slot="close">取消</Button>
                            <Button type="submit" form="node-form" variant="primary">{editing ? "保存" : "添加"}</Button>
                        </Modal.Footer>
                    </Modal.Dialog>
                </Modal.Container>
            </Modal.Backdrop>
        </Modal>
    );
}
