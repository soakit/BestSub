// 字节数转可读单位，用于流量与测速展示
export function formatBytes(bytes: number): string {
    if (bytes <= 0) return "0 B";
    const units = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / 1024 ** i).toFixed(i > 0 ? 1 : 0)} ${units[i]}`;
}

// 相对时间；空值或后端零值时间戳(0001-...)显示占位符 "-"
export function formatRelativeTime(s: string): string {
    if (!s || s.startsWith("0001")) return "-";
    const d = new Date(s);
    if (isNaN(d.getTime())) return "-";
    const diff = Date.now() - d.getTime();
    if (diff < 0) return "未来";
    if (diff < 60_000) return "刚刚";
    if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}分钟前`;
    if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}小时前`;
    if (diff < 172_800_000) return "昨天";
    return `${String(d.getMonth() + 1).padStart(2, "0")}/${String(d.getDate()).padStart(2, "0")}`;
}
