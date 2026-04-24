"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { adminCreateQuiz } from "@/lib/api";

interface Answer { text: string; is_correct: boolean; }
interface Question { text: string; answers: Answer[]; }

const emptyQuestion = (): Question => ({
  text: "",
  answers: [
    { text: "", is_correct: true },
    { text: "", is_correct: false },
  ],
});

export default function NewQuizPage() {
  const router = useRouter();
  const [title, setTitle] = useState("");
  const [isPublished, setIsPublished] = useState(false);
  const [passThreshold, setPassThreshold] = useState(0);
  const [oneAttempt, setOneAttempt] = useState(false);
  const [showAnswers, setShowAnswers] = useState(false);
  const [questions, setQuestions] = useState<Question[]>([emptyQuestion()]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const addQuestion = () => setQuestions([...questions, emptyQuestion()]);

  const removeQuestion = (qi: number) =>
    setQuestions(questions.filter((_, i) => i !== qi));

  const updateQuestion = (qi: number, text: string) =>
    setQuestions(questions.map((q, i) => i === qi ? { ...q, text } : q));

  const addAnswer = (qi: number) =>
    setQuestions(questions.map((q, i) =>
      i === qi ? { ...q, answers: [...q.answers, { text: "", is_correct: false }] } : q
    ));

  const removeAnswer = (qi: number, ai: number) =>
    setQuestions(questions.map((q, i) =>
      i === qi ? { ...q, answers: q.answers.filter((_, j) => j !== ai) } : q
    ));

  const updateAnswer = (qi: number, ai: number, text: string) =>
    setQuestions(questions.map((q, i) =>
      i === qi ? { ...q, answers: q.answers.map((a, j) => j === ai ? { ...a, text } : a) } : q
    ));

  const setCorrect = (qi: number, ai: number) =>
    setQuestions(questions.map((q, i) =>
      i === qi ? { ...q, answers: q.answers.map((a, j) => ({ ...a, is_correct: j === ai })) } : q
    ));

  const handleSave = async () => {
    setError("");
    if (!title.trim()) { setError("Введите название"); return; }
    for (let i = 0; i < questions.length; i++) {
      if (!questions[i].text.trim()) { setError(`Заполните текст вопроса ${i + 1}`); return; }
      if (questions[i].answers.length < 2) { setError(`Вопрос ${i + 1}: минимум 2 варианта`); return; }
      if (!questions[i].answers.some((a) => a.is_correct)) { setError(`Вопрос ${i + 1}: выберите правильный ответ`); return; }
    }

    setLoading(true);
    try {
      await adminCreateQuiz({
        title,
        is_published: isPublished,
        pass_threshold: passThreshold,
        one_attempt: oneAttempt,
        show_answers: showAnswers,
        questions: questions.map((q, qi) => ({
          text: q.text,
          order_index: qi,
          answers: q.answers.map((a, ai) => ({ text: a.text, is_correct: a.is_correct, order_index: ai })),
        })),
      });
      router.push("/admin/quizzes");
    } catch (e: any) {
      setError(e.response?.data?.message || "Ошибка сохранения");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container" style={{ paddingTop: 40, paddingBottom: 60 }}>
      <div style={{ display: "flex", alignItems: "center", gap: 16, marginBottom: 32 }}>
        <button className="btn btn-gray" onClick={() => router.push("/admin/quizzes")}>← Назад</button>
        <h1 style={{ fontSize: 24, fontWeight: 700 }}>Новый квиз</h1>
      </div>

      <div className="card" style={{ marginBottom: 24 }}>
        <h2 style={{ fontSize: 18, fontWeight: 600, marginBottom: 16 }}>Настройки</h2>
        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <input className="input" placeholder="Название квиза *" value={title} onChange={(e) => setTitle(e.target.value)} />

          <div style={{ display: "flex", gap: 24, flexWrap: "wrap" }}>
            <label style={{ display: "flex", alignItems: "center", gap: 8, cursor: "pointer" }}>
              <input type="checkbox" checked={isPublished} onChange={(e) => setIsPublished(e.target.checked)} />
              Опубликовать сразу
            </label>
            <label style={{ display: "flex", alignItems: "center", gap: 8, cursor: "pointer" }}>
              <input type="checkbox" checked={oneAttempt} onChange={(e) => setOneAttempt(e.target.checked)} />
              Одна попытка
            </label>
            <label style={{ display: "flex", alignItems: "center", gap: 8, cursor: "pointer" }}>
              <input type="checkbox" checked={showAnswers} onChange={(e) => setShowAnswers(e.target.checked)} />
              Показать ответы после
            </label>
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
            <label style={{ whiteSpace: "nowrap" }}>Порог прохождения %:</label>
            <input
              className="input"
              type="number"
              min={0}
              max={100}
              value={passThreshold}
              onChange={(e) => setPassThreshold(Number(e.target.value))}
              style={{ width: 100 }}
            />
          </div>
        </div>
      </div>

      {/* Вопросы */}
      {questions.map((q, qi) => (
        <div key={qi} className="card" style={{ marginBottom: 16 }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
            <h3 style={{ fontWeight: 600 }}>Вопрос {qi + 1}</h3>
            {questions.length > 1 && (
              <button className="btn btn-danger" style={{ padding: "6px 12px" }} onClick={() => removeQuestion(qi)}>✕</button>
            )}
          </div>

          <input
            className="input"
            placeholder="Текст вопроса *"
            value={q.text}
            onChange={(e) => updateQuestion(qi, e.target.value)}
            style={{ marginBottom: 12 }}
          />

          {q.answers.map((a, ai) => (
            <div key={ai} style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 8 }}>
              <input
                type="radio"
                name={`correct-${qi}`}
                checked={a.is_correct}
                onChange={() => setCorrect(qi, ai)}
                title="Правильный ответ"
              />
              <input
                className="input"
                placeholder={`Вариант ${ai + 1}`}
                value={a.text}
                onChange={(e) => updateAnswer(qi, ai, e.target.value)}
              />
              {q.answers.length > 2 && (
                <button className="btn btn-danger" style={{ padding: "6px 10px", whiteSpace: "nowrap" }} onClick={() => removeAnswer(qi, ai)}>✕</button>
              )}
            </div>
          ))}

          <button className="btn btn-gray" style={{ marginTop: 8 }} onClick={() => addAnswer(qi)}>
            + Добавить вариант
          </button>
        </div>
      ))}

      <button className="btn btn-gray" style={{ marginBottom: 24, width: "100%" }} onClick={addQuestion}>
        + Добавить вопрос
      </button>

      {error && <div className="error" style={{ marginBottom: 16 }}>{error}</div>}

      <button className="btn btn-primary" style={{ width: "100%", padding: 14 }} onClick={handleSave} disabled={loading}>
        {loading ? "Сохранение..." : "Сохранить квиз"}
      </button>
    </div>
  );
}