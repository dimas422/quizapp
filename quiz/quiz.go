package quiz

import (
	"context"
	"errors"

	"encore.app/ent"
	entanswer "encore.app/ent/answer"
	entquestion "encore.app/ent/question"
	entquiz "encore.app/ent/quiz"
	"encore.dev/beta/auth"
	"encore.dev/storage/sqldb"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

var quizDB = sqldb.Named("quiz")

var entClient *ent.Client

func init() {
	var err error
	entClient, err = ent.OpenEntClient(quizDB.Stdlib())
	if err != nil {
		panic(err)
	}
}

// ===== ТИПЫ =====

type Answer struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	IsCorrect  bool   `json:"is_correct,omitempty"`
	OrderIndex int    `json:"order_index"`
}

type Question struct {
	ID         string   `json:"id"`
	Text       string   `json:"text"`
	OrderIndex int      `json:"order_index"`
	Answers    []Answer `json:"answers"`
}

type Quiz struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	IsPublished   bool       `json:"is_published"`
	PassThreshold int        `json:"pass_threshold"`
	OneAttempt    bool       `json:"one_attempt"`
	ShowAnswers   bool       `json:"show_answers"`
	CreatedBy     string     `json:"created_by"`
	Questions     []Question `json:"questions,omitempty"`
}

type QuizListItem struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	IsPublished   bool   `json:"is_published"`
	QuestionCount int    `json:"question_count"`
	PassThreshold int    `json:"pass_threshold"`
	OneAttempt    bool   `json:"one_attempt"`
	ShowAnswers   bool   `json:"show_answers"`
}

type CreateQuizRequest struct {
	Title         string           `json:"title"`
	IsPublished   bool             `json:"is_published"`
	PassThreshold int              `json:"pass_threshold"`
	OneAttempt    bool             `json:"one_attempt"`
	ShowAnswers   bool             `json:"show_answers"`
	Questions     []CreateQuestion `json:"questions"`
}

type CreateQuestion struct {
	Text       string         `json:"text"`
	OrderIndex int            `json:"order_index"`
	Answers    []CreateAnswer `json:"answers"`
}

type CreateAnswer struct {
	Text       string `json:"text"`
	IsCorrect  bool   `json:"is_correct"`
	OrderIndex int    `json:"order_index"`
}

type QuizListResponse struct {
	Quizzes []QuizListItem `json:"quizzes"`
}

