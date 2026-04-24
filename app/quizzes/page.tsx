"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getQuizzes } from "@/lib/api";
import { getRole, logout } from "@/lib/auth";

interface Quiz {
  id: string;
  title: string;
  question_count: number;
  one_attempt: boolean;
}

export default function QuizzesPage() {
  const router = useRouter();
  const [quizzes, setQuizzes] = useState<Quiz[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (getRole() !== "user") {
      router.push("/login");
      return;
    }
    loadQuizzes();
  }, []);

  const loadQuizzes = async () => {
    try {
      const res = await getQuizzes();
      setQuizzes(res.data.quizzes);
    } catch {
      setError("Ошибка загрузки");
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  return (
    <div className="container" style={{ paddingTop: 40 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 32 }}>
        <h1 style={{ fontSize: 28, fontWeight: 700 }}>Доступные квизы</h1>
        <button className="btn btn-gray" onClick={handleLogout}>Выйти</button>
      </div>

      {error && <div className="error" style={{ marginBottom: 16 }}>{error}</div>}
      {loading && <p>Загрузка...</p>}

      {!loading && quizzes.length === 0 && (
        <div className="card" style={{ textAlign: "center", padding: 60, color: "#888" }}>
          <p style={{ fontSize: 18 }}>Пока нет доступных квизов</p>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        {quizzes.map((quiz) => (
          <div key={quiz.id} className="card"
            style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <div>
              <h3 style={{ fontSize: 18, fontWeight: 600 }}>{quiz.title}</h3>
              <p style={{ color: "#888", marginTop: 4, fontSize: 14 }}>
                {quiz.question_count} вопросов
              </p>
            </div>
            <button
              className="btn btn-primary"
              onClick={() => router.push(`/quizzes/${quiz.id}`)}
            >
              Пройти
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}   