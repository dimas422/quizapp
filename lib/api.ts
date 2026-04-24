import axios from "axios";

const BASE_URL = "http://127.0.0.1:4000";

// Создаём axios с токеном
const api = axios.create({ baseURL: BASE_URL });

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// ===== AUTH =====
export const register = (email: string, password: string, role: string) =>
  api.post("/auth/register", { email, password, role });

export const login = (email: string, password: string) =>
  api.post("/auth/login", { email, password });

// ===== ADMIN =====
export const adminGetQuizzes = () => api.get("/admin/quizzes");
export const adminGetQuiz = (id: string) => api.get(`/admin/quizzes/${id}`);
export const adminCreateQuiz = (data: any) => api.post("/admin/quizzes", data);
export const adminUpdateQuiz = (id: string, data: any) => api.put(`/admin/quizzes/${id}`, data);
export const adminDeleteQuiz = (id: string) => api.delete(`/admin/quizzes/${id}`);
export const adminPublishQuiz = (id: string, isPublished: boolean) =>
  api.patch(`/admin/quizzes/${id}/publish`, { is_published: isPublished });

// ===== USER =====
export const getQuizzes = () => api.get("/quizzes");
export const getQuiz = (id: string) => api.get(`/quizzes/${id}`);
export const submitQuiz = (id: string, answers: { question_id: string; answer_id: string }[]) =>
  api.post(`/quizzes/${id}/submit`, { answers });