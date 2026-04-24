"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { getRole, isLoggedIn } from "@/lib/auth";

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    if (!isLoggedIn()) {
      router.push("/login");
    } else {
      const role = getRole();
      router.push(role === "admin" ? "/admin/quizzes" : "/quizzes");
    }
  }, []);

  return <div style={{ padding: 40 }}>Загрузка...</div>;
}