import { useEffect, useMemo, useState } from "react";
import type { Key } from "@heroui/react";
import { Autocomplete, Button, Disclosure, Dropdown, EmptyState, Form, Input, Label, ListBox, ListLayout, SearchField, Select, Tag, TagGroup, TextArea, TextField, Virtualizer, useFilter } from "@heroui/react";
import { FileArrowDown, Plus, TrashBin } from "@gravity-ui/icons";
import { useNodes } from "../../api/node";
import { useRenamePreview, useRenameTemplates } from "../../api/rename";
import { type Share, type ShareConfig, type ShareFilter, useCreateShare, useUpdateShare } from "../../api/share";
import { useSubscription } from "../../api/sub";
import { useTags } from "../../api/tags";
import { useTasks } from "../../api/task";
import { countryOptions } from "./countries";

type ShareSourceType = "subscription" | "node" | "tag" | "result_task"; // 分享支持的输入来源类型。

const shareSourceTypes: { id: ShareSourceType; name: string }[] = [ // 来源类型及界面名称。
    { id: "subscription", name: "订阅" },
    { id: "tag", name: "标签" },
    { id: "node", name: "节点" },
    { id: "result_task", name: "任务" },
];

const defaultConfig: ShareConfig = { // 新建分享时提交的最小配置。
    name: "",
    filter: {},
    node_rename_expression: "",
    subscriptions: [],
    nodes: [],
    tags: [],
    result_tasks: [],
};

function keysToStrings(keys: Key | Key[] | null) {
    return (Array.isArray(keys) ? keys : keys ? [keys] : []).map(String);
}

// 编辑时只恢复已有来源；新建时默认提供一行订阅来源。
function initialSourceRows(share?: Share): ShareSourceType[] {
    if (!share) return ["subscription"];
    const rows = [
        ...(share.subscriptions.length > 0 ? ["subscription" as const] : []),
        ...(share.nodes.length > 0 ? ["node" as const] : []),
        ...(share.tags.length > 0 ? ["tag" as const] : []),
        ...(share.result_tasks.length > 0 ? ["result_task" as const] : []),
    ];
    return rows.length > 0 ? rows : ["subscription"];
}

