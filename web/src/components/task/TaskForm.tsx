import { useEffect, useImperativeHandle, useMemo, useState } from "react";
import type { Key } from "@heroui/react";
import { Autocomplete, Button, Disclosure, Drawer, Dropdown, EmptyState, Form, Input, Label, ListBox, Modal, SearchField, Select, Switch, Tag, TagGroup, TextArea, TextField, useFilter, useOverlayState } from "@heroui/react";
import { FileArrowDown, Pencil, Plus, TrashBin } from "@gravity-ui/icons";
import { useNodes } from "../../api/node";
import { useRenamePreview, useRenameTemplates } from "../../api/rename";
import { useStorages } from "../../api/storage";
import { useSubscription } from "../../api/sub";
import { useTags } from "../../api/tags";
import { taskSaveFormats, type Task, type TaskConfig, type TaskSaveFormat, type TaskStep, useCreateTask, useUpdateTask } from "../../api/task";

type TaskSourceType = "subscription" | "node" | "tag" | "result_task";

const taskSourceTypes: { id: TaskSourceType; name: string }[] = [
    { id: "subscription", name: "订阅" },
    { id: "tag", name: "标签" },
    { id: "node", name: "节点" },
    { id: "result_task", name: "任务" },
];

const defaultConfig: TaskConfig = {
    name: "",
    auto_run: 0,
    cron_expr: "",
    steps: [],
    subscriptions: [],
    nodes: [],
    tags: [],
    result_tasks: [],
    all_input_enable: 0,
    custom_landing_node_enable: 0,
    landing_node: { id: "" },
    storage_enable: 0,
    storage_id: null,
    save_format: "Mihomo",
    save_path: "",
    node_rename_expression: "",
};

function newStep(): TaskStep {
    return {
        type: "delay",
        params: defaultStepParams("delay"),
        concurrency: 8,
        node_pool_delete: 0,
        skip_existing: 0,
        pass: {},
        order: 0,
    };
}

function defaultStepParams(type: TaskStep["type"]) {
    if (type === "speed") return { url: "https://speed.cloudflare.com/__down?during=download&bytes=999999" };
    if (type === "country") return {};
    return { url: "https://gstatic.com/generate_204", timeout_ms: 10000, attempts: 1 };
}

function keysToStrings(keys: Key | Key[] | null) {
    return (Array.isArray(keys) ? keys : keys ? [keys] : []).map(String);
}

function stepName(step: TaskStep) {
    if (step.type === "speed") return "测速";
    if (step.type === "country") return "落地";
    return "延迟";
}

function initialSourceRows(task?: Task): TaskSourceType[] {
    if (!task) return ["subscription"];
    const rows = [
        ...(task.subscriptions.length > 0 ? ["subscription" as const] : []),
        ...(task.nodes.length > 0 ? ["node" as const] : []),
        ...(task.tags.length > 0 ? ["tag" as const] : []),
        ...(task.result_tasks.length > 0 ? ["result_task" as const] : []),
    ];
    return rows.length > 0 ? rows : ["subscription"];
}

