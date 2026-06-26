import {
  Button, Modal, Input, Label, Table, Form, TextField, useOverlayState
} from "@heroui/react";
import { PageLayout } from "./PageLayout";
import { Plus } from "@gravity-ui/icons";

export default function Sharing() {
  const modalState = useOverlayState({ defaultOpen: false });

  return (
    <PageLayout
      title="分享链接"
      actions={
        <Button isIconOnly variant="ghost" onPress={modalState.open} className="rounded-xl">
          <Plus className="size-5" />
        </Button>
      }
    >

      <div className="bg-surface rounded-4xl p-2 flex flex-col flex-1 overflow-hidden">
        <Table className="w-full"><Table.ScrollContainer><Table.Content aria-label="" className=" min-w-full">
          <Table.Header>
            <Table.Column isRowHeader>标识</Table.Column>
            <Table.Column>链接</Table.Column>
            <Table.Column>包含节点数</Table.Column>
            <Table.Column className="text-right">操作</Table.Column>
          </Table.Header>
          <Table.Body>
            <Table.Row id="1" className="hover:bg-surface-secondary transition-colors group">
              <Table.Cell className="font-medium text-foreground">default-mix</Table.Cell>
              <Table.Cell className="text-muted font-mono text-sm max-w-[200px] truncate">https://example.com/share/df39a2</Table.Cell>
              <Table.Cell className="text-muted">24个</Table.Cell>
              <Table.Cell>
                <div className="flex justify-end gap-2">
                  <button className="text-muted hover:text-focus text-sm cursor-pointer transition-colors">复制</button>
                  <button className="text-muted hover:text-danger text-sm cursor-pointer transition-colors">删除</button>
                </div>
              </Table.Cell>
            </Table.Row>
          </Table.Body>
        </Table.Content></Table.ScrollContainer></Table>
      </div>

      <Modal state={modalState}>
        <Modal.Backdrop>
          <Modal.Container>
            <Modal.Dialog>
              <Modal.CloseTrigger />
              <Modal.Header>
                <Modal.Heading>创建分享</Modal.Heading>
              </Modal.Header>
              <Modal.Body>
                <Form id="modal-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={(e) => { e.preventDefault(); modalState.close(); }}>
                  <TextField>
                    <Label>分享标识</Label>
                    <Input placeholder="便于记忆的名称，例如 default-mix" variant="secondary" />
                  </TextField>
                </Form>
              </Modal.Body>
              <Modal.Footer>
                <Button variant="ghost" onPress={modalState.close}>
                  取消
                </Button>
                <Button type="submit" form="modal-form" variant="primary">
                  创建
                </Button>
              </Modal.Footer>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>
    </PageLayout>
  );
}
