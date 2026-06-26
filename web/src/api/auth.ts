import { ApiError, apiRequest } from "./client";

export async function checkAuth() {
	try {
		return await apiRequest<boolean>("/api/v1/user/check", { dispatchUnauthorized: false });
	} catch (error) {
		if (error instanceof ApiError && error.status === 401) return false;
		throw error;
	}
}

export async function login(payload: { username: string; password: string; trust: boolean }) {
	return apiRequest<string>("/api/v1/user/login", {
		method: "POST",
		body: payload,
		dispatchUnauthorized: false,
	});
}

export async function logout() {
	return apiRequest<string>("/api/v1/user/logout", {
		method: "POST",
	});
}

export async function changePassword(payload: { oldPassword: string; newPassword: string }) {
	return apiRequest<string>("/api/v1/user/change-password", {
		method: "POST",
		body: {
			old_password: payload.oldPassword,
			new_password: payload.newPassword,
		},
	});
}

export async function changeUsername(newUsername: string) {
	return apiRequest<string>("/api/v1/user/change-username", {
		method: "POST",
		body: { new_username: newUsername },
	});
}
