import { toast } from "@heroui/react";
import { Copy, Pencil, Server, TrashBin } from "@gravity-ui/icons";
import type { Share } from "../../api/share";

export function ShareItem({ share, onEdit, onDelete }: { share: Share; onEdit: (share: Share) => void; onDelete: (share: Share) => void }) {
    const shareUrl = `${window.location.origin}/share/${share.token}`;

    const copyLink = async () => {
        try {
            if (window.isSecureContext && navigator.clipboard) {
                await navigator.clipboard.writeText(shareUrl);
            } else {
                const textarea = document.createElement("textarea");
                textarea.value = shareUrl;
                textarea.readOnly = true;
                textarea.style.position = "fixed";
                textarea.style.opacity = "0";
                document.body.appendChild(textarea);
                textarea.select();
                const copied = document.execCommand("copy");
                textarea.remove();
                if (!copied) throw new Error("复制命令执行失败");
            }
            toast.success(share.name, { description: "分享链接已复制" });
        } catch {
            toast.danger(share.name, { description: "复制失败，请手动复制" });
        }
    };

    return (
        <div className="flex h-full flex-col rounded-2xl bg-surface p-4">
            <div className="mb-4 flex items-start justify-between gap-4">
                <h3 className="min-w-0 flex-1 line-clamp-1 text-xl leading-snug text-foreground" title={share.name}>{share.name}</h3>
                <div className="flex shrink-0 gap-0.5 rounded-lg bg-surface-secondary p-0.5">
                    <button type="button" aria-label="复制分享链接" onClick={copyLink} className="rounded-md p-1.5 text-muted transition-all hover:bg-surface hover:text-success">
                        <Copy className="size-3.5" />
                    </button>
                    <button type="button" aria-label="编辑分享" onClick={() => onEdit(share)} className="rounded-md p-1.5 text-muted transition-all hover:bg-surface hover:text-accent">
                        <Pencil className="size-3.5" />
                    </button>
                    <button type="button" aria-label="删除分享" onClick={() => onDelete(share)} className="rounded-md p-1.5 text-muted transition-all hover:bg-surface hover:text-danger">
                        <TrashBin className="size-3.5" />
                    </button>
                </div>
            </div>

            <div className="flex flex-1 flex-col gap-2 rounded-xl bg-surface-secondary p-3">
                <div className="flex items-center gap-1 text-accent"><Server className="size-4" /><span className="text-xs font-medium text-muted">包含节点</span></div>
                <div className="flex flex-1 items-end justify-end text-xl leading-none text-foreground">{share.node_count}</div>
            </div>
        </div>
    );
}