export function ShareForm({ share, onClose }: { share?: Share; onClose: () => void }) {
    const { data: subscriptions = [] } = useSubscription();
    const { data: nodes = [] } = useNodes();
    const { data: tags = [] } = useTags();
    const { data: tasks = [] } = useTasks();
    const { data: renameTemplates = [] } = useRenameTemplates();
    const renamePreview = useRenamePreview();
    const createShare = useCreateShare();
    const updateShare = useUpdateShare();
    const [formState, setFormState] = useState<ShareConfig>(() => // 当前弹窗内可提交的分享配置，仅保留后端允许编辑的字段。
        share ? {
            name: share.name,
            node_rename_expression: share.node_rename_expression,
            filter: {
                min_delay: share.filter.min_delay,
                max_delay: share.filter.max_delay,
                min_download_speed: share.filter.min_download_speed,
                max_download_speed: share.filter.max_download_speed,
                include_country_codes: share.filter.include_country_codes,
                exclude_country_codes: share.filter.exclude_country_codes,
            },
            subscriptions: share.subscriptions.map(({ id }) => ({ id })),
            nodes: share.nodes.map(({ id }) => ({ id })),
            tags: share.tags.map(({ id }) => ({ id })),
            result_tasks: share.result_tasks.map(({ id }) => ({ id })),
        } : { ...defaultConfig, filter: {} }
    );
    const [sourceRows, setSourceRows] = useState<ShareSourceType[]>(() => initialSourceRows(share)); // 当前显示的来源选择行。

    useEffect(() => {
        renamePreview.reset();
        if (!formState.node_rename_expression.trim()) return;
        const timeout = window.setTimeout(() => renamePreview.mutate(formState.node_rename_expression.trim()), 400);
        return () => window.clearTimeout(timeout);
    }, [formState.node_rename_expression]);

    const setForm = <K extends keyof ShareConfig>(key: K, value: ShareConfig[K]) => {
        setFormState((prev) => ({ ...prev, [key]: value }));
    };

    const setFilter = <K extends keyof ShareFilter>(key: K, value: ShareFilter[K]) => {
        setFormState((prev) => ({ ...prev, filter: { ...prev.filter, [key]: value } }));
    };

    const getSourceValue = (type: ShareSourceType) => {
        if (type === "subscription") return formState.subscriptions.map((sub) => sub.id);
        if (type === "node") return formState.nodes.map((node) => node.id);
        if (type === "tag") return formState.tags.map((tag) => String(tag.id));
        return formState.result_tasks.map((task) => task.id);
    };

    const getSourceOptions = (type: ShareSourceType) => {
        if (type === "subscription") return subscriptions.map((sub) => ({ id: sub.id, name: sub.name }));
        if (type === "node") return nodes.map((node) => ({ id: node.id, name: node.name || node.id }));
        if (type === "tag") return tags.map((tag) => ({ id: String(tag.id), name: tag.name }));
        return tasks.map((task) => ({ id: task.id, name: task.name }));
    };

    // 每类来源写回后端要求的仅含 id 的引用数组。
    const setSourceValue = (type: ShareSourceType, ids: string[]) => {
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

    const clearSourceValue = (type: ShareSourceType) => {
        if (type === "subscription") return setForm("subscriptions", []);
        if (type === "node") return setForm("nodes", []);
        if (type === "tag") return setForm("tags", []);
        setForm("result_tasks", []);
    };

    const handleSubmit = (event: React.SyntheticEvent<HTMLFormElement>) => {
        event.preventDefault();
        const payload: ShareConfig = {
            ...formState,
            name: formState.name.trim(),
            node_rename_expression: formState.node_rename_expression.trim(),
            filter: {
                min_delay: formState.filter.min_delay,
                max_delay: formState.filter.max_delay,
                min_download_speed: formState.filter.min_download_speed,
                max_download_speed: formState.filter.max_download_speed,
                include_country_codes: formState.filter.include_country_codes ?? [],
                exclude_country_codes: formState.filter.exclude_country_codes ?? [],
            },
            subscriptions: formState.subscriptions.map(({ id }) => ({ id })),
            nodes: formState.nodes.map(({ id }) => ({ id })),
            tags: formState.tags.map(({ id }) => ({ id })),
            result_tasks: formState.result_tasks.map(({ id }) => ({ id })),
        };
        if (!payload.name || payload.subscriptions.length + payload.nodes.length + payload.tags.length + payload.result_tasks.length === 0) return;
        if (share) {
            updateShare.mutate({ ...payload, id: share.id }, { onSuccess: onClose });
        } else {
            createShare.mutate(payload, { onSuccess: onClose });
        }
    };

    return (
        <Form id="share-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={handleSubmit}>
            <TextField isRequired value={formState.name} onChange={(value) => setForm("name", value)}>
                <Label>分享名称</Label>
                <Input name="name" placeholder="我的分享" variant="secondary" />
            </TextField>

            <div className="flex flex-col gap-3">
                <span className="text-sm font-medium text-foreground">节点来源</span>
                {sourceRows.map((type, index) => (
                    <div key={type} className="flex items-center gap-2">
                        <div className="grid min-w-0 flex-1 gap-2 sm:grid-cols-[6rem_minmax(0,1fr)] sm:items-center">
                            <Select className="w-full" aria-label="来源类型" variant="secondary" value={type} onChange={(key) => {
                                const next = String(key) as ShareSourceType;
                                if (next === type) return;
                                clearSourceValue(type);
                                setSourceRows((rows) => rows.map((row, i) => (i === index ? next : row)));
                            }}>
                                <Select.Trigger><Select.Value /><Select.Indicator /></Select.Trigger>
                                <Select.Popover>
                                    <ListBox>
                                        {shareSourceTypes.filter((source) => source.id === type || !sourceRows.includes(source.id)).map((source) => (
                                            <ListBox.Item key={source.id} id={source.id}>{source.name}</ListBox.Item>
                                        ))}
                                    </ListBox>
                                </Select.Popover>
                            </Select>
                            <SourcePicker
                                placeholder={`选择${shareSourceTypes.find((source) => source.id === type)!.name}`}
                                value={getSourceValue(type)}
                                options={getSourceOptions(type)}
                                onChange={(ids) => setSourceValue(type, ids)}
                            />
                        </div>
                        <Button isIconOnly size="sm" variant="ghost" className="text-muted hover:text-danger" onPress={() => {
                            clearSourceValue(type);
                            setSourceRows((rows) => rows.filter((_, i) => i !== index));
                        }}>
                            <TrashBin className="size-4" />
                        </Button>
                    </div>
                ))}
                <Button variant="ghost" className="w-full" isDisabled={sourceRows.length >= shareSourceTypes.length} onPress={() => setSourceRows((rows) => [...rows, shareSourceTypes.find((source) => !rows.includes(source.id))!.id])}>
                    <Plus className="size-4" />
                    添加来源
                </Button>
            </div>

            <Disclosure className="w-full">
                <Disclosure.Heading>
                    <Disclosure.Trigger className="flex w-full items-center justify-between text-sm font-medium text-foreground">
                        节点筛选
                        <Disclosure.Indicator />
                    </Disclosure.Trigger>
                </Disclosure.Heading>
                <Disclosure.Content className="flex flex-col gap-4 !overflow-visible">
                    <div className="grid gap-4 pt-2 sm:grid-cols-2">
                        <OptionalNumberField label="最小延迟 ms" value={formState.filter.min_delay} onChange={(value) => setFilter("min_delay", value)} />
                        <OptionalNumberField label="最大延迟 ms" value={formState.filter.max_delay} onChange={(value) => setFilter("max_delay", value)} />
                        <OptionalNumberField label="最小速度 kb/s" value={formState.filter.min_download_speed} onChange={(value) => setFilter("min_download_speed", value)} />
                        <OptionalNumberField label="最大速度 kb/s" value={formState.filter.max_download_speed} onChange={(value) => setFilter("max_download_speed", value)} />
                    </div>
                    <CountryPicker label="包含国家" value={formState.filter.include_country_codes ?? []} onChange={(value) => setFilter("include_country_codes", value)} />
                    <CountryPicker label="排除国家" value={formState.filter.exclude_country_codes ?? []} onChange={(value) => setFilter("exclude_country_codes", value)} />
                </Disclosure.Content>
            </Disclosure>

            <TextField value={formState.node_rename_expression} onChange={(value) => setForm("node_rename_expression", value)}>
                <div className="flex items-center gap-2">
                    <Label className="flex-1">重命名表达式</Label>
                    <Dropdown>
                        <Button isIconOnly type="button" size="sm" variant="ghost" aria-label="选择重命名模板" isDisabled={renameTemplates.length === 0}>
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
                <TextArea rows={2} name="node_rename_expression" placeholder="{{.Country.NameZh}}-{{.Index}}" variant="secondary" />
                <span className={`text-xs ${renamePreview.isError ? "text-danger" : "text-muted"}`}>
                    预览：{renamePreview.isError ? renamePreview.error.message : renamePreview.data?.result}
                </span>
            </TextField>
        </Form>
    );
}

function SourcePicker({ placeholder, value, options, onChange }: { placeholder: string; value: string[]; options: { id: string; name: string }[]; onChange: (ids: string[]) => void }) {
    return (
        <Autocomplete aria-label="选择来源" className="w-full" placeholder={placeholder} selectionMode="multiple" value={value} variant="secondary" onChange={(keys) => onChange(keysToStrings(keys))}>
            <Autocomplete.Trigger>
                <Autocomplete.Value>
                    {({ defaultChildren, isPlaceholder, state }: any) => isPlaceholder || state.selectedItems.length === 0 ? defaultChildren : (
                        <TagGroup size="sm"><TagGroup.List>
                            {state.selectedItems.map((item: any) => <Tag key={item.key} id={item.key}>{item.textValue}</Tag>)}
                        </TagGroup.List></TagGroup>
                    )}
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

function CountryPicker({ label, value, onChange }: { label: string; value: string[]; onChange: (ids: string[]) => void }) {
    const [searchQuery, setSearchQuery] = useState("");
    const { contains } = useFilter({ sensitivity: "base" });
    const filteredCountries = useMemo(() => searchQuery ? countryOptions.filter((country) =>
        contains(country.nameZh, searchQuery) || contains(country.nameEn, searchQuery) || contains(country.id, searchQuery)
    ) : countryOptions, [contains, searchQuery]);

    return (
        <Autocomplete allowsEmptyCollection className="w-full" placeholder="选择国家" selectionMode="multiple" value={value} variant="secondary" onChange={(keys) => onChange(keysToStrings(keys))}>
            <Label>{label}</Label>
            <Autocomplete.Trigger>
                <Autocomplete.Value>
                    {({ defaultChildren, isPlaceholder, state }: any) => isPlaceholder || state.selectedItems.length === 0 ? defaultChildren : (
                        <TagGroup size="sm"><TagGroup.List>
                            {state.selectedItems.map((item: any) => <Tag key={item.key} id={item.key}>{item.textValue}</Tag>)}
                        </TagGroup.List></TagGroup>
                    )}
                </Autocomplete.Value>
                <Autocomplete.Indicator />
            </Autocomplete.Trigger>
            <Autocomplete.Popover>
                <Autocomplete.Filter inputValue={searchQuery} onInputChange={setSearchQuery}>
                    <SearchField autoFocus className="sticky top-0 z-10" name="country-search" variant="secondary">
                        <SearchField.Group>
                            <SearchField.SearchIcon />
                            <SearchField.Input placeholder="搜索国家" />
                            <SearchField.ClearButton />
                        </SearchField.Group>
                    </SearchField>
                    <Virtualizer layout={ListLayout} layoutOptions={{ rowHeight: 50 }}>
                        <ListBox items={filteredCountries} renderEmptyState={() => <EmptyState>没有匹配的国家</EmptyState>}>
                            {(country) => (
                                <ListBox.Item id={country.id} textValue={`${country.nameZh} · ${country.nameEn}`}>
                                    <div className="flex min-w-0 flex-col">
                                        <span className="truncate text-foreground">{country.nameZh}</span>
                                        <span className="truncate text-xs text-muted">{country.nameEn}</span>
                                    </div>
                                    <ListBox.ItemIndicator />
                                </ListBox.Item>
                            )}
                        </ListBox>
                    </Virtualizer>
                </Autocomplete.Filter>
            </Autocomplete.Popover>
        </Autocomplete>
    );
}

function OptionalNumberField({ label, value, onChange }: { label: string; value?: number; onChange: (value?: number) => void }) {
    return (
        <TextField value={value ? String(value) : ""} onChange={(next) => onChange(next === "" || Number.isNaN(Number(next)) ? undefined : Number(next))}>
            <Label>{label}</Label>
            <Input type="number" min="0" variant="secondary" />
        </TextField>
    );
}
