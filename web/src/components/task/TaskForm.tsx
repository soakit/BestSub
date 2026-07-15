import { useMemo, useState } from "react";
import type { Key } from "@heroui/react";
import { Autocomplete, Button, Disclosure, Drawer, Form, Input, Label, ListBox, Select, Switch, Tag, TagGroup, TextField } from "@heroui/react";
import { Pencil, Plus, TrashBin } from "@gravity-ui/icons";
import { useNodes } from "../../api/node";
import { useStorages } from "../../api/storage";
import { useSubscription } from "../../api/sub";
import { useTags } from "../../api/tags";
import { type Task, type TaskConfig, type TaskStep, useCreateTask, useUpdateTask } from "../../api/task";

type TaskSourceType = "subscription" | "node" | "tag" | "result_task";
type LandingSourceType = "subscription" | "node";

const taskSourceTypes: { id: TaskSourceType; name: string }[] = [
    { id: "subscription", name: "订阅" },
    { id: "tag", name: "标签" },
    { id: "node", name: "节点" },
    { id: "result_task", name: "任务" },
];

const landingSourceTypes: { id: LandingSourceType; name: string }[] = [
    { id: "subscription", name: "订阅" },
    { id: "node", name: "节点" },
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
    custom_landing_node_enable: 0,
    landing_subscriptions: [],
    landing_nodes: [],
    storage_enable: 0,
    storage_id: "",
    save_path: "",
    node_rename_expression: "",
};

