"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

interface Detail {
  question_text: string;
  your_answer: string;
  correct_answer: string;
  is_correct: boolean;
}

interface Result {
  score: number;
  total: number;
  percent: number;
  passed: boolean;
  show_answers: boolean;
  details: Detail[];
}

export default function ResultPage() {
  const router = useRouter();
  const [result, setResult] = useState<Result | null>(null);

  useEffect(() => {
    // Даём время странице загрузиться
    const timer = setTimeout(() => {
      const data = localStorage.getItem("quiz_result");
      if (!data) {
        router.push("/quizzes");
        return;
      }
      const parsed = JSON.parse(data);
      setResult(parsed);
      // НЕ удаляем сразу — удалим когда юзер уйдёт
    }, 100);

    return () => clearTimeout(timer);
  }, []);

  const handleBack = () => {
    localStorage.removeItem("quiz_result");
    router.push("/quizzes");
  };

  if (!result) return (
    <div style={{ padding: 40, textAlign: "center" }}>
      <p>Загрузка результата...</p>
    </div>
  );

  return (
    <div className="container" style={{ paddingTop: 40, maxWidth: 700 }}>
      <h1 style={{ fontSize: 28, fontWeight: 700, marginBottom: 24 }}>Результат</h1>

      {/* Итог */}
      <div className="card" style={{ textAlign: "center", marginBottom: 24, padding: 40 }}>
        <div style={{ fontSize: 64, marginBottom: 8 }}>
          {result.percent >= 80 ? "🎉" : result.percent >= 50 ? "👍" : "😔"}
        </div>
        <h2 style={{ fontSize: 32, fontWeight: 700, color: "#6366f1" }}>
          {result.percent}%
        </h2>
        <p style={{ fontSize: 18, marginTop: 8, color: "#555" }}>
          Вы ответили правильно на <strong>{result.score}</strong> из <strong>{result.total}</strong> вопросов
        </p>
        <div style={{
          marginTop: 16,
          display: "inline-block",
          padding: "8px 20px",
          borderRadius: 20,
          background: result.passed ? "#dcfce7" : "#fee2e2",
          color: result.passed ? "#16a34a" : "#dc2626",
          fontWeight: 700,
          fontSize: 16,
        }}>
          {result.passed ? "✓ Тест пройден!" : "✗ Тест не пройден"}
        </div>
      </div>

      {/* Детали ответов */}
      {result.show_answers && result.details && result.details.length > 0 && (
        <div>
          <h2 style={{ fontSize: 20, fontWeight: 700, marginBottom: 16 }}>Разбор ответов</h2>
          <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
            {result.details.map((d, i) => (
              <div key={i} className="card" style={{
                borderLeft: `4px solid ${d.is_correct ? "#22c55e" : "#ef4444"}`
              }}>
                <p style={{ fontWeight: 600, marginBottom: 8 }}>{d.question_text}</p>
                <p style={{ fontSize: 14, color: d.is_correct ? "#16a34a" : "#dc2626" }}>
                  Ваш ответ: {d.your_answer}
                </p>
                {!d.is_correct && (
                  <p style={{ fontSize: 14, color: "#16a34a", marginTop: 4 }}>
                    Правильный: {d.correct_answer}
                  </p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      <button
        className="btn btn-primary"
        style={{ width: "100%", padding: 14, marginTop: 24 }}
        onClick={handleBack}
      >
        ← Вернуться к квизам
      </button>
    </div>
  );
}