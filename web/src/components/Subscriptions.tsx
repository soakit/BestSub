import {
  Button, Modal, Input, Label, Table, Form, TextField, useOverlayState
} from "@heroui/react";
import { PageLayout } from "./PageLayout";
import { Plus } from "@gravity-ui/icons";

export default function Subscriptions() {
  const modalState = useOverlayState({ defaultOpen: false });

  return (
    <PageLayout
      title="订阅管理"
      actions={
        <Button isIconOnly variant="ghost" onPress={modalState.open} className="rounded-xl">
          <Plus className="size-5" />
        </Button>
      }
    >

      <div className="bg-surface rounded-4xl p-2 flex flex-col flex-1 overflow-hidden">
        <Table className="w-full"><Table.ScrollContainer><Table.Content aria-label="" className=" min-w-full">
          <Table.Header>
            <Table.Column isRowHeader>订阅名称</Table.Column>
            <Table.Column>更新时间</Table.Column>
            <Table.Column>节点数</Table.Column>
            <Table.Column className="text-right">操作</Table.Column>
          </Table.Header>
          <Table.Body>
            <Table.Row id="1" className="hover:bg-surface-secondary transition-colors group">
              <Table.Cell className="font-medium text-foreground">默认提供商</Table.Cell>
              <Table.Cell className="text-muted">2023-11-24 14:20</Table.Cell>
              <Table.Cell className="text-foreground">240</Table.Cell>
              <Table.Cell>
                <button className="text-muted hover:text-focus text-sm cursor-pointer transition-colors">编辑</button>
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
                <Modal.Heading>添加订阅</Modal.Heading>
              </Modal.Header>
              <Modal.Body>
                <Form id="modal-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={(e) => { e.preventDefault(); modalState.close(); }}>
                  <TextField>
                    <Label>订阅名称</Label>
                    <Input placeholder="在此输入订阅的备注名称" variant="secondary" />
                  </TextField>
                  <TextField>
                    <Label>订阅链接</Label>
                    <Input placeholder="https://" variant="secondary" />
                  </TextField>
                </Form>
              </Modal.Body>
              <Modal.Footer>
                <Button variant="ghost" onPress={modalState.close}>
                  取消
                </Button>
                <Button type="submit" form="modal-form" variant="primary">
                  保存
                </Button>
              </Modal.Footer>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>
    </PageLayout>
  );
}
