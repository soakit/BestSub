import { useState, type FormEvent } from "react";
import { AlertDialog, Button, Form, Input, Label, ListBox, Modal, Select, TextField, toast, useOverlayState } from "@heroui/react";
import { Cloud, Flask, Pencil, Plus, TrashBin } from "@gravity-ui/icons";
import { type Storage, type StorageConfig, type StorageType, useCreateStorage, useDeleteStorage, useStorages, useTestStorage, useUpdateStorage } from "../../api/storage";

type StorageField = { key: string; label: string; placeholder: string; required?: boolean };

type StorageForm = {
    name: string;
    type: StorageType;
    params: Record<string, string>;
};

const storageTypeLabels: Record<StorageType, string> = {
    local: "本地",
    webdav: "WebDAV",
    gist: "GitHub Gist",
};

const storageDefaultParams: Record<StorageType, Record<string, string>> = {
    local: {},
    webdav: { endpoint: "", username: "", password: "" },
    gist: { token: "", gist_id: "" },
};

// 表单字段由类型配置驱动；提交时再收敛成后端需要的 StorageConfig。
const storageFields: Record<StorageType, StorageField[]> = {
    local: [],
    webdav: [
        { key: "endpoint", label: "WebDAV 地址", placeholder: "https://example.com/dav", required: true },
        { key: "username", label: "用户名", placeholder: "可选" },
        { key: "password", label: "密码", placeholder: "可选" },
    ],
    gist: [
        { key: "token", label: "GitHub Token", placeholder: "gist 写权限 token", required: true },
        { key: "gist_id", label: "Gist ID", placeholder: "目标 Gist ID", required: true },
    ],
};

