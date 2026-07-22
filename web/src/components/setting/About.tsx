import { CircleInfo, LogoGithub } from "@gravity-ui/icons";
import packageJson from "../../../package.json";
import { useVersion } from "../../api/settings";

export function About() {
    const { data: backendVersion, isError } = useVersion();

    return (
        <div className="settings-category">
            <div className="text-foreground/85 mt-1 mb-2.5 flex items-center gap-2 px-1 text-base font-semibold tracking-tight">
                <CircleInfo className="size-4 shrink-0" />
                <span className="flex-1">关于</span>
            </div>
            <div className="bg-surface grid grid-cols-1 overflow-hidden rounded-xl">
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <span className="text-foreground grow text-sm">前端版本</span>
                    <span className="text-muted text-sm">{packageJson.version}</span>
                </div>
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <span className="text-foreground grow text-sm">后端版本</span>
                    <span className="text-muted text-sm">{isError ? "获取失败" : backendVersion ?? "加载中"}</span>
                </div>
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <span className="text-foreground grow text-sm">项目主页</span>
                    <a
                        href="https://github.com/bestruirui/bestsub"
                        target="_blank"
                        rel="noreferrer"
                        className="text-muted hover:text-foreground flex items-center gap-1.5 text-sm"
                    >
                        <LogoGithub className="size-4" />
                        GitHub
                    </a>
                </div>
            </div>
        </div>
    );
}