type QuizResponse struct {
	Quiz Quiz `json:"quiz"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type PublishRequest struct {
	IsPublished bool `json:"is_published"`
}

type SubmitRequest struct {
	Answers []SubmitAnswer `json:"answers"`
}

type SubmitAnswer struct {
	QuestionID string `json:"question_id"`
	AnswerID   string `json:"answer_id"`
}

type SubmitResult struct {
	Score       int            `json:"score"`
	Total       int            `json:"total"`
	Percent     int            `json:"percent"`
	Passed      bool           `json:"passed"`
	ShowAnswers bool           `json:"show_answers"`
	Details     []AnswerDetail `json:"details,omitempty"`
}

type AnswerDetail struct {
	QuestionText  string `json:"question_text"`
	YourAnswer    string `json:"your_answer"`
	CorrectAnswer string `json:"correct_answer"`
	IsCorrect     bool   `json:"is_correct"`
}

// ===== ADMIN: список всех квизов =====

//encore:api auth method=GET path=/admin/quizzes
func AdminListQuizzes(ctx context.Context) (*QuizListResponse, error) {
	ud := auth.Data().(*UserData)
	if ud.Role != "admin" {
		return nil, errors.New("доступ запрещён")
	}

	client := entClient

	quizzes, err := client.Quiz.Query().WithQuestions().All(ctx)
	if err != nil {
		return nil, err
	}

	var result []QuizListItem
	for _, q := range quizzes {
		result = append(result, QuizListItem{
			ID:            q.ID.String(),
			Title:         q.Title,
			IsPublished:   q.IsPublished,
			PassThreshold: q.PassThreshold,
			OneAttempt:    q.OneAttempt,
			ShowAnswers:   q.ShowAnswers,
			QuestionCount: len(q.Edges.Questions),
		})
	}
	if result == nil {
		result = []QuizListItem{}
	}
	return &QuizListResponse{Quizzes: result}, nil
}

// ===== ADMIN: создать квиз =====

//encore:api auth method=POST path=/admin/quizzes
func AdminCreateQuiz(ctx context.Context, req *CreateQuizRequest) (*QuizResponse, error) {
	ud := auth.Data().(*UserData)
	if ud.Role != "admin" {
		return nil, errors.New("доступ запрещён")
	}
	if req.Title == "" {
		return nil, errors.New("название обязательно")
	}
	if len(req.Questions) == 0 {
		return nil, errors.New("минимум 1 вопрос")
	}

	client := entClient

	q, err := client.Quiz.Create().
		SetTitle(req.Title).
		SetIsPublished(req.IsPublished).
		SetPassThreshold(req.PassThreshold).
		SetOneAttempt(req.OneAttempt).
		SetShowAnswers(req.ShowAnswers).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	for _, qReq := range req.Questions {
		if len(qReq.Answers) < 2 {
			return nil, errors.New("минимум 2 варианта ответа")
		}
		question, err := client.Question.Create().
			SetText(qReq.Text).
			SetOrderIndex(qReq.OrderIndex).
			SetQuiz(q).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		for _, aReq := range qReq.Answers {
			_, err := client.Answer.Create().
				SetText(aReq.Text).
				SetIsCorrect(aReq.IsCorrect).
				SetOrderIndex(aReq.OrderIndex).
				SetQuestion(question).
				Save(ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	return getQuizByID(ctx, q.ID.String(), true)
}

// ===== ADMIN: получить квиз =====

//encore:api auth method=GET path=/admin/quizzes/:id
func AdminGetQuiz(ctx context.Context, id string) (*QuizResponse, error) {
	ud := auth.Data().(*UserData)
	if ud.Role != "admin" {
		return nil, errors.New("доступ запрещён")
	}
	return getQuizByID(ctx, id, true)
}

// ===== ADMIN: обновить квиз =====

//encore:api auth method=PUT path=/admin/quizzes/:id
func AdminUpdateQuiz(ctx context.Context, id string, req *CreateQuizRequest) (*QuizResponse, error) {
	ud := auth.Data().(*UserData)
	if ud.Role != "admin" {
		return nil, errors.New("доступ запрещён")
	}
	if req.Title == "" {
		return nil, errors.New("название обязательно")
	}

	client := entClient

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("неверный id")
	}

	_, err = client.Quiz.UpdateOneID(uid).
		SetTitle(req.Title).
		SetIsPublished(req.IsPublished).
		SetPassThreshold(req.PassThreshold).
		SetOneAttempt(req.OneAttempt).
		SetShowAnswers(req.ShowAnswers).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// Удаляем старые вопросы
	questions, _ := client.Question.Query().
		Where(entquestion.HasQuizWith(entquiz.ID(uid))).
		WithAnswers().
		All(ctx)

	for _, q := range questions {
		for _, a := range q.Edges.Answers {
			client.Answer.DeleteOneID(a.ID).Exec(ctx)
		}
		client.Question.DeleteOneID(q.ID).Exec(ctx)
	}

	// Создаём новые вопросы
	quizEnt, _ := client.Quiz.Get(ctx, uid)
	for _, qReq := range req.Questions {
		question, err := client.Question.Create().
			SetText(qReq.Text).
			SetOrderIndex(qReq.OrderIndex).
			SetQuiz(quizEnt).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		for _, aReq := range qReq.Answers {
			_, err := client.Answer.Create().
				SetText(aReq.Text).
				SetIsCorrect(aReq.IsCorrect).
				SetOrderIndex(aReq.OrderIndex).
				SetQuestion(question).
				Save(ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	return getQuizByID(ctx, id, true)
}

// ===== ADMIN: удалить квиз =====

//encore:api auth method=DELETE path=/admin/quizzes/:id
func AdminDeleteQuiz(ctx context.Context, id string) (*MessageResponse, error) {
	ud := auth.Data().(*UserData)
	if ud.Role != "admin" {
		return nil, errors.New("доступ запрещён")
	}

	client := entClient

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("неверный id")
	}

	err = client.Quiz.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &MessageResponse{Message: "квиз удалён"}, nil
}

// ===== ADMIN: опубликовать/скрыть =====

//encore:api auth method=PATCH path=/admin/quizzes/:id/publish
func AdminPublishQuiz(ctx context.Context, id string, req *PublishRequest) (*MessageResponse, error) {
	ud := auth.Data().(*UserData)
	if ud.Role != "admin" {
		return nil, errors.New("доступ запрещён")
	}

	client := entClient

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("неверный id")
	}

	_, err = client.Quiz.UpdateOneID(uid).
		SetIsPublished(req.IsPublished).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &MessageResponse{Message: "статус обновлён"}, nil
}

// ===== USER: список опубликованных квизов =====

//encore:api auth method=GET path=/quizzes
func ListQuizzes(ctx context.Context) (*QuizListResponse, error) {
	client := entClient

	quizzes, err := client.Quiz.Query().
		Where(entquiz.IsPublished(true)).
		WithQuestions().
		All(ctx)
	if err != nil {
		return nil, err
	}

	var result []QuizListItem
	for _, q := range quizzes {
		result = append(result, QuizListItem{
			ID:            q.ID.String(),
			Title:         q.Title,
			IsPublished:   q.IsPublished,
			PassThreshold: q.PassThreshold,
			OneAttempt:    q.OneAttempt,
			ShowAnswers:   q.ShowAnswers,
			QuestionCount: len(q.Edges.Questions),
		})
	}
	if result == nil {
		result = []QuizListItem{}
	}
	return &QuizListResponse{Quizzes: result}, nil
}

// ===== USER: получить квиз для прохождения =====

//encore:api auth method=GET path=/quizzes/:id
func GetQuiz(ctx context.Context, id string) (*QuizResponse, error) {
	return getQuizByID(ctx, id, false)
}

// ===== USER: отправить ответы =====

//encore:api auth method=POST path=/quizzes/:id/submit
func SubmitQuiz(ctx context.Context, id string, req *SubmitRequest) (*SubmitResult, error) {
	ud := auth.Data().(*UserData)

	client := entClient

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("неверный id")
	}

	q, err := client.Quiz.Get(ctx, uid)
	if err != nil {
		return nil, errors.New("квиз не найден")
	}

	if q.OneAttempt {
		userUID, _ := uuid.Parse(ud.UserID)
		count, _ := client.Attempt.Query().
			Where(func(s *sql.Selector) {
				s.Where(sql.And(
					sql.EQ("quiz_id", uid),
					sql.EQ("user_id", userUID),
				))
			}).Count(ctx)
		if count > 0 {
			return nil, errors.New("вы уже проходили этот квиз")
		}
	}

	score := 0
	total := len(req.Answers)
	var details []AnswerDetail

	for _, a := range req.Answers {
		answerUID, err := uuid.Parse(a.AnswerID)
		if err != nil {
			continue
		}
		answerEnt, err := client.Answer.Get(ctx, answerUID)
		if err != nil {
			continue
		}
		if answerEnt.IsCorrect {
			score++
		}
		if q.ShowAnswers {
			questionUID, _ := uuid.Parse(a.QuestionID)
			questionEnt, _ := client.Question.Get(ctx, questionUID)
			correctAnswer, _ := client.Answer.Query().
				Where(entanswer.IsCorrect(true), entanswer.HasQuestionWith(entquestion.ID(questionUID))).
				First(ctx)

			detail := AnswerDetail{
				IsCorrect:  answerEnt.IsCorrect,
				YourAnswer: answerEnt.Text,
			}
			if questionEnt != nil {
				detail.QuestionText = questionEnt.Text
			}
			if correctAnswer != nil {
				detail.CorrectAnswer = correctAnswer.Text
			}
			details = append(details, detail)
		}
	}

	userUID, _ := uuid.Parse(ud.UserID)
	attempt, err := client.Attempt.Create().
		SetScore(score).
		SetTotal(total).
		SetQuizID(uid).
		SetUserID(userUID).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	for _, a := range req.Answers {
		questionUID, _ := uuid.Parse(a.QuestionID)
		answerUID, _ := uuid.Parse(a.AnswerID)
		client.AttemptAnswer.Create().
			SetAttemptID(attempt.ID).
			SetQuestionID(questionUID).
			SetAnswerID(answerUID).
			Save(ctx)
	}

	percent := 0
	if total > 0 {
		percent = (score * 100) / total
	}

	return &SubmitResult{
		Score:       score,
		Total:       total,
		Percent:     percent,
		Passed:      percent >= q.PassThreshold,
		ShowAnswers: q.ShowAnswers,
		Details:     details,
	}, nil
}

// ===== ВСПОМОГАТЕЛЬНАЯ =====

func getQuizByID(ctx context.Context, id string, withCorrect bool) (*QuizResponse, error) {
	client := entClient

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("неверный id")
	}

	q, err := client.Quiz.Query().
		Where(entquiz.ID(uid)).
		WithQuestions(func(qq *ent.QuestionQuery) {
			qq.WithAnswers()
		}).
		Only(ctx)
	if err != nil {
		return nil, errors.New("квиз не найден")
	}

	var questions []Question
	for _, question := range q.Edges.Questions {
		var answers []Answer
		for _, a := range question.Edges.Answers {
			answer := Answer{
				ID:         a.ID.String(),
				Text:       a.Text,
				OrderIndex: a.OrderIndex,
			}
			if withCorrect {
				answer.IsCorrect = a.IsCorrect
			}
			answers = append(answers, answer)
		}
		questions = append(questions, Question{
			ID:         question.ID.String(),
			Text:       question.Text,
			OrderIndex: question.OrderIndex,
			Answers:    answers,
		})
	}

	return &QuizResponse{Quiz: Quiz{
		ID:            q.ID.String(),
		Title:         q.Title,
		IsPublished:   q.IsPublished,
		PassThreshold: q.PassThreshold,
		OneAttempt:    q.OneAttempt,
		ShowAnswers:   q.ShowAnswers,
		Questions:     questions,
	}}, nil
}