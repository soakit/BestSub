import {
  Button, Chip, Modal, Input, Label, Table, Form, TextField, useOverlayState
} from "@heroui/react";
import { PageLayout } from "./PageLayout";
import { Plus } from "@gravity-ui/icons";

export default function Nodes() {
  const modalState = useOverlayState({ defaultOpen: false });

  return (
    <PageLayout
      title="节点列表"
      actions={
        <Button isIconOnly variant="ghost" onPress={modalState.open} className="rounded-xl">
          <Plus className="size-5" />
        </Button>
      }
    >

      <div className="bg-surface rounded-4xl p-2 flex flex-col flex-1 overflow-hidden">
        <Table className="w-full"><Table.ScrollContainer><Table.Content aria-label="" className=" min-w-full">
          <Table.Header>
            <Table.Column isRowHeader>节点名称</Table.Column>
            <Table.Column>协议</Table.Column>
            <Table.Column>地址</Table.Column>
            <Table.Column className="text-right">操作</Table.Column>
          </Table.Header>
          <Table.Body>
            <Table.Row id="1" className="hover:bg-surface-secondary transition-colors group">
              <Table.Cell className="font-medium text-foreground">香港 Premium 01</Table.Cell>
              <Table.Cell>
                <Chip size="sm" variant="soft" className="bg-surface-secondary text-foreground text-xs">vless</Chip>
              </Table.Cell>
              <Table.Cell className="text-muted font-mono text-sm">hk01.example.com</Table.Cell>
              <Table.Cell>
                <button className="text-muted hover:text-danger text-sm cursor-pointer transition-colors">删除</button>
              </Table.Cell>
            </Table.Row>
            <Table.Row id="2" className="hover:bg-surface-secondary transition-colors group">
              <Table.Cell className="font-medium text-foreground">日本 BGP</Table.Cell>
              <Table.Cell>
                <Chip size="sm" variant="soft" className="bg-surface-secondary text-foreground text-xs">trojan</Chip>
              </Table.Cell>
              <Table.Cell className="text-muted font-mono text-sm">jp.example.com</Table.Cell>
              <Table.Cell>
                <button className="text-muted hover:text-danger text-sm cursor-pointer transition-colors">删除</button>
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
                <Modal.Heading>添加节点</Modal.Heading>
              </Modal.Header>
              <Modal.Body>
                <Form id="modal-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={(e) => { e.preventDefault(); modalState.close(); }}>
                  <TextField>
                    <Label>节点分享链接</Label>
                    <Input placeholder="vmess://..." variant="secondary" />
                  </TextField>
                </Form>
              </Modal.Body>
              <Modal.Footer>
                <Button variant="ghost" onPress={modalState.close}>
                  取消
                </Button>
                <Button type="submit" form="modal-form" variant="primary">
                  添加
                </Button>
              </Modal.Footer>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>
    </PageLayout>
  );
}
