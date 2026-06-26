import {
  Button, Chip, Modal, Input, Label,
  Table, Select, ListBox, Form, TextField, useOverlayState
} from "@heroui/react";
import { PageLayout } from "./PageLayout";
import { Plus } from "@gravity-ui/icons";

export default function Testing() {
  const modalState = useOverlayState({ defaultOpen: false });

  return (
    <PageLayout
      title="检测任务"
      actions={
        <Button isIconOnly variant="ghost" onPress={modalState.open} className="rounded-xl">
          <Plus className="size-5" />
        </Button>
      }
    >

      <div className="bg-surface rounded-4xl p-2 flex flex-col flex-1 overflow-hidden">
        <Table className="w-full"><Table.ScrollContainer><Table.Content aria-label="" className=" min-w-full">
          <Table.Header>
            <Table.Column isRowHeader>任务名称</Table.Column>
            <Table.Column>包含订阅</Table.Column>
            <Table.Column>检测项</Table.Column>
            <Table.Column>状态</Table.Column>
            <Table.Column className="text-right">操作</Table.Column>
          </Table.Header>
          <Table.Body>
            <Table.Row id="1" className="hover:bg-surface-secondary transition-colors group">
              <Table.Cell className="font-medium text-foreground">公益节点测活</Table.Cell>
              <Table.Cell className="text-muted">默认提供商、测试源</Table.Cell>
              <Table.Cell>
                <div className="flex gap-2">
                  <Chip size="sm" variant="soft" className="bg-surface-secondary text-foreground text-xs">测活</Chip>
                  <Chip size="sm" variant="soft" className="bg-surface-secondary text-foreground text-xs">测速</Chip>
                </div>
              </Table.Cell>
              <Table.Cell>
                <span className="text-success font-medium">空闲中</span>
              </Table.Cell>
              <Table.Cell>
                <button className="text-muted hover:text-focus text-sm cursor-pointer transition-colors">执行</button>
              </Table.Cell>
            </Table.Row>
            <Table.Row id="2" className="hover:bg-surface-secondary transition-colors group">
              <Table.Cell className="font-medium text-foreground">解锁检测</Table.Cell>
              <Table.Cell className="text-muted">所有订阅</Table.Cell>
              <Table.Cell>
                <div className="flex gap-2">
                  <Chip size="sm" variant="soft" className="bg-surface-secondary text-foreground text-xs">节点国家</Chip>
                  <Chip size="sm" variant="soft" className="bg-surface-secondary text-foreground text-xs">流媒体</Chip>
                </div>
              </Table.Cell>
              <Table.Cell>
                <span className="text-warning font-medium">检测中 (45%)</span>
              </Table.Cell>
              <Table.Cell>
                <button className="text-muted hover:text-danger text-sm cursor-pointer transition-colors">停止</button>
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
                <Modal.Heading>创建检测任务</Modal.Heading>
              </Modal.Header>
              <Modal.Body>
                <Form id="modal-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={(e) => { e.preventDefault(); modalState.close(); }}>
                  <TextField>
                    <Label>任务名称</Label>
                    <Input placeholder="在此输入任务的备注名称" variant="secondary" />
                  </TextField>

                  <Select selectionMode="multiple" className="w-full" variant="secondary" placeholder="选择检测项">
                    <Label>检测项</Label>
                    <Select.Trigger>
                      <Select.Value />
                      <Select.Indicator />
                    </Select.Trigger>
                    <Select.Popover>
                      <ListBox selectionMode="multiple">
                        <ListBox.Item id="ping">延迟测活</ListBox.Item>
                        <ListBox.Item id="speed">带宽测速</ListBox.Item>
                        <ListBox.Item id="country">落地国家</ListBox.Item>
                        <ListBox.Item id="media">流媒体解锁</ListBox.Item>
                      </ListBox>
                    </Select.Popover>
                  </Select>

                  <Select selectionMode="multiple" className="w-full" variant="secondary" placeholder="选择目标订阅">
                    <Label>目标订阅</Label>
                    <Select.Trigger>
                      <Select.Value />
                      <Select.Indicator />
                    </Select.Trigger>
                    <Select.Popover>
                      <ListBox selectionMode="multiple">
                        <ListBox.Item id="all">所有订阅</ListBox.Item>
                        <ListBox.Item id="default">默认提供商</ListBox.Item>
                        <ListBox.Item id="test">测试源</ListBox.Item>
                      </ListBox>
                    </Select.Popover>
                  </Select>
                </Form>
              </Modal.Body>
              <Modal.Footer>
                <Button variant="ghost" onPress={modalState.close}>
                  取消
                </Button>
                <Button variant="primary" onPress={modalState.close}>
                  保存任务
                </Button>
              </Modal.Footer>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>
    </PageLayout>
  );
}
