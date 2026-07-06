import { Switch, Label, Select, ListBox } from "@heroui/react";

export function AppearanceSetting({ val, set }: { val: (key: string) => string; set: (key: string, value: string) => void }) {
    return (
        <div className="settings-category">
            <div className="text-foreground/85 mb-2.5 flex items-center gap-2 px-1 text-base font-semibold tracking-tight">
                <span className="flex-1">外观设置</span>
            </div>
            <div className="bg-surface grid grid-cols-1 overflow-hidden rounded-xl">
                {/* 自动切换主题 */}
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <span className="text-foreground text-sm shrink-0 grow">跟随系统主题</span>
                    <Switch
                        isSelected={val("theme_auto") !== "0"}
                        onChange={() => set("theme_auto", val("theme_auto") !== "0" ? "0" : "1")}
                    >
                        <Switch.Control>
                            <Switch.Thumb />
                        </Switch.Control>
                    </Switch>
                </div>

                {/* 主题选择 */}
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <Label className="text-foreground text-sm shrink-0 grow">主题</Label>
                    <Select
                        className="w-48"
                        variant="secondary"
                        placeholder="选择主题"
                        value={val("theme") || "light"}
                        onChange={(k) => set("theme", String(k ?? ""))}
                        isDisabled={val("theme_auto") !== "0"}
                    >
                        <Select.Trigger>
                            <Select.Value />
                            <Select.Indicator />
                        </Select.Trigger>
                        <Select.Popover>
                            <ListBox>
                                <ListBox.Item id="light">浅色</ListBox.Item>
                                <ListBox.Item id="dark">深色</ListBox.Item>
                            </ListBox>
                        </Select.Popover>
                    </Select>
                </div>
            </div>
        </div>
    );
}