export function Storage() {
    const emptyForm: StorageForm = { name: "", type: "local", params: storageDefaultParams.local };
    const { data: storages = [] } = useStorages();
    const createStorage = useCreateStorage();
    const updateStorage = useUpdateStorage();
    const deleteStorage = useDeleteStorage();
    const testStorage = useTestStorage();
    const editorState = useOverlayState({ defaultOpen: false });
    const [editing, setEditing] = useState<Storage | null>(null);
    const [deletingId, setDeletingId] = useState<string | null>(null);
    const [form, setForm] = useState<StorageForm>(emptyForm);

    const closeEditor = () => {
        editorState.close();
        setEditing(null);
        setForm(emptyForm);
    };

    const openCreate = () => {
        setEditing(null);
        setForm(emptyForm);
        editorState.open();
    };

    const openEdit = (storage: Storage) => {
        setEditing(storage);
        if (storage.type === "webdav") {
            setForm({ name: storage.name, type: "webdav", params: { endpoint: storage.params.endpoint, username: storage.params.username ?? "", password: storage.params.password ?? "" } });
        } else if (storage.type === "gist") {
            setForm({ name: storage.name, type: "gist", params: { token: storage.params.token, gist_id: storage.params.gist_id } });
        } else {
            setForm({ name: storage.name, type: "local", params: storageDefaultParams.local });
        }
        editorState.open();
    };

    const setParam = (key: string, value: string) => {
        setForm((prev) => ({ ...prev, params: { ...prev.params, [key]: value } }));
    };

    const handleTypeChange = (value: unknown) => {
        if (value === form.type) return;
        if (value === "local" || value === "webdav" || value === "gist") {
            setForm({ name: form.name, type: value, params: storageDefaultParams[value] });
        }
    };

    const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        const name = form.name.trim();
        if (!name) return;

        // 按 internal/storage 的真实类型协议组装 params，避免把无关字段写入配置。
        let payload: StorageConfig;
        if (form.type === "local") {
            payload = { name, type: "local", params: {} };
        } else if (form.type === "webdav") {
            if (!(form.params.endpoint ?? "").trim()) return;
            payload = { name, type: "webdav", params: { endpoint: form.params.endpoint.trim(), username: (form.params.username ?? "").trim(), password: form.params.password ?? "" } };
        } else {
            if (!(form.params.token ?? "").trim() || !(form.params.gist_id ?? "").trim()) return;
            payload = { name, type: "gist", params: { token: form.params.token.trim(), gist_id: form.params.gist_id.trim() } };
        }

        if (editing) {
            updateStorage.mutate({ ...payload, id: editing.id }, { onSuccess: closeEditor });
        } else {
            createStorage.mutate(payload, { onSuccess: closeEditor });
        }
    };

    return (
        <div className="settings-category">
            <div className="text-foreground/85 mt-1 mb-2.5 flex items-center gap-2 px-1 text-base font-semibold tracking-tight">
                <Cloud className="size-4 shrink-0" />
                <span className="flex-1">储存</span>
                <Button isIconOnly size="sm" variant="ghost" onPress={openCreate}>
                    <Plus className="size-4" />
                </Button>
            </div>
            <div className="bg-surface grid grid-cols-1 overflow-hidden rounded-xl">
                {storages.length === 0 ? (
                    <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                        <span className="min-w-0 grow truncate text-sm text-muted">暂无储存</span>
                        <div className="flex shrink-0 gap-1">
                            <Button isIconOnly size="sm" variant="ghost" isDisabled>
                                <Flask className="size-4" />
                            </Button>
                            <Button isIconOnly size="sm" variant="ghost" isDisabled>
                                <Pencil className="size-4" />
                            </Button>
                            <Button isIconOnly size="sm" variant="ghost" isDisabled>
                                <TrashBin className="size-4" />
                            </Button>
                        </div>
                    </div>
                ) : (
                    [...storages].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).map((storage) => (
                        <div key={storage.id} className="flex min-h-11 items-center gap-3 px-4 py-2">
                            <span className="text-foreground min-w-0 grow truncate text-sm" title={storage.name}>{storage.name}</span>
                            <div className="flex shrink-0 gap-1">
                                <Button
                                    isIconOnly
                                    size="sm"
                                    variant="ghost"
                                    className="text-muted hover:text-success"
                                    isDisabled={testStorage.isPending}
                                    onPress={() => testStorage.mutate(storage, {
                                        onSuccess: () => toast.success(storage.name, { description: "测试成功" }),
                                        onError: (err) => toast.danger(storage.name, { description: err instanceof Error ? err.message : "测试失败" }),
                                    })}
                                >
                                    <Flask className="size-4" />
                                </Button>
                                <Button isIconOnly size="sm" variant="ghost" className="text-muted hover:text-accent" onPress={() => openEdit(storage)}>
                                    <Pencil className="size-4" />
                                </Button>
                                <Button isIconOnly size="sm" variant="ghost" className="text-muted hover:text-danger" onPress={() => setDeletingId(storage.id)}>
                                    <TrashBin className="size-4" />
                                </Button>
                            </div>
                        </div>
                    ))
                )}
            </div>

            <Modal state={editorState}>
                <Modal.Backdrop>
                    <Modal.Container>
                        <Modal.Dialog>
                            <Modal.CloseTrigger />
                            <Modal.Header>
                                <Modal.Heading>{editing ? "编辑储存" : "添加储存"}</Modal.Heading>
                            </Modal.Header>
                            <Modal.Body>
                                <Form id="storage-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={handleSubmit}>
                                    <TextField isRequired value={form.name} onChange={(value) => setForm({ ...form, name: value })}>
                                        <Label>名称</Label>
                                        <Input placeholder="储存名称" variant="secondary" />
                                    </TextField>
                                    <Select className="w-full" variant="secondary" value={form.type} onChange={handleTypeChange}>
                                        <Label>类型</Label>
                                        <Select.Trigger><Select.Value /><Select.Indicator /></Select.Trigger>
                                        <Select.Popover>
                                            <ListBox>
                                                {(Object.keys(storageTypeLabels) as StorageType[]).map((type) => (
                                                    <ListBox.Item key={type} id={type}>{storageTypeLabels[type]}</ListBox.Item>
                                                ))}
                                            </ListBox>
                                        </Select.Popover>
                                    </Select>
                                    {storageFields[form.type].map((field) => (
                                        <TextField key={field.key} isRequired={field.required} value={form.params[field.key] ?? ""} onChange={(value) => setParam(field.key, value)}>
                                            <Label>{field.label}</Label>
                                            <Input placeholder={field.placeholder} variant="secondary" />
                                        </TextField>
                                    ))}
                                </Form>
                            </Modal.Body>
                            <Modal.Footer>
                                <Button variant="ghost" onPress={closeEditor}>取消</Button>
                                <Button type="submit" form="storage-form" variant="primary" isDisabled={createStorage.isPending || updateStorage.isPending}>{editing ? "保存" : "添加"}</Button>
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
                                isDisabled={deleteStorage.isPending}
                                onPress={() => {
                                    if (deletingId) deleteStorage.mutate(deletingId);
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
