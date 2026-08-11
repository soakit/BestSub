import { useState } from "react";
import { Input, Label } from "@heroui/react";
import { Globe } from "@gravity-ui/icons";

export function Dns({ val, set }: { val: (key: string) => string; set: (key: string, value: string) => void }) {
    const [pending, setPending] = useState<Record<string, string>>({});

    return (
        <div className="settings-category">
            <div className="text-foreground/85 mt-1 mb-2.5 flex items-center gap-2 px-1 text-base font-semibold tracking-tight">
                <Globe className="size-4 shrink-0" />
                <span className="flex-1">DNS</span>
            </div>
            <div className="bg-surface grid grid-cols-1 overflow-hidden rounded-xl">
                {/* 默认 DNS */}
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <Label className="text-foreground text-sm shrink-0 grow">默认 DNS</Label>
                    <Input
                        className="w-64"
                        value={pending.dns_default ?? val("dns_default")}
                        placeholder="119.29.29.29,223.5.5.5"
                        variant="secondary"
                        onChange={(e) => setPending((prev) => ({ ...prev, dns_default: e.target.value }))}
                        onBlur={(e) => set("dns_default", e.currentTarget.value)}
                    />
                </div>

                {/* 主 DNS */}
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <Label className="text-foreground text-sm shrink-0 grow">主 DNS</Label>
                    <Input
                        className="w-64"
                        value={pending.dns_main ?? val("dns_main")}
                        placeholder="https://doh.pub/dns-query"
                        variant="secondary"
                        onChange={(e) => setPending((prev) => ({ ...prev, dns_main: e.target.value }))}
                        onBlur={(e) => set("dns_main", e.currentTarget.value)}
                    />
                </div>
            </div>
        </div>
    );
}
