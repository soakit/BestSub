import { Button, Switch } from "@heroui/react";
import { PageLayout } from "./PageLayout";
import { Plus } from "@gravity-ui/icons";

export default function Settings() {
  return (
    <PageLayout
      title="设置"
      actions={
        <Button isIconOnly variant="ghost" className="rounded-xl">
          <Plus className="size-5" />
        </Button>
      }
    >
      <div className="bg-surface rounded-4xl p-2 flex flex-col flex-1 overflow-hidden">
      </div>
    </PageLayout>
  );
}
