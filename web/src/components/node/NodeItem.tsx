import { Pencil, TrashBin, Server, Thunderbolt, Globe } from "@gravity-ui/icons";
import { formatBytes } from "../../lib/format";
import type { Node } from "../../api/node";

// country_code 为 ISO 3166-1 alpha-2，转成国旗 emoji（两个字母映射到区域指示符）
function countryFlag(code: string): string {
    if (!/^[A-Za-z]{2}$/.test(code)) return "";
    return String.fromCodePoint(...[...code.toUpperCase()].map((c) => 0x1f1e6 + c.charCodeAt(0) - 65));
}

export function NodeItem({
    node,
    onEdit,
    onDelete,
}: {
    node: Node;
    onEdit: (node: Node) => void;
    onDelete: (node: Node) => void;
}) {
    // delay 为 0 表示未测试，其余按可用/不可用配色
    const delayLabel = node.delay > 0 ? `${node.delay}ms` : "-";
    const flag = countryFlag(node.country_code);

    return (
        <div className="bg-surface rounded-2xl p-4 flex flex-col h-full">
            <div className="flex justify-between items-start gap-4 mb-4">
                <h3 className="flex-1 min-w-0 text-foreground text-xl leading-snug line-clamp-1" title={node.name}>{node.name || "未命名节点"}</h3>
                <div className="flex gap-0.5 bg-surface-secondary rounded-lg p-0.5 shrink-0">
                    <button onClick={() => onEdit(node)} className="p-1.5 text-muted hover:text-accent hover:bg-surface rounded-md transition-all">
                        <Pencil className="size-3.5" />
                    </button>
                    <button onClick={() => onDelete(node)} className="p-1.5 text-muted hover:text-danger hover:bg-surface rounded-md transition-all">
                        <TrashBin className="size-3.5" />
                    </button>
                </div>
            </div>

            <div className="grid grid-cols-2 gap-2 flex-1">
                <div className="bg-surface-secondary rounded-xl p-3 flex flex-col gap-2">
                    <div className="flex items-center gap-1 text-accent"><Thunderbolt className="size-4" /><span className="text-xs font-medium text-muted">延迟</span></div>
                    <div className="flex-1 flex items-end justify-end text-xl text-foreground leading-none">{delayLabel}</div>
                </div>
                <div className="bg-surface-secondary rounded-xl p-3 flex flex-col gap-2">
                    <div className="flex items-center gap-1 text-accent"><Globe className="size-4" /><span className="text-xs font-medium text-muted">落地</span></div>
                    <div className="flex-1 flex items-end justify-end text-xl text-foreground leading-none">{flag || node.country_code || "-"}</div>
                </div>
                <div className="bg-surface-secondary rounded-xl p-3 flex flex-col gap-2">
                    <div className="flex items-center gap-1 text-accent"><Server className="size-4" /><span className="text-xs font-medium text-muted">测速</span></div>
                    <div className="flex-1 flex items-end justify-end text-xl text-foreground leading-none">{formatBytes(node.download_speed)}/s</div>
                </div>
            </div>
        </div>
    );
}
