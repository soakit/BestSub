import { type ComponentProps, useState } from "react";
import { Autocomplete, EmptyState, Label, ListBox, SearchField, Tag, TagGroup, useFilter, Button } from "@heroui/react";
import { useTags, useCreateTag, useDeleteTag } from "../../api/tags";
import { TrashBin } from "@gravity-ui/icons";
import type { Tag as SubTag } from "../../api/sub";

export function TagSelector({ value, onChange }: { value: SubTag[]; onChange: (tags: SubTag[]) => void }) {
    const { data: allTags = [] } = useTags();
    const createTag = useCreateTag();
    const deleteTag = useDeleteTag();
    const { contains } = useFilter({ sensitivity: "base" });

    const selectedKeys = value.map((t) => String(t.id));

    const handleSelectionChange = (keys: string[]) => {
        const selected = allTags.filter((t) => keys.includes(String(t.id)));
        onChange(selected);
    };

    const onRemoveTags: NonNullable<ComponentProps<typeof TagGroup>["onRemove"]> = (keys) => {
        const newKeys = selectedKeys.filter((key) => !keys.has(key));
        handleSelectionChange(newKeys);
    };

    const [inputValue, setInputValue] = useState("");

    const handleCreate = (name: string) => {
        name = name.trim();
        if (!name) return;
        const existing = allTags.find((t) => t.name === name);
        if (existing) {
            if (!selectedKeys.includes(String(existing.id))) {
                onChange([...value, existing]);
            }
            setInputValue("");
            return;
        }
        createTag.mutate(name, {
            onSuccess: (newTag) => {
                onChange([...value, newTag]);
                setInputValue("");
            },
        });
    };

    return (
        <Autocomplete
            allowsEmptyCollection
            className="w-full"
            placeholder="选择或输入新标签"
            selectionMode="multiple"
            value={selectedKeys}
            variant="secondary"
            onChange={(keys) => handleSelectionChange(keys as string[])}
        >
            <Label>标签</Label>
            <Autocomplete.Trigger>
                <Autocomplete.Value>
                    {({ defaultChildren, isPlaceholder, state }: any) => {
                        if (isPlaceholder || state.selectedItems.length === 0) {
                            return defaultChildren;
                        }

                        const selectedItemsKeys = state.selectedItems.map((item: any) => item.key);

                        return (
                            <TagGroup size="sm" onRemove={onRemoveTags}>
                                <TagGroup.List>
                                    {selectedItemsKeys.map((selectedItemKey: string) => {
                                        const item = allTags.find((s) => String(s.id) === selectedItemKey);
                                        if (!item) return null;
                                        return (
                                            <Tag key={item.id} id={String(item.id)}>
                                                {item.name}
                                            </Tag>
                                        );
                                    })}
                                </TagGroup.List>
                            </TagGroup>
                        );
                    }}
                </Autocomplete.Value>
                <Autocomplete.ClearButton />
                <Autocomplete.Indicator />
            </Autocomplete.Trigger>
            <Autocomplete.Popover>
                <Autocomplete.Filter filter={contains} inputValue={inputValue} onInputChange={setInputValue}>
                    <SearchField autoFocus name="search" variant="secondary">
                        <SearchField.Group>
                            <SearchField.SearchIcon />
                            <SearchField.Input
                                placeholder="搜索或输入新标签…"
                                onKeyDown={(e) => {
                                    if (e.key === "Enter") {
                                        e.preventDefault();
                                        handleCreate(inputValue);
                                    }
                                }}
                            />
                            <SearchField.ClearButton />
                        </SearchField.Group>
                    </SearchField>
                    <ListBox renderEmptyState={() => <EmptyState>按 Enter 创建新标签</EmptyState>}>
                        {allTags.map((tag) => (
                            <ListBox.Item key={tag.id} id={String(tag.id)} textValue={tag.name}>
                                <div className="flex w-full items-center justify-between">
                                    <span>{tag.name}</span>
                                    <Button
                                        isIconOnly
                                        size="sm"
                                        variant="ghost"
                                        className="text-muted hover:text-danger"
                                        onPress={() => {
                                            deleteTag.mutate(tag.id);
                                            // 如果当前已选中该标签，也从选中列表中移除
                                            if (selectedKeys.includes(String(tag.id))) {
                                                onChange(value.filter((t) => t.id !== tag.id));
                                            }
                                        }}
                                    >
                                        <TrashBin className="size-4" />
                                    </Button>
                                </div>
                            </ListBox.Item>
                        ))}
                    </ListBox>
                </Autocomplete.Filter>
            </Autocomplete.Popover>
        </Autocomplete>
    );
}