// 管理任务表单状态并在不重渲染任务列表的情况下打开编辑弹窗。
export function TaskForm({ ref, tasks }: { ref?: React.Ref<(task?: Task) => void>; tasks: Task[] }) {
    const state = useOverlayState();
    const [editing, setEditing] = useState<Task | null>(null); // 当前正在编辑的任务，空值表示新建。
    const { contains } = useFilter({ sensitivity: "base" });
    const { data: subscriptions = [] } = useSubscription();
    const { data: nodes = [] } = useNodes();
    const { data: tags = [] } = useTags();
    const { data: storages = [] } = useStorages();
    const { data: renameTemplates = [] } = useRenameTemplates();
    const renamePreview = useRenamePreview();
    const createTask = useCreateTask();
    const updateTask = useUpdateTask();
    const [editingStep, setEditingStep] = useState<number | null>(null);
    const [formState, setFormState] = useState<TaskConfig>({ ...defaultConfig, steps: [newStep()] });
    const [sourceRows, setSourceRows] = useState<TaskSourceType[]>(["subscription"]);
    const resultTasks = useMemo(() => tasks.filter((task) => task.id !== editing?.id), [tasks, editing?.id]);
    const allInputEnabled = formState.all_input_enable === 1; // 是否在本次编辑中动态使用全部订阅和单独节点。

    useImperativeHandle(ref, () => (task?: Task) => {
        setEditing(task ?? null);
        setFormState(task ? {
            ...defaultConfig,
            ...task,
            steps: task.steps.map((step) => ({
                ...step,
                params: step.type === "country"
                    ? { timeout_ms: step.params?.timeout_ms }
                    : { ...defaultStepParams(step.type), ...(step.params ?? {}) },
            })),
        } : { ...defaultConfig, steps: [newStep()] });
        setSourceRows(initialSourceRows(task));
        setEditingStep(null);
        state.open();
    });

    useEffect(() => {
        // 输入变化时清掉旧结果并延迟请求，避免把过期预览显示为当前表达式。
        renamePreview.reset();
        if (formState.storage_enable !== 1 || !formState.node_rename_expression.trim()) return;
        const timeout = window.setTimeout(() => renamePreview.mutate(formState.node_rename_expression.trim()), 400);
        return () => window.clearTimeout(timeout);
    }, [formState.storage_enable, formState.node_rename_expression]);

    const setForm = <K extends keyof TaskConfig>(key: K, value: TaskConfig[K]) => {
        setFormState((prev) => ({ ...prev, [key]: value }));
    };

    const setStep = (index: number, step: TaskStep) => {
        setForm("steps", formState.steps.map((s, i) => (i === index ? step : s)));
    };

    const getSourceValue = (type: TaskSourceType) => {
        if (type === "subscription") return formState.subscriptions.map((sub) => sub.id);
        if (type === "node") return formState.nodes.map((node) => node.id);
        if (type === "tag") return formState.tags.map((tag) => String(tag.id));
        return formState.result_tasks.map((t) => t.id);
    };

    const getSourceOptions = (type: TaskSourceType) => {
        if (type === "subscription") return subscriptions.map((sub) => ({ id: sub.id, name: sub.name }));
        if (type === "node") return nodes.map((node) => ({ id: node.id, name: node.name || node.id }));
        if (type === "tag") return tags.map((tag) => ({ id: String(tag.id), name: tag.name }));
        return resultTasks.map((t) => ({ id: t.id, name: t.name }));
    };

    const setSourceValue = (type: TaskSourceType, ids: string[]) => {
        if (type === "subscription") {
            setForm("subscriptions", ids.map((id) => ({ id })));
            return;
        }
        if (type === "node") {
            setForm("nodes", ids.map((id) => ({ id })));
            return;
        }
        if (type === "tag") {
            setForm("tags", ids.map((id) => ({ id: Number(id) })));
            return;
        }
        setForm("result_tasks", ids.map((id) => ({ id })));
    };

    const clearSourceValue = (type: TaskSourceType) => {
        if (type === "subscription") {
            setForm("subscriptions", []);
            return;
        }
        if (type === "node") {
            setForm("nodes", []);
            return;
        }
        if (type === "tag") {
            setForm("tags", []);
            return;
        }
        setForm("result_tasks", []);
    };

    const handleSubmit = (e: React.SyntheticEvent<HTMLFormElement>) => {
        e.preventDefault();
        const customLandingEnabled =
            formState.custom_landing_node_enable === 1 &&
            formState.landing_node.id !== "";

        const payload: TaskConfig = {
            ...formState,
            name: formState.name.trim(),
            cron_expr: formState.auto_run === 1 ? formState.cron_expr.trim() : "",
            subscriptions: formState.subscriptions.map((sub) => ({ id: sub.id })),
            nodes: formState.nodes.map((node) => ({ id: node.id })),
            tags: formState.tags.map((tag) => ({ id: tag.id })),
            result_tasks: formState.result_tasks.map((task) => ({ id: task.id })),
            custom_landing_node_enable: customLandingEnabled ? 1 : 0,
            landing_node: customLandingEnabled ? { id: formState.landing_node.id } : { id: "" },
            storage_id: formState.storage_enable === 1 ? formState.storage_id : null,
            save_path: formState.storage_enable === 1 ? formState.save_path.trim() : "",
            node_rename_expression: formState.storage_enable === 1 ? formState.node_rename_expression.trim() : "",
        };
        if (payload.steps.length === 0) return;
        if (editing) {
            updateTask.mutate({ ...payload, id: editing.id }, { onSuccess: state.close });
        } else {
            createTask.mutate(payload, { onSuccess: state.close });
        }
    };

    return (
        <Modal state={state}>
            <Modal.Backdrop variant="blur">
                <Modal.Container>
                    <Modal.Dialog>
                        <Modal.CloseTrigger />
                        <Modal.Header>
                            <Modal.Heading>{editing ? "编辑任务" : "添加任务"}</Modal.Heading>
                        </Modal.Header>
                        <Modal.Body>
                            <Form id="task-form" className="flex w-full flex-col gap-4" onSubmit={handleSubmit}>
                                <TextField isRequired name="name" value={formState.name} onChange={(value) => setForm("name", value)}>
                                    <Label>任务名称</Label>
                                    <Input placeholder="我的任务" variant="secondary" />
                                </TextField>

                                <Disclosure className="w-full" isExpanded={formState.auto_run === 1}>
                                    <Disclosure.Heading className="flex items-center">
                                        <span className="text-sm font-medium text-foreground">自动运行</span>
                                        <div className="flex-1" />
                                        <Switch isSelected={formState.auto_run === 1} onChange={() => setForm("auto_run", formState.auto_run === 1 ? 0 : 1)}>
                                            <Switch.Content><Switch.Control><Switch.Thumb /></Switch.Control></Switch.Content>
                                        </Switch>
                                    </Disclosure.Heading>
                                    <Disclosure.Content className="!overflow-visible">
                                        <TextField value={formState.cron_expr} onChange={(value) => setForm("cron_expr", value)} className="pt-2">
                                            <Label>Cron 表达式</Label>
                                            <Input name="cron_expr" placeholder="0 */6 * * *" variant="secondary" />
                                        </TextField>
                                    </Disclosure.Content>
                                </Disclosure>

                                <div className="flex flex-col gap-3">
                                    <div className="flex items-center">
                                        <span className="text-sm font-medium text-foreground">前置节点来源</span>
                                        <div className="flex-1" />
                                        <Switch isSelected={allInputEnabled} onChange={(selected) => setForm("all_input_enable", selected ? 1 : 0)}>
                                            <Switch.Content>
                                                全部输入
                                                <Switch.Control><Switch.Thumb /></Switch.Control>
                                            </Switch.Content>
                                        </Switch>
                                    </div>
                                    {sourceRows.filter((type) => !allInputEnabled || type === "result_task").map((type, index) => (
                                        <div key={index} className="flex items-center gap-2">
                                            <div className="grid min-w-0 flex-1 gap-2 sm:grid-cols-[6rem_minmax(0,1fr)] sm:items-center">
                                                <Select className="w-full" aria-label="来源类型" variant="secondary" value={type} onChange={(key) => {
                                                    const next = String(key) as TaskSourceType;
                                                    if (next === type) return;
                                                    clearSourceValue(type);
                                                    setSourceRows((rows) => rows.map((row, i) => (i === index ? next : row)));
                                                }}>
                                                    <Select.Trigger><Select.Value /><Select.Indicator /></Select.Trigger>
                                                    <Select.Popover>
                                                        <ListBox>
                                                            {taskSourceTypes.filter((source) =>
                                                                (!allInputEnabled || source.id === "result_task") &&
                                                                (source.id === type || !sourceRows.includes(source.id))
                                                            ).map((source) => (
                                                                <ListBox.Item key={source.id} id={source.id}>{source.name}</ListBox.Item>
                                                            ))}
                                                        </ListBox>
                                                    </Select.Popover>
                                                </Select>
                                                <TaskSourcePicker ariaLabel="选择来源" placeholder={`选择${taskSourceTypes.find((source) => source.id === type)!.name}`} value={getSourceValue(type)} options={getSourceOptions(type)} onChange={(ids) => setSourceValue(type, ids)} />
                                            </div>
                                            <Button isIconOnly size="sm" variant="ghost" className="text-muted hover:text-danger" onPress={() => {
                                                clearSourceValue(type);
                                                setSourceRows((rows) => rows.filter((_, i) => i !== index));
                                            }}>
                                                <TrashBin className="size-4" />
                                            </Button>
                                        </div>
                                    ))}
                                    <Button variant="ghost" className="w-full" isDisabled={allInputEnabled ? sourceRows.includes("result_task") : sourceRows.length >= taskSourceTypes.length} onPress={() => setSourceRows((rows) => [
                                        ...rows,
                                        allInputEnabled ? "result_task" : taskSourceTypes.find((source) => !rows.includes(source.id))!.id,
                                    ])}>
                                        <Plus className="size-4" />
                                        添加来源
                                    </Button>

                                    <Disclosure className="w-full" isExpanded={formState.custom_landing_node_enable === 1}>
                                        <Disclosure.Heading className="flex items-center">
                                            <span className="text-sm font-medium text-foreground">自定义落地节点</span>
                                            <div className="flex-1" />
                                            <Switch isSelected={formState.custom_landing_node_enable === 1} onChange={() => setForm("custom_landing_node_enable", formState.custom_landing_node_enable === 1 ? 0 : 1)}>
                                                <Switch.Content><Switch.Control><Switch.Thumb /></Switch.Control></Switch.Content>
                                            </Switch>
                                        </Disclosure.Heading>
                                        <Disclosure.Content className="pt-2 !overflow-visible">
                                            <Autocomplete isRequired={formState.custom_landing_node_enable === 1} className="w-full" placeholder="选择节点" selectionMode="single" value={formState.landing_node.id || null} variant="secondary" onChange={(key) => setForm("landing_node", { id: typeof key === "string" ? key : "" })}>
                                                <Label>落地节点</Label>
                                                <Autocomplete.Trigger>
                                                    <Autocomplete.Value />
                                                    <Autocomplete.ClearButton />
                                                    <Autocomplete.Indicator />
                                                </Autocomplete.Trigger>
                                                <Autocomplete.Popover>
                                                    <Autocomplete.Filter filter={contains}>
                                                        <SearchField autoFocus name="landing-node-search" variant="secondary">
                                                            <SearchField.Group>
                                                                <SearchField.SearchIcon />
                                                                <SearchField.Input placeholder="搜索节点" />
                                                                <SearchField.ClearButton />
                                                            </SearchField.Group>
                                                        </SearchField>
                                                        <ListBox renderEmptyState={() => <EmptyState>没有可用节点</EmptyState>}>
                                                            {nodes.map((node) => (
                                                                <ListBox.Item key={node.id} id={node.id} textValue={node.name || node.id}>{node.name || node.id}<ListBox.ItemIndicator /></ListBox.Item>
                                                            ))}
                                                        </ListBox>
                                                    </Autocomplete.Filter>
                                                </Autocomplete.Popover>
                                            </Autocomplete>
                                        </Disclosure.Content>
                                    </Disclosure>
                                </div>

                                <div className="flex flex-col gap-3">
                                    <span className="text-sm font-medium text-foreground">检测步骤</span>
                                    {formState.steps.map((step, index) => (
                                        <div key={index} className="flex items-center gap-2">
                                            <div className="min-w-0 flex-1 rounded-xl bg-surface-secondary p-3">
                                                <div className="text-sm text-foreground">{index + 1}. {stepName(step)}</div>
                                            </div>
                                            <Button isIconOnly size="sm" variant="ghost" className="text-muted hover:text-accent" onPress={() => setEditingStep(index)}>
                                                <Pencil className="size-4" />
                                            </Button>
                                            <Button isIconOnly size="sm" variant="ghost" className="text-muted hover:text-danger" onPress={() => setForm("steps", formState.steps.filter((_, i) => i !== index))}>
                                                <TrashBin className="size-4" />
                                            </Button>
                                        </div>
                                    ))}
                                    <Button variant="ghost" className="w-full" onPress={() => {
                                        setForm("steps", [...formState.steps, newStep()]);
                                        setEditingStep(formState.steps.length);
                                    }}>
                                        <Plus className="size-4" />
                                        添加步骤
                                    </Button>
                                </div>

                                <Disclosure className="w-full" isExpanded={formState.storage_enable === 1}>
                                    <Disclosure.Heading className="flex items-center">
                                        <span className="text-sm font-medium text-foreground">完成后储存</span>
                                        <div className="flex-1" />
                                        <Switch isSelected={formState.storage_enable === 1} onChange={() => setForm("storage_enable", formState.storage_enable === 1 ? 0 : 1)}>
                                            <Switch.Content><Switch.Control><Switch.Thumb /></Switch.Control></Switch.Content>
                                        </Switch>
                                    </Disclosure.Heading>
                                    <Disclosure.Content className="flex flex-col gap-4 pt-2 !overflow-visible">
                                        <Select isRequired={formState.storage_enable === 1} className="w-full" variant="secondary" value={formState.storage_id} onChange={(key) => setForm("storage_id", typeof key === "string" ? key : null)}>
                                                    <Label>储存目标</Label>
                                                    <Select.Trigger><Select.Value /><Select.Indicator /></Select.Trigger>
                                                    <Select.Popover>
                                                        <ListBox>
                                                            {storages.map((storage) => (
                                                                <ListBox.Item key={storage.id} id={storage.id} textValue={storage.name}>{storage.name}</ListBox.Item>
                                                            ))}
                                                        </ListBox>
                                                    </Select.Popover>
                                                </Select>
                                                <Select className="w-full" variant="secondary" value={formState.save_format} onChange={(key) => setForm("save_format", String(key) as TaskSaveFormat)}>
                                                    <Label>保存格式</Label>
                                                    <Select.Trigger><Select.Value /><Select.Indicator /></Select.Trigger>
                                                    <Select.Popover>
                                                        <ListBox>
                                                            {taskSaveFormats.map((format) => (
                                                                <ListBox.Item key={format} id={format} textValue={format}>{format}</ListBox.Item>
                                                            ))}
                                                        </ListBox>
                                                    </Select.Popover>
                                                </Select>
                                                <TextField value={formState.node_rename_expression} onChange={(value) => setForm("node_rename_expression", value)}>
                                                    <div className="flex items-center gap-2">
                                                        <Label className="flex-1">重命名表达式</Label>
                                                        <Dropdown>
                                                            <Button isIconOnly type="button" size="sm" variant="ghost" isDisabled={renameTemplates.length === 0}>
                                                                <FileArrowDown className="size-4" />
                                                            </Button>
                                                            <Dropdown.Popover className="min-w-64">
                                                                <Dropdown.Menu onAction={(key) => {
                                                                    const template = renameTemplates.find((template) => template.id === Number(key));
                                                                    if (template) setForm("node_rename_expression", template.expression);
                                                                }}>
                                                                    {[...renameTemplates].sort((a, b) => b.id - a.id).map((template) => (
                                                                        <Dropdown.Item key={template.id} id={String(template.id)} textValue={template.preview}>
                                                                            <Label>{template.preview}</Label>
                                                                        </Dropdown.Item>
                                                                    ))}
                                                                </Dropdown.Menu>
                                                            </Dropdown.Popover>
                                                        </Dropdown>
                                                    </div>
                                                    <TextArea rows={2} name="node_rename_expression" placeholder="{{.Country.NameZh}}-{{.Index}}" variant="secondary" />                        <span className={`text-xs ${renamePreview.isError ? "text-danger" : "text-muted"}`}>
                                                        预览：{renamePreview.isError ? renamePreview.error.message : renamePreview.data?.result}
                                                    </span>
                                                </TextField>
                                        <TextField value={formState.save_path} onChange={(value) => setForm("save_path", value)}>
                                            <Label>保存路径</Label>
                                            <Input name="save_path" placeholder="/nodes.yaml" variant="secondary" />
                                        </TextField>
                                    </Disclosure.Content>
                                </Disclosure>

                                <Drawer.Backdrop isOpen={editingStep !== null} onOpenChange={(open) => { if (!open) setEditingStep(null); }} variant="blur">
                                    <Drawer.Content placement="right">
                                        <Drawer.Dialog>
                                            <Drawer.CloseTrigger />
                                            <Drawer.Header><Drawer.Heading>编辑检测步骤</Drawer.Heading></Drawer.Header>
                                            <Drawer.Body>
                                                {editingStep !== null && formState.steps[editingStep] && (
                                                    <TaskStepFields step={formState.steps[editingStep]} onChange={(step) => setStep(editingStep, step)} />
                                                )}
                                            </Drawer.Body>
                                            <Drawer.Footer>
                                                <Button variant="ghost" onPress={() => setEditingStep(null)}>完成</Button>
                                            </Drawer.Footer>
                                        </Drawer.Dialog>
                                    </Drawer.Content>
                                </Drawer.Backdrop>
                            </Form>
                        </Modal.Body>
                        <Modal.Footer>
                            <Button variant="ghost" slot="close">取消</Button>
                            <Button type="submit" form="task-form" variant="primary">{editing ? "保存" : "添加"}</Button>
                        </Modal.Footer>
                    </Modal.Dialog>
                </Modal.Container>
            </Modal.Backdrop>
        </Modal>
    );
}

