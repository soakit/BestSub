import {
  Button, Switch, Modal, Input, Label, Table, Form, TextField, useOverlayState
} from "@heroui/react";
import { PageLayout } from "./PageLayout";
import { Plus } from "@gravity-ui/icons";

export default function Storage() {
  const modalState = useOverlayState({ defaultOpen: false });

  return (
    <PageLayout
      title="云端储存"
      actions={
        <Button isIconOnly variant="ghost" onPress={modalState.open} className="rounded-xl">
          <Plus className="size-5" />
        </Button>
      }
    >

      <div className="bg-surface rounded-4xl p-2 flex flex-col flex-1 overflow-hidden">
        <Table className="w-full"><Table.ScrollContainer><Table.Content aria-label="" className=" min-w-full">
          <Table.Header>
            <Table.Column isRowHeader>服务类型</Table.Column>
            <Table.Column>目标地址</Table.Column>
            <Table.Column>定时同步</Table.Column>
            <Table.Column className="text-right">操作</Table.Column>
          </Table.Header>
          <Table.Body>
            <Table.Row id="1" className="hover:bg-surface-secondary transition-colors group">
              <Table.Cell className="font-medium text-foreground">Amazon S3</Table.Cell>
              <Table.Cell className="text-muted max-w-[200px] truncate font-mono text-sm">s3://mihomo-backup/nodes.yaml</Table.Cell>
              <Table.Cell>
                <Switch size="sm" defaultSelected />
              </Table.Cell>
              <Table.Cell>
                <div className="flex justify-end gap-2">
                  <button className="text-muted hover:text-focus text-sm cursor-pointer transition-colors">编辑</button>
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
                <Modal.Heading>添加储存</Modal.Heading>
              </Modal.Header>
              <Modal.Body>
                <Form id="modal-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={(e) => { e.preventDefault(); modalState.close(); }}>
                  <TextField>
                    <Label>服务类型</Label>
                    <Input placeholder="例如 Amazon S3, WebDAV 等" variant="secondary" />
                  </TextField>
                  <TextField>
                    <Label>目标地址</Label>
                    <Input placeholder="配置信息、链接或 Bucket 名称" variant="secondary" />
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
