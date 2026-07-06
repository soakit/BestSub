import { useState } from "react";
import { Switch, Input, Label, Select, ListBox } from "@heroui/react";
import { useInterfaces } from "../../api/settings";

export function GeneralSetting({ val, set }: { val: (key: string) => string; set: (key: string, value: string) => void }) {
    const { data: ifaces } = useInterfaces();
    const [pending, setPending] = useState<Record<string, string>>({});

    return (
        <div className="settings-category">
            <div className="text-foreground/85 mt-1 mb-2.5 flex items-center gap-2 px-1 text-base font-semibold tracking-tight">
                <span className="flex-1">通用设置</span>
            </div>
            <div className="bg-surface grid grid-cols-1 overflow-hidden rounded-xl">
                {/* 全局代理 */}
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <span className="text-foreground text-sm shrink-0 grow">全局代理</span>
                    <Switch
                        isSelected={val("proxy_enable") === "1"}
                        onChange={() => set("proxy_enable", val("proxy_enable") === "1" ? "0" : "1")}
                    >
                        <Switch.Control>
                            <Switch.Thumb />
                        </Switch.Control>
                    </Switch>
                </div>

                {/* 代理地址 */}
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <Label className="text-foreground text-sm shrink-0 grow">代理地址</Label>
                    <Input
                        className="w-48"
                        value={pending.proxy_url ?? val("proxy_url")}
                        placeholder="http://127.0.0.1:7890"
                        variant="secondary"
                        onChange={(e) => setPending((prev) => ({ ...prev, proxy_url: e.target.value }))}
                        onBlur={() => set("proxy_url", pending.proxy_url ?? "")}
                    />
                </div>

                {/* 订阅转换地址 */}
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <Label className="text-foreground text-sm shrink-0 grow">订阅转换地址</Label>
                    <Input
                        className="w-48"
                        value={pending.sub_convert_url ?? val("sub_convert_url")}
                        placeholder="https://example.com/sub"
                        variant="secondary"
                        onChange={(e) => setPending((prev) => ({ ...prev, sub_convert_url: e.target.value }))}
                        onBlur={() => set("sub_convert_url", pending.sub_convert_url ?? "")}
                    />
                </div>

                {/* 订阅转换使用代理 */}
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <span className="text-foreground text-sm shrink-0 grow">订阅转换使用代理</span>
                    <Switch
                        isSelected={val("sub_convert_proxy") === "1"}
                        onChange={() => set("sub_convert_proxy", val("sub_convert_proxy") === "1" ? "0" : "1")}
                    >
                        <Switch.Control>
                            <Switch.Thumb />
                        </Switch.Control>
                    </Switch>
                </div>

                {/* 绑定网卡 */}
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <Label className="text-foreground text-sm shrink-0 grow">绑定网卡</Label>
                    <Select
                        className="w-48"
                        variant="secondary"
                        placeholder="选择网卡"
                        value={val("bind_interface") || null}
                        onChange={(k) => set("bind_interface", String(k ?? ""))}
                    >
                        <Select.Trigger>
                            <Select.Value />
                            <Select.Indicator />
                        </Select.Trigger>
                        <Select.Popover>
                            <ListBox>
                                {(ifaces ?? []).map((iface) => (
                                    <ListBox.Item key={iface.name} id={iface.name}>
                                        {iface.name}({iface.ip})
                                    </ListBox.Item>
                                ))}
                            </ListBox>
                        </Select.Popover>
                    </Select>
                </div>
            </div>
        </div>
    );
}