function newStep(): TaskStep {
    return {
        type: "delay",
        params: defaultStepParams("delay"),
        concurrency: 8,
        node_pool_delete: 0,
        pass: {},
        order: "none",
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

function initialLandingRows(task?: Task): LandingSourceType[] {
    if (!task) return ["subscription"];
    const rows = [
        ...(task.landing_subscriptions.length > 0 ? ["subscription" as const] : []),
        ...(task.landing_nodes.length > 0 ? ["node" as const] : []),
    ];
    return rows.length > 0 ? rows : ["subscription"];
}

export function TaskForm({ task, tasks, onClose }: { task?: Task; tasks: Task[]; onClose: () => void }) {
    const { data: subscriptions = [] } = useSubscription();
    const { data: nodes = [] } = useNodes();
    const { data: tags = [] } = useTags();
    const { data: storages = [] } = useStorages();
    const createTask = useCreateTask();
    const updateTask = useUpdateTask();
    const [editingStep, setEditingStep] = useState<number | null>(null);
    const [formState, setFormState] = useState<TaskConfig>(
        task
            ? {
                ...defaultConfig,
                ...task,
                steps: task.steps.map((step) => ({
                    ...step,
                    params: step.type === "country"
                        ? { timeout_ms: step.params?.timeout_ms }
                        : { ...defaultStepParams(step.type), ...(step.params ?? {}) },
                })),
            }
            : { ...defaultConfig, steps: [newStep()] }
    );
    const [sourceRows, setSourceRows] = useState<TaskSourceType[]>(() => initialSourceRows(task));
    const [landingSourceRows, setLandingSourceRows] = useState<LandingSourceType[]>(() => initialLandingRows(task));
    const resultTasks = useMemo(() => tasks.filter((t) => t.id !== task?.id), [tasks, task?.id]);

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

    const getLandingSourceValue = (type: LandingSourceType) => {
        if (type === "subscription") return formState.landing_subscriptions.map((sub) => sub.id);
        return formState.landing_nodes.map((node) => node.id);
    };

    const getLandingSourceOptions = (type: LandingSourceType) => {
        if (type === "subscription") return subscriptions.map((sub) => ({ id: sub.id, name: sub.name }));
        return nodes.map((node) => ({ id: node.id, name: node.name || node.id }));
    };

    const setLandingSourceValue = (type: LandingSourceType, ids: string[]) => {
        if (type === "subscription") {
            setForm("landing_subscriptions", ids.map((id) => ({ id })));
            return;
        }
        setForm("landing_nodes", ids.map((id) => ({ id })));
    };

    const clearLandingSourceValue = (type: LandingSourceType) => {
        if (type === "subscription") {
            setForm("landing_subscriptions", []);
            return;
        }
        setForm("landing_nodes", []);
    };

    const handleSubmit = (e: React.SyntheticEvent<HTMLFormElement>) => {
        e.preventDefault();
        const customLandingEnabled =
            formState.custom_landing_node_enable === 1 &&
            (formState.landing_subscriptions.length > 0 || formState.landing_nodes.length > 0);

        const payload: TaskConfig = {
            ...formState,
            name: formState.name.trim(),
            cron_expr: formState.auto_run === 1 ? formState.cron_expr.trim() : "",
            subscriptions: formState.subscriptions.map((sub) => ({ id: sub.id })),
            nodes: formState.nodes.map((node) => ({ id: node.id })),
            tags: formState.tags.map((tag) => ({ id: tag.id })),
            result_tasks: formState.result_tasks.map((task) => ({ id: task.id })),
            custom_landing_node_enable: customLandingEnabled ? 1 : 0,
            landing_subscriptions: customLandingEnabled ? formState.landing_subscriptions.map((sub) => ({ id: sub.id })) : [],
            landing_nodes: customLandingEnabled ? formState.landing_nodes.map((node) => ({ id: node.id })) : [],
            storage_id: formState.storage_enable === 1 ? formState.storage_id : "",
            save_path: formState.storage_enable === 1 ? formState.save_path.trim() : "",
            node_rename_expression: formState.storage_enable === 1 ? formState.node_rename_expression.trim() : "",
        };
        if (!payload.name || payload.steps.length === 0) return;
        if (task) {
            updateTask.mutate({ ...payload, id: task.id }, { onSuccess: onClose });
        } else {
            createTask.mutate(payload, { onSuccess: onClose });
        }
    };

    return (
        <Form id="task-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={handleSubmit}>
            <TextField isRequired value={formState.name} onChange={(value) => setForm("name", value)}>
                <Label>任务名称</Label>
                <Input name="name" placeholder="我的任务" variant="secondary" />
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
                <span className="text-sm font-medium text-foreground">前置节点来源</span>
                {sourceRows.map((type, index) => (
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
                                        {taskSourceTypes.filter((source) => source.id === type || !sourceRows.includes(source.id)).map((source) => (
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
                <Button variant="ghost" className="w-full" isDisabled={sourceRows.length >= taskSourceTypes.length} onPress={() => setSourceRows((rows) => [...rows, taskSourceTypes.find((source) => !rows.includes(source.id))!.id])}>
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
                    <Disclosure.Content className="!overflow-visible">
                        <div className="flex flex-col gap-3 pt-2">
                            {landingSourceRows.map((type, index) => (
                                <div key={index} className="flex items-center gap-2">
                                    <div className="grid min-w-0 flex-1 gap-2 sm:grid-cols-[6rem_minmax(0,1fr)] sm:items-center">
                                        <Select className="w-full" aria-label="落地来源类型" variant="secondary" value={type} onChange={(key) => {
                                            const next = String(key) as LandingSourceType;
                                            if (next === type) return;
                                            clearLandingSourceValue(type);
                                            setLandingSourceRows((rows) => rows.map((row, i) => (i === index ? next : row)));
                                        }}>
                                            <Select.Trigger><Select.Value /><Select.Indicator /></Select.Trigger>
                                            <Select.Popover>
                                                <ListBox>
                                                    {landingSourceTypes.filter((source) => source.id === type || !landingSourceRows.includes(source.id)).map((source) => (
                                                        <ListBox.Item key={source.id} id={source.id}>{source.name}</ListBox.Item>
                                                    ))}
                                                </ListBox>
                                            </Select.Popover>
                                        </Select>
                                        <TaskSourcePicker ariaLabel="选择落地来源" placeholder={`选择${landingSourceTypes.find((source) => source.id === type)!.name}`} value={getLandingSourceValue(type)} options={getLandingSourceOptions(type)} onChange={(ids) => setLandingSourceValue(type, ids)} />
                                    </div>
                                    <Button isIconOnly size="sm" variant="ghost" className="text-muted hover:text-danger" onPress={() => {
                                        clearLandingSourceValue(type);
                                        setLandingSourceRows((rows) => rows.filter((_, i) => i !== index));
                                    }}>
                                        <TrashBin className="size-4" />
                                    </Button>
                                </div>
                            ))}
                            <Button variant="ghost" className="w-full" isDisabled={landingSourceRows.length >= landingSourceTypes.length} onPress={() => setLandingSourceRows((rows) => [...rows, landingSourceTypes.find((source) => !rows.includes(source.id))!.id])}>
                                <Plus className="size-4" />
                                添加来源
                            </Button>
                        </div>
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
                    <Select className="w-full" variant="secondary" value={formState.storage_id} onChange={(key) => setForm("storage_id", String(key))}>
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
                    <TextField value={formState.save_path} onChange={(value) => setForm("save_path", value)}>
                        <Label>保存路径</Label>
                        <Input name="save_path" placeholder="/nodes.yaml" variant="secondary" />
                    </TextField>
                    <TextField value={formState.node_rename_expression} onChange={(value) => setForm("node_rename_expression", value)}>
                        <Label>重命名表达式</Label>
                        <Input name="node_rename_expression" placeholder="{{index}}-{{name}}" variant="secondary" />
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
            <Select className="w-full" variant="secondary" value={step.order || "none"} onChange={(key) => onChange({ ...step, order: String(key) as TaskStep["order"] })}>
                <Label>处理顺序</Label>
                <Select.Trigger><Select.Value /><Select.Indicator /></Select.Trigger>
                <Select.Popover><ListBox><ListBox.Item id="none">不排序</ListBox.Item><ListBox.Item id="delay">延迟优先</ListBox.Item><ListBox.Item id="speed">速度优先</ListBox.Item></ListBox></Select.Popover>
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
