import { QueryClient } from "@tanstack/react-query";

export const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			retry: 1,
			staleTime: 60 * 1000,
			refetchOnReconnect: 'always',
			refetchOnWindowFocus: 'always',
		},
	},
});

type RequestOptions = {
	method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
	body?: unknown;
	dispatchUnauthorized?: boolean;
	signal?: AbortSignal;
};

export class ApiError extends Error {
	status: number;

	constructor(status: number, message: string) {
		super(message);
		this.name = "ApiError";
		this.status = status;
	}
}

export const apiUnauthorizedEvent = "api:unauthorized";

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
	const response = await fetch(`${import.meta.env.VITE_API_BASE_URL || ""}${path}`, {
		method: options.method || "GET",
		headers: options.body === undefined ? undefined : { "content-type": "application/json" },
		body: options.body === undefined ? undefined : JSON.stringify(options.body),
		credentials: "include",
		signal: options.signal,
	});
	const data = await response.json().catch(() => null) as { code: number; message: string; data?: T } | null;
	if (!response.ok) {
		// 受保护接口返回 401 时广播事件，调用方可统一回到登录态。
		if (response.status === 401 && options.dispatchUnauthorized !== false && typeof window !== "undefined") {
			window.dispatchEvent(new Event(apiUnauthorizedEvent));
		}
		throw new ApiError(response.status, data?.message || `Request failed: ${response.status}`);
	}
	return data?.data as T;
}
