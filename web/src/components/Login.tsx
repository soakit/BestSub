import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Button, Checkbox, Form, Input, Label, TextField } from "@heroui/react";
import { login } from "../api/auth";

export default function Login({ onSuccess }: { onSuccess: () => void }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [trust, setTrust] = useState(false);
  const loginMutation = useMutation({
    mutationFn: login,
    onSuccess,
  });

  return (
    <div className="min-h-screen w-full bg-background text-foreground font-sans flex items-center justify-center p-6">
      <div className="w-full max-w-sm bg-surface rounded-4xl p-6 flex flex-col gap-6 shadow-sm">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">登录</h1>
        </div>

        <Form
          validationBehavior="native"
          className="flex w-full flex-col gap-4"
          onSubmit={(event) => {
            event.preventDefault();
            loginMutation.mutate({ username, password, trust });
          }}
        >
          <TextField>
            <Label>用户名</Label>
            <Input
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="username"
              placeholder="输入用户名"
              required
              variant="secondary"
            />
          </TextField>

          <TextField>
            <Label>密码</Label>
            <Input
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
              placeholder="输入密码"
              required
              type="password"
              variant="secondary"
            />
          </TextField>

          <Checkbox isSelected={trust} onChange={setTrust}>
            <Checkbox.Control>
              <Checkbox.Indicator />
            </Checkbox.Control>
            <Checkbox.Content>
              <Label className="font-medium text-foreground">信任设备</Label>
            </Checkbox.Content>
          </Checkbox>

          {loginMutation.error instanceof Error && (
            <div className="w-full text-sm text-danger bg-danger/10 rounded-2xl px-4 py-3">
              {loginMutation.error.message}
            </div>
          )}

          <Button
            type="submit"
            variant="primary"
            className="w-full"
            isDisabled={loginMutation.isPending || !username || !password}
          >
            {loginMutation.isPending ? "登录中" : "登录"}
          </Button>
        </Form>
      </div>
    </div>
  );
}
