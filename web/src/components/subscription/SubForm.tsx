import { useImperativeHandle, useState } from "react";
import type { Key } from "@heroui/react";
import { Button, Switch, Modal, Form, TextField, Label, Input, useOverlayState, TextArea, TagGroup, Tag, Disclosure, CloseButton, Autocomplete, ListBox } from "@heroui/react";
import { useCreateSubscription, useUpdateSubscription, type Subscription, type SubscriptionConfig } from "../../api/sub";
import { TagSelector } from "../common/TagSelector";

// mihomo 支持的代理协议
const PROTOCOLS = [
    { id: "ss", name: "Shadowsocks" },
    { id: "ssr", name: "ShadowsocksR" },
    { id: "socks5", name: "SOCKS5" },
    { id: "http", name: "HTTP" },
    { id: "vmess", name: "VMess" },
    { id: "vless", name: "VLESS" },
    { id: "snell", name: "Snell" },
    { id: "trojan", name: "Trojan" },
    { id: "hysteria", name: "Hysteria" },
    { id: "hysteria2", name: "Hysteria2" },
    { id: "wireguard", name: "WireGuard" },
    { id: "tuic", name: "TUIC" },
    { id: "ssh", name: "SSH" },
    { id: "mieru", name: "Mieru" },
    { id: "anytls", name: "AnyTLS" },
];

const defaultConfig: SubscriptionConfig = {
    url_type: 0,
    url: [],
    enable: 1,
    name: "",
    tags: [],
    header: {},
    auto_update: 1,
    cron_expr: "0 */6 * * *",
    proxy_mode: 0,
    proxy_url: "",
    protocol_filter_mode: 0,
    protocol_filter: [],
};