function TaskSourcePicker({ label, ariaLabel, placeholder, value, options, onChange }: { label?: string; ariaLabel?: string; placeholder?: string; value: string[]; options: { id: string; name: string }[]; onChange: (ids: string[]) => void }) {
    return (
        <Autocomplete aria-label={ariaLabel ?? label} className="w-full" placeholder={placeholder ?? `选择${label}`} selectionMode="multiple" value={value} variant="secondary" onChange={(keys) => onChange(keysToStrings(keys))}>
            {label && <Label>{label}</Label>}
            <Autocomplete.Trigger>
                <Autocomplete.Value>
                    {({ defaultChildren, isPlaceholder, state }: any) => {
                        if (isPlaceholder || state.selectedItems.length === 0) return defaultChildren;
                        return (
                            <TagGroup size="sm"><TagGroup.List>
                                {state.selectedItems.map((item: any) => <Tag key={item.key} id={item.key}>{item.textValue}</Tag>)}
                            </TagGroup.List></TagGroup>
                        );
                    }}
                </Autocomplete.Value>
                <Autocomplete.Indicator />
            </Autocomplete.Trigger>
            <Autocomplete.Popover>
                <ListBox>
                    {options.map((option) => (
                        <ListBox.Item key={option.id} id={option.id} textValue={option.name}>{option.name}<ListBox.ItemIndicator /></ListBox.Item>
                    ))}
                </ListBox>
            </Autocomplete.Popover>
        </Autocomplete>
    );
}

