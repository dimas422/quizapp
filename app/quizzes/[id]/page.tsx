"use client";
import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { getQuiz, submitQuiz } from "@/lib/api";

interface Answer { id: string; text: string; }
interface Question { id: string; text: string; answers: Answer[]; }

export default function QuizPage() {
  const router = useRouter();
  const params = useParams();
  const id = params.id as string;

  const [questions, setQuestions] = useState<Question[]>([]);
  const [title, setTitle] = useState("");
  const [current, setCurrent] = useState(0);
  const [selected, setSelected] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    loadQuiz();
  }, []);

  const loadQuiz = async () => {
    try {
      const res = await getQuiz(id);
      setTitle(res.data.quiz.title);
      setQuestions(res.data.quiz.questions);
    } catch {
      setError("Ошибка загрузки квиза");
    } finally {
      setLoading(false);
    }
  };

  const handleSelect = (questionId: string, answerId: string) => {
    setSelected({ ...selected, [questionId]: answerId });
  };

  const handleSubmit = async () => {
    if (Object.keys(selected).length < questions.length) {
      setError("Ответьте на все вопросы");
      return;
    }
    setSubmitting(true);
    try {
      const answers = Object.entries(selected).map(([question_id, answer_id]) => ({
        question_id,
        answer_id,
      }));
      const res = await submitQuiz(id, answers);
      // Сохраняем результат и переходим
      localStorage.setItem("quiz_result", JSON.stringify(res.data));
      router.push(`/quizzes/${id}/result`);
    } catch (e: any) {
      setError(e.response?.data?.message || "Ошибка отправки");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <div style={{ padding: 40 }}>Загрузка...</div>;
  if (error && questions.length === 0) return (
    <div style={{ padding: 40 }}>
      <div className="error">{error}</div>
      <button className="btn btn-gray" style={{ marginTop: 16 }} onClick={() => router.push("/quizzes")}>← Назад</button>
    </div>
  );

  const question = questions[current];

  return (
    <div className="container" style={{ paddingTop: 40, maxWidth: 700 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
        <button className="btn btn-gray" onClick={() => router.push("/quizzes")}>← Назад</button>
        <span style={{ color: "#888", fontSize: 14 }}>Вопрос {current + 1} из {questions.length}</span>
      </div>

      <h1 style={{ fontSize: 22, fontWeight: 700, marginBottom: 24 }}>{title}</h1>

      {/* Прогресс бар */}
      <div style={{ background: "#e5e7eb", borderRadius: 8, height: 6, marginBottom: 32 }}>
        <div style={{
          background: "#6366f1",
          height: 6,
          borderRadius: 8,
          width: `${((current + 1) / questions.length) * 100}%`,
          transition: "width 0.3s"
        }} />
      </div>

      <div className="card">
        <h2 style={{ fontSize: 18, fontWeight: 600, marginBottom: 20 }}>{question.text}</h2>

        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          {question.answers.map((answer) => {
            const isSelected = selected[question.id] === answer.id;
            return (
              <div
                key={answer.id}
                onClick={() => handleSelect(question.id, answer.id)}
                style={{
                  padding: "14px 18px",
                  border: `2px solid ${isSelected ? "#6366f1" : "#e5e7eb"}`,
                  borderRadius: 10,
                  cursor: "pointer",
                  background: isSelected ? "#eef2ff" : "white",
                  fontWeight: isSelected ? 600 : 400,
                  transition: "all 0.15s",
                }}
              >
                {answer.text}
              </div>
            );
          })}
        </div>
      </div>

      {error && <div className="error" style={{ marginTop: 16 }}>{error}</div>}

      <div style={{ display: "flex", gap: 12, marginTop: 24 }}>
        {current > 0 && (
          <button className="btn btn-gray" style={{ flex: 1, padding: 14 }}
            onClick={() => setCurrent(current - 1)}>
            ← Назад
          </button>
        )}
        {current < questions.length - 1 ? (
          <button
            className="btn btn-primary"
            style={{ flex: 1, padding: 14 }}
            onClick={() => {
              if (!selected[question.id]) { setError("Выберите ответ"); return; }
              setError("");
              setCurrent(current + 1);
            }}
          >
            Следующий →
          </button>
        ) : (
          <button
            className="btn btn-success"
            style={{ flex: 1, padding: 14 }}
            onClick={handleSubmit}
            disabled={submitting}
          >
            {submitting ? "Отправка..." : "Завершить квиз ✓"}
          </button>
        )}
      </div>
    </div>
  );
}