export function SubForm({ ref }: { ref?: React.Ref<(sub?: Subscription) => void> }) {
    const state = useOverlayState();
    const [editing, setEditing] = useState<Subscription | null>(null);
    const [formState, setFormState] = useState<Partial<SubscriptionConfig>>({});
    const [headerPairs, setHeaderPairs] = useState<[string, string][]>([["user-agent", "clash.meta/v1.19.27"]]);
    const createSub = useCreateSubscription();
    const updateSub = useUpdateSubscription();

    // 更新指定位置的请求头 key 或 value
    const updateHeader = (index: number, field: "key" | "value", val: string) =>
        setHeaderPairs((prev) => prev.map((pair, i) => (i === index ? (field === "key" ? [val, pair[1]] : [pair[0], val]) as [string, string] : pair)));

    const setForm = <K extends keyof SubscriptionConfig>(key: K, value: SubscriptionConfig[K]) => {
        setFormState((prev) => ({ ...prev, [key]: value }));
    };

    useImperativeHandle(ref, () => (sub?: Subscription) => {
        setEditing(sub ?? null);
        if (sub) {
            setFormState(sub);
            setHeaderPairs(Object.entries(sub.header ?? {}));
        } else {
            setFormState({ ...defaultConfig });
            setHeaderPairs([["user-agent", "clash.meta/v1.19.27"]]);
        }
        state.open();
    });

    const handleSubmit = (e: React.SyntheticEvent<HTMLFormElement>) => {
        e.preventDefault();
        const fd = new FormData(e.currentTarget);
        if (!fd.get("name") || !fd.get("url")) return;

        const payload: SubscriptionConfig = {
            ...defaultConfig,
            ...formState,
            name: String(fd.get("name")),
            url: String(fd.get("url")).split("\n").map((s) => s.trim()).filter(Boolean),
            header: Object.fromEntries(headerPairs.filter(([k]) => k.trim())),
        };
        if (editing) {
            updateSub.mutate({ ...payload, id: editing.id }, { onSuccess: () => state.setOpen(false) });
        } else {
            createSub.mutate(payload, { onSuccess: () => state.setOpen(false) });
        }
    };

    return (
        <Modal state={state} data-scrollbar="none" >
            <Modal.Backdrop variant="blur">
                <Modal.Container>
                    <Modal.Dialog>
                        <Modal.CloseTrigger />
                        <Modal.Header>
                            <Modal.Heading>{editing ? "编辑订阅" : "添加订阅"}</Modal.Heading>
                        </Modal.Header>
                        <Modal.Body>
                            <Form id="sub-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={handleSubmit}>
                                <TextField isRequired defaultValue={editing?.name ?? ""}>
                                    <Label>订阅名称</Label>
                                    <Input name="name" placeholder="我的订阅" variant="secondary" />
                                </TextField>

                                <TagGroup aria-label="链接类型" selectionMode="single" selectedKeys={new Set([String(formState.url_type ?? 0)])} onSelectionChange={(keys) => setForm("url_type", Number([...keys][0]))}>
                                    <Label>链接类型</Label>
                                    <TagGroup.List className="flex w-full gap-2">
                                        <Tag id="0" className="flex-1 justify-center">直接获取</Tag>
                                        <Tag id="1" className="flex-1 justify-center">链接列表</Tag>
                                    </TagGroup.List>
                                </TagGroup>

                                <TextField isRequired defaultValue={editing ? editing.url.join("\n") : ""}>
                                    <Label>订阅地址</Label>
                                    <TextArea name="url" placeholder={formState.url_type === 1 ? "https://example.com/sublist" : "https://example.com/sub"} variant="secondary" />
                                </TextField>

                                <TagSelector value={formState.tags ?? []} onChange={(tags) => setForm("tags", tags)} />

                                <Disclosure className="w-full" isExpanded={formState.auto_update === 1}>
                                    <Disclosure.Heading className="flex items-center">
                                        <span className="text-foreground text-sm shrink-0 grow">自动更新</span>
                                        <Switch
                                            isSelected={formState.auto_update === 1}
                                            onChange={() => setForm("auto_update", formState.auto_update === 1 ? 0 : 1)}
                                        >
                                            <Switch.Content>
                                                <Switch.Control><Switch.Thumb /></Switch.Control>
                                            </Switch.Content>
                                        </Switch>
                                    </Disclosure.Heading>
                                    <Disclosure.Content className="!overflow-visible">
                                        <TextField defaultValue={editing?.cron_expr || "0 */6 * * *"} className="pt-2">
                                            <Label>Cron 表达式</Label>
                                            <Input name="cron_expr" placeholder="0 */6 * * *" variant="secondary" onChange={(e) => setForm("cron_expr", e.target.value)} />
                                        </TextField>
                                    </Disclosure.Content>
                                </Disclosure>

                                <Disclosure className="w-full">
                                    <Disclosure.Heading>
                                        <Disclosure.Trigger className="flex w-full items-center justify-between text-sm font-medium text-foreground">
                                            代理设置
                                            <Disclosure.Indicator />
                                        </Disclosure.Trigger>
                                    </Disclosure.Heading>
                                    <Disclosure.Content className="flex flex-col gap-4 !overflow-visible">
                                        <TagGroup aria-label="代理模式" className="pt-2" selectionMode="single" selectedKeys={new Set([String(formState.proxy_mode ?? 0)])} onSelectionChange={(keys) => setForm("proxy_mode", Number([...keys][0]))}>
                                            <TagGroup.List className="flex w-full gap-2">
                                                <Tag id="0" className="flex-1 justify-center">自动</Tag>
                                                <Tag id="1" className="flex-1 justify-center">禁用</Tag>
                                                <Tag id="2" className="flex-1 justify-center">启用</Tag>
                                            </TagGroup.List>
                                        </TagGroup>

                                        {formState.proxy_mode === 2 && (
                                            <TextField defaultValue={editing?.proxy_url ?? ""}>
                                                <Label>代理地址</Label>
                                                <Input name="proxy_url" placeholder="http://127.0.0.1:7890" variant="secondary" onChange={(e) => setForm("proxy_url", e.target.value)} />
                                            </TextField>
                                        )}
                                    </Disclosure.Content>
                                </Disclosure>

                                <Disclosure className="w-full">
                                    <Disclosure.Heading>
                                        <Disclosure.Trigger className="flex w-full items-center justify-between text-sm font-medium text-foreground">
                                            协议过滤
                                            <Disclosure.Indicator />
                                        </Disclosure.Trigger>
                                    </Disclosure.Heading>
                                    <Disclosure.Content className="flex flex-col gap-4 !overflow-visible">
                                        <TagGroup aria-label="协议过滤" className="pt-2" selectionMode="single" selectedKeys={new Set([String(formState.protocol_filter_mode ?? 0)])} onSelectionChange={(keys) => setForm("protocol_filter_mode", Number([...keys][0]))}>
                                            <TagGroup.List className="flex w-full gap-2">
                                                <Tag id="0" className="flex-1 justify-center">不启用</Tag>
                                                <Tag id="1" className="flex-1 justify-center">包含</Tag>
                                                <Tag id="2" className="flex-1 justify-center">排除</Tag>
                                            </TagGroup.List>
                                        </TagGroup>

                                        {formState.protocol_filter_mode !== 0 && (
                                            <Autocomplete
                                                className="w-full"
                                                placeholder="选择协议"
                                                selectionMode="multiple"
                                                variant="secondary"
                                                value={formState.protocol_filter ?? []}
                                                onChange={(keys: Key | Key[] | null) => setForm("protocol_filter", (Array.isArray(keys) ? keys : keys ? [keys] : []).map(String))}
                                            >
                                                <Label>协议列表</Label>
                                                <Autocomplete.Trigger>
                                                    <Autocomplete.Value>
                                                        {({ defaultChildren, isPlaceholder, state }: any) => {
                                                            if (isPlaceholder || state.selectedItems.length === 0) return defaultChildren;
                                                            return (
                                                                <TagGroup size="sm">
                                                                    <TagGroup.List>
                                                                        {state.selectedItems.map((item: any) => (
                                                                            <Tag key={item.key} id={item.key} className="bg-default">{item.textValue}</Tag>
                                                                        ))}
                                                                    </TagGroup.List>
                                                                </TagGroup>
                                                            );
                                                        }}
                                                    </Autocomplete.Value>
                                                    <Autocomplete.Indicator />
                                                </Autocomplete.Trigger>
                                                <Autocomplete.Popover>
                                                    <ListBox>
                                                        {PROTOCOLS.map((p) => (
                                                            <ListBox.Item key={p.id} id={p.id} textValue={p.name}>
                                                                {p.name}
                                                                <ListBox.ItemIndicator />
                                                            </ListBox.Item>
                                                        ))}
                                                    </ListBox>
                                                </Autocomplete.Popover>
                                            </Autocomplete>
                                        )}
                                    </Disclosure.Content>
                                </Disclosure>

                                <Disclosure className="w-full">
                                    <Disclosure.Heading>
                                        <Disclosure.Trigger className="flex w-full items-center justify-between text-sm font-medium text-foreground">
                                            请求头
                                            <Disclosure.Indicator />
                                        </Disclosure.Trigger>
                                    </Disclosure.Heading>
                                    <Disclosure.Content className="flex flex-col gap-2 pt-2">
                                        {headerPairs.map(([key, value], i) => (
                                            <div key={i} className="flex flex-col sm:flex-row sm:items-center gap-2">
                                                <Input
                                                    className="flex-1 min-w-0"
                                                    placeholder="Header-Name"
                                                    variant="secondary"
                                                    value={key}
                                                    onChange={(e) => updateHeader(i, "key", e.target.value)}
                                                />
                                                <div className="flex items-center gap-2">
                                                    <Input
                                                        className="flex-1 min-w-0"
                                                        placeholder="value"
                                                        variant="secondary"
                                                        value={value}
                                                        onChange={(e) => updateHeader(i, "value", e.target.value)}
                                                    />
                                                    <CloseButton onPress={() => setHeaderPairs((prev) => prev.filter((_, j) => j !== i))} />
                                                </div>
                                            </div>
                                        ))}
                                        <Button variant="ghost" className="w-full" onPress={() => setHeaderPairs((prev) => [...prev, ["", ""]])}>+ 添加请求头</Button>
                                    </Disclosure.Content>
                                </Disclosure>
                            </Form>
                        </Modal.Body>
                        <Modal.Footer>
                            <Button variant="ghost" slot="close">取消</Button>
                            <Button type="submit" form="sub-form" variant="primary">{editing ? "保存" : "添加"}</Button>
                        </Modal.Footer>
                    </Modal.Dialog>
                </Modal.Container>
            </Modal.Backdrop>
        </Modal>
    );
}