function TaskStepFields({ step, onChange }: { step: TaskStep; onChange: (step: TaskStep) => void }) {
    return (
        <div className="flex flex-col gap-4">
            <Select className="w-full" variant="secondary" value={step.type} onChange={(key) => {
                const type = String(key) as TaskStep["type"];
                if (type === step.type) return;
                onChange({ ...step, type, params: defaultStepParams(type) });
            }}>
                <Label>检测类型</Label>
                <Select.Trigger><Select.Value /><Select.Indicator /></Select.Trigger>
                <Select.Popover>
                    <ListBox>
                        <ListBox.Item id="delay">延迟</ListBox.Item>
                        <ListBox.Item id="speed">测速</ListBox.Item>
                        <ListBox.Item id="country">落地</ListBox.Item>
                    </ListBox>
                </Select.Popover>
            </Select>
            <Select className="w-full" variant="secondary" value={step.order} onChange={(key) => {
                if (key === 0 || key === 1 || key === 2) onChange({ ...step, order: key });
            }}>
                <Label>处理顺序</Label>
                <Select.Trigger><Select.Value /><Select.Indicator /></Select.Trigger>
                <Select.Popover><ListBox><ListBox.Item id={0}>不排序</ListBox.Item><ListBox.Item id={1}>延迟优先</ListBox.Item><ListBox.Item id={2}>速度优先</ListBox.Item></ListBox></Select.Popover>
            </Select>
            {step.type !== "country" && (
                <TextField value={step.params?.url ?? ""} onChange={(value) => onChange({ ...step, params: { ...(step.params ?? {}), url: value } })}>
                    <Label>请求地址</Label>
                    <Input variant="secondary" placeholder="https://example.com" />
                </TextField>
            )}
            <NumberField label="超时 ms" value={step.params?.timeout_ms} onChange={(value) => onChange({ ...step, params: { ...(step.params ?? {}), timeout_ms: value } })} />
            <NumberField label="并发" value={step.concurrency} onChange={(value) => onChange({ ...step, concurrency: value })} />
            <div className="flex items-center">
                <span className="text-sm text-foreground">跳过已有检测结果</span>
                <div className="flex-1" />
                <Switch isSelected={step.skip_existing === 1} onChange={() => onChange({ ...step, skip_existing: step.skip_existing === 1 ? 0 : 1 })}>
                    <Switch.Content><Switch.Control><Switch.Thumb /></Switch.Control></Switch.Content>
                </Switch>
            </div>
            <div className="flex items-center">
                <span className="text-sm text-foreground">失败时删除订阅节点</span>
                <div className="flex-1" />
                <Switch isSelected={step.node_pool_delete === 1} onChange={() => onChange({ ...step, node_pool_delete: step.node_pool_delete === 1 ? 0 : 1 })}>
                    <Switch.Content><Switch.Control><Switch.Thumb /></Switch.Control></Switch.Content>
                </Switch>
            </div>
            <NumberField label="通过数量" value={step.pass.limit} onChange={(value) => onChange({ ...step, pass: { ...step.pass, limit: value } })} />
            {step.type === "delay" && <NumberField label="尝试次数" value={step.params?.attempts} onChange={(value) => onChange({ ...step, params: { ...(step.params ?? {}), attempts: value } })} />}
            {step.type === "speed" && <NumberField label="最大读取 kb" value={step.params?.max_kb} onChange={(value) => onChange({ ...step, params: { ...(step.params ?? {}), max_kb: value } })} />}
            {step.type === "speed" && <NumberField label="读取时长 ms" value={step.params?.max_duration_ms} onChange={(value) => onChange({ ...step, params: { ...(step.params ?? {}), max_duration_ms: value } })} />}
        </div>
    );
}

function NumberField({ label, value, onChange }: { label: string; value?: number; onChange: (value: number) => void }) {
    return (
        <TextField value={value ? String(value) : ""} onChange={(next) => onChange(Number(next))}>
            <Label>{label}</Label>
            <Input type="number" variant="secondary" />
        </TextField>
    );
}
