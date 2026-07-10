import { useState } from "react";
import { Button, Input, Label } from "@heroui/react";
import { Plus, Tag as TagIcon, TrashBin } from "@gravity-ui/icons";
import { useCreateTag, useDeleteTag, useTags } from "../../api/tags";

export function Tag() {
    const { data: tags = [] } = useTags();
    const createTag = useCreateTag();
    const deleteTag = useDeleteTag();
    const [name, setName] = useState("");

    return (
        <div className="settings-category">
            <div className="text-foreground/85 mt-1 mb-2.5 flex items-center gap-2 px-1 text-base font-semibold tracking-tight">
                <TagIcon className="size-4 shrink-0" />
                <span className="flex-1">标签</span>
            </div>
            <div className="bg-surface grid grid-cols-1 overflow-hidden rounded-xl">
                <form
                    className="flex min-h-11 items-center gap-3 px-4 py-2"
                    onSubmit={(e) => {
                        e.preventDefault();
                        if (!name.trim() || tags.some((tag) => tag.name === name.trim())) return;
                        createTag.mutate(name.trim(), { onSuccess: () => setName("") });
                    }}
                >
                    <Label className="text-foreground text-sm shrink-0 grow">新增标签</Label>
                    <Input
                        className="w-48"
                        value={name}
                        placeholder="标签名称"
                        variant="secondary"
                        onChange={(e) => setName(e.target.value)}
                    />
                    <Button
                        isIconOnly
                        type="submit"
                        variant="ghost"
                        isDisabled={!name.trim() || tags.some((tag) => tag.name === name.trim()) || createTag.isPending}
                    >
                        <Plus className="size-4" />
                    </Button>
                </form>

                {tags.map((tag) => (
                    <div key={tag.id} className="flex min-h-11 items-center gap-3 px-4 py-2">
                        <span className="text-foreground min-w-0 grow truncate text-sm" title={tag.name}>{tag.name}</span>
                        <Button
                            isIconOnly
                            size="sm"
                            variant="ghost"
                            className="text-muted hover:text-danger"
                            isDisabled={deleteTag.isPending}
                            onPress={() => deleteTag.mutate(tag.id)}
                        >
                            <TrashBin className="size-4" />
                        </Button>
                    </div>
                ))}
            </div>
        </div>
    );
}
