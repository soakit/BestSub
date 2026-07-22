import { useState, type SyntheticEvent } from "react";
import { useMutation } from "@tanstack/react-query";
import { Button, FieldError, Form, Input, Label, Modal, TextField, useOverlayState } from "@heroui/react";
import { Key, Pencil, Person } from "@gravity-ui/icons";
import { changePassword, changeUsername } from "../../api/auth";

type CredentialMode = "username" | "password"; // 当前弹窗修改的凭据类型。

export function Account() {
    const editorState = useOverlayState({ defaultOpen: false });
    const [mode, setMode] = useState<CredentialMode>("username"); // 当前弹窗的表单模式。
    const [username, setUsername] = useState("");
    const [oldPassword, setOldPassword] = useState("");
    const [newPassword, setNewPassword] = useState("");
    const [confirmPassword, setConfirmPassword] = useState("");
    const reloadAfterCredentialChange = () => window.location.reload(); // 凭据会改变鉴权签名，刷新后重新登录。
    const usernameMutation = useMutation({ mutationFn: changeUsername, onSuccess: reloadAfterCredentialChange });
    const passwordMutation = useMutation({ mutationFn: changePassword, onSuccess: reloadAfterCredentialChange });
    const currentMutation = mode === "username" ? usernameMutation : passwordMutation; // 当前弹窗对应的请求状态。
    const submitDisabled = currentMutation.isPending || (mode === "username"
        ? !username.trim()
        : !oldPassword || !newPassword || newPassword !== confirmPassword);

    // 每次打开都清空另一种凭据及旧错误，避免在两个表单之间残留敏感输入。
    const openEditor = (nextMode: CredentialMode) => {
        setMode(nextMode);
        setUsername("");
        setOldPassword("");
        setNewPassword("");
        setConfirmPassword("");
        usernameMutation.reset();
        passwordMutation.reset();
        editorState.open();
    };

    // 根据当前模式提交已有接口，按钮状态会先拦截空值和密码不一致。
    const handleSubmit = (event: SyntheticEvent<HTMLFormElement>) => {
        event.preventDefault();
        if (mode === "username") {
            if (username.trim()) usernameMutation.mutate(username.trim());
            return;
        }
        if (oldPassword && newPassword === confirmPassword) passwordMutation.mutate({ oldPassword, newPassword });
    };

    return (
        <div className="settings-category">
            <div className="text-foreground/85 mt-1 mb-2.5 flex items-center gap-2 px-1 text-base font-semibold tracking-tight">
                <Person className="size-4 shrink-0" />
                <span className="flex-1">账户</span>
            </div>
            <div className="bg-surface grid grid-cols-1 overflow-hidden rounded-xl">
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <span className="text-foreground grow text-sm">用户名</span>
                    <Button isIconOnly size="sm" variant="ghost" aria-label="修改用户名" onPress={() => openEditor("username")}>
                        <Pencil className="size-4" />
                    </Button>
                </div>
                <div className="flex min-h-11 items-center gap-3 px-4 py-2">
                    <span className="text-foreground grow text-sm">密码</span>
                    <Button isIconOnly size="sm" variant="ghost" aria-label="修改密码" onPress={() => openEditor("password")}>
                        <Pencil className="size-4" />
                    </Button>
                </div>
            </div>

            <Modal state={editorState}>
                <Modal.Backdrop>
                    <Modal.Container size="xs">
                        <Modal.Dialog>
                            <Modal.CloseTrigger />
                            <Modal.Header>
                                <Modal.Icon className="bg-default text-foreground">
                                    {mode === "username" ? <Person className="size-5" /> : <Key className="size-5" />}
                                </Modal.Icon>
                                <Modal.Heading>{mode === "username" ? "修改用户名" : "修改密码"}</Modal.Heading>
                            </Modal.Header>
                            <Modal.Body>
                                <Form id="account-form" validationBehavior="native" className="flex w-full flex-col gap-4" onSubmit={handleSubmit}>
                                    {mode === "username" ? (
                                        <TextField isRequired value={username} onChange={setUsername}>
                                            <Label>新用户名</Label>
                                            <Input autoComplete="username" placeholder="输入新用户名" variant="secondary" />
                                        </TextField>
                                    ) : (
                                        <>
                                            <TextField isRequired value={oldPassword} onChange={setOldPassword}>
                                                <Label>当前密码</Label>
                                                <Input autoComplete="current-password" placeholder="输入当前密码" type="password" variant="secondary" />
                                            </TextField>
                                            <TextField isRequired value={newPassword} onChange={setNewPassword}>
                                                <Label>新密码</Label>
                                                <Input autoComplete="new-password" placeholder="输入新密码" type="password" variant="secondary" />
                                            </TextField>
                                            <TextField
                                                isRequired
                                                isInvalid={!!confirmPassword && newPassword !== confirmPassword}
                                                value={confirmPassword}
                                                onChange={setConfirmPassword}
                                            >
                                                <Label>确认密码</Label>
                                                <Input autoComplete="new-password" placeholder="再次输入新密码" type="password" variant="secondary" />
                                                <FieldError>两次输入的密码不一致</FieldError>
                                            </TextField>
                                        </>
                                    )}
                                    {currentMutation.error instanceof Error && (
                                        <div aria-live="polite" className="rounded-xl bg-danger/10 px-3 py-2 text-sm text-danger">
                                            {currentMutation.error.message}
                                        </div>
                                    )}
                                </Form>
                            </Modal.Body>
                            <Modal.Footer>
                                <Button variant="ghost" onPress={editorState.close}>取消</Button>
                                <Button type="submit" form="account-form" isDisabled={submitDisabled}>
                                    {currentMutation.isPending ? "保存中" : "保存"}
                                </Button>
                            </Modal.Footer>
                        </Modal.Dialog>
                    </Modal.Container>
                </Modal.Backdrop>
            </Modal>
        </div>
    );
}
