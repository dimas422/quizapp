package quiz

import (
	"context"
	"errors"

	"encore.dev/beta/auth"
	"encore.dev/storage/sqldb"
)

var db = sqldb.Named("quiz")

// ===== ТИПЫ =====

type Answer struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	IsCorrect  bool   `json:"is_correct,omitempty"` // скрываем для юзера
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

// ===== ADMIN: список всех квизов =====

//encore:api auth method=GET path=/admin/quizzes
func AdminListQuizzes(ctx context.Context) (*QuizListResponse, error) {
	ud := auth.Data().(*UserData)
	if ud.Role != "admin" {
		return nil, errors.New("доступ запрещён")
	}

	rows, err := db.Query(ctx, `
		SELECT q.id, q.title, q.is_published, q.pass_threshold, q.one_attempt, q.show_answers,
		       COUNT(qu.id) as question_count
		FROM quizzes q
		LEFT JOIN questions qu ON qu.quiz_id = q.id
		GROUP BY q.id
		ORDER BY q.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quizzes []QuizListItem
	for rows.Next() {
		var q QuizListItem
		err := rows.Scan(&q.ID, &q.Title, &q.IsPublished, &q.PassThreshold, &q.OneAttempt, &q.ShowAnswers, &q.QuestionCount)
		if err != nil {
			return nil, err
		}
		quizzes = append(quizzes, q)
	}
	if quizzes == nil {
		quizzes = []QuizListItem{}
	}
	return &QuizListResponse{Quizzes: quizzes}, nil
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

	var quizID string
	err := db.QueryRow(ctx, `
		INSERT INTO quizzes (title, is_published, pass_threshold, one_attempt, show_answers, created_by)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		req.Title, req.IsPublished, req.PassThreshold, req.OneAttempt, req.ShowAnswers, ud.UserID,
	).Scan(&quizID)
	if err != nil {
		return nil, err
	}

	for _, q := range req.Questions {
		if len(q.Answers) < 2 {
			return nil, errors.New("минимум 2 варианта ответа")
		}
		var questionID string
		err := db.QueryRow(ctx, `
			INSERT INTO questions (quiz_id, text, order_index) VALUES ($1, $2, $3) RETURNING id`,
			quizID, q.Text, q.OrderIndex,
		).Scan(&questionID)
		if err != nil {
			return nil, err
		}
		for _, a := range q.Answers {
			_, err := db.Exec(ctx, `
				INSERT INTO answers (question_id, text, is_correct, order_index) VALUES ($1, $2, $3, $4)`,
				questionID, a.Text, a.IsCorrect, a.OrderIndex,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	return getQuizByID(ctx, quizID, true)
}

// ===== ADMIN: получить квиз для редактирования =====

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

	// Обновляем квиз
	_, err := db.Exec(ctx, `
		UPDATE quizzes SET title=$1, is_published=$2, pass_threshold=$3, one_attempt=$4, show_answers=$5
		WHERE id=$6`,
		req.Title, req.IsPublished, req.PassThreshold, req.OneAttempt, req.ShowAnswers, id,
	)
	if err != nil {
		return nil, err
	}

	// Шаг 1: находим все попытки этого квиза
	attemptRows, err := db.Query(ctx, `SELECT id FROM attempts WHERE quiz_id=$1`, id)
	if err != nil {
		return nil, err
	}
	var attemptIDs []string
	for attemptRows.Next() {
		var aid string
		attemptRows.Scan(&aid)
		attemptIDs = append(attemptIDs, aid)
	}
	attemptRows.Close()

	// Шаг 2: удаляем attempt_answers
	for _, aid := range attemptIDs {
		db.Exec(ctx, `DELETE FROM attempt_answers WHERE attempt_id=$1`, aid)
	}

	// Шаг 3: удаляем attempts
	db.Exec(ctx, `DELETE FROM attempts WHERE quiz_id=$1`, id)

	// Шаг 4: находим все вопросы
	qrows, err := db.Query(ctx, `SELECT id FROM questions WHERE quiz_id=$1`, id)
	if err != nil {
		return nil, err
	}
	var questionIDs []string
	for qrows.Next() {
		var qid string
		qrows.Scan(&qid)
		questionIDs = append(questionIDs, qid)
	}
	qrows.Close()

	// Шаг 5: удаляем answers
	for _, qid := range questionIDs {
		db.Exec(ctx, `DELETE FROM answers WHERE question_id=$1`, qid)
	}

	// Шаг 6: удаляем questions
	db.Exec(ctx, `DELETE FROM questions WHERE quiz_id=$1`, id)

	// Шаг 7: создаём новые вопросы и ответы
	for _, q := range req.Questions {
		var questionID string
		err := db.QueryRow(ctx, `
			INSERT INTO questions (quiz_id, text, order_index) VALUES ($1, $2, $3) RETURNING id`,
			id, q.Text, q.OrderIndex,
		).Scan(&questionID)
		if err != nil {
			return nil, err
		}
		for _, a := range q.Answers {
			_, err := db.Exec(ctx, `
				INSERT INTO answers (question_id, text, is_correct, order_index) VALUES ($1, $2, $3, $4)`,
				questionID, a.Text, a.IsCorrect, a.OrderIndex,
			)
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

	_, err := db.Query(ctx, `DELETE FROM quizzes WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &MessageResponse{Message: "квиз удалён"}, nil
}

// ===== ADMIN: опубликовать / скрыть =====

type PublishRequest struct {
	IsPublished bool `json:"is_published"`
}

//encore:api auth method=PATCH path=/admin/quizzes/:id/publish
func AdminPublishQuiz(ctx context.Context, id string, req *PublishRequest) (*MessageResponse, error) {
	ud := auth.Data().(*UserData)
	if ud.Role != "admin" {
		return nil, errors.New("доступ запрещён")
	}

	_, err := db.Query(ctx, `UPDATE quizzes SET is_published=$1 WHERE id=$2`, req.IsPublished, id)
	if err != nil {
		return nil, err
	}
	return &MessageResponse{Message: "статус обновлён"}, nil
}

// ===== USER: список опубликованных квизов =====

//encore:api auth method=GET path=/quizzes
func ListQuizzes(ctx context.Context) (*QuizListResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT q.id, q.title, q.is_published, q.pass_threshold, q.one_attempt, q.show_answers,
		       COUNT(qu.id) as question_count
		FROM quizzes q
		LEFT JOIN questions qu ON qu.quiz_id = q.id
		WHERE q.is_published = true
		GROUP BY q.id
		ORDER BY q.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quizzes []QuizListItem
	for rows.Next() {
		var q QuizListItem
		err := rows.Scan(&q.ID, &q.Title, &q.IsPublished, &q.PassThreshold, &q.OneAttempt, &q.ShowAnswers, &q.QuestionCount)
		if err != nil {
			return nil, err
		}
		quizzes = append(quizzes, q)
	}
	if quizzes == nil {
		quizzes = []QuizListItem{}
	}
	return &QuizListResponse{Quizzes: quizzes}, nil
}

// ===== USER: получить квиз для прохождения (без правильных ответов) =====

//encore:api auth method=GET path=/quizzes/:id
func GetQuiz(ctx context.Context, id string) (*QuizResponse, error) {
	return getQuizByID(ctx, id, false)
}

// ===== USER: отправить ответы =====

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

//encore:api auth method=POST path=/quizzes/:id/submit
func SubmitQuiz(ctx context.Context, id string, req *SubmitRequest) (*SubmitResult, error) {
	ud := auth.Data().(*UserData)

	// Проверяем one_attempt
	var oneAttempt, showAnswers bool
	var passThreshold int
	err := db.QueryRow(ctx,
		`SELECT one_attempt, show_answers, pass_threshold FROM quizzes WHERE id=$1`, id,
	).Scan(&oneAttempt, &showAnswers, &passThreshold)
	if err != nil {
		return nil, errors.New("квиз не найден")
	}

	if oneAttempt {
		var count int
		db.QueryRow(ctx,
			`SELECT COUNT(*) FROM attempts WHERE quiz_id=$1 AND user_id=$2`, id, ud.UserID,
		).Scan(&count)
		if count > 0 {
			return nil, errors.New("вы уже проходили этот квиз")
		}
	}

	// Считаем результат
	score := 0
	total := len(req.Answers)
	var details []AnswerDetail

	for _, a := range req.Answers {
		var isCorrect bool
		var questionText, answerText, correctAnswerText string

		db.QueryRow(ctx,
			`SELECT is_correct FROM answers WHERE id=$1`, a.AnswerID,
		).Scan(&isCorrect)

		if isCorrect {
			score++
		}

		if showAnswers {
			db.QueryRow(ctx, `SELECT text FROM questions WHERE id=$1`, a.QuestionID).Scan(&questionText)
			db.QueryRow(ctx, `SELECT text FROM answers WHERE id=$1`, a.AnswerID).Scan(&answerText)
			db.QueryRow(ctx, `SELECT text FROM answers WHERE question_id=$1 AND is_correct=true`, a.QuestionID).Scan(&correctAnswerText)

			details = append(details, AnswerDetail{
				QuestionText:  questionText,
				YourAnswer:    answerText,
				CorrectAnswer: correctAnswerText,
				IsCorrect:     isCorrect,
			})
		}
	}

	// Сохраняем попытку
	var attemptID string
	db.QueryRow(ctx,
		`INSERT INTO attempts (quiz_id, user_id, score, total) VALUES ($1, $2, $3, $4) RETURNING id`,
		id, ud.UserID, score, total,
	).Scan(&attemptID)

	for _, a := range req.Answers {
		db.Query(ctx,
			`INSERT INTO attempt_answers (attempt_id, question_id, answer_id) VALUES ($1, $2, $3)`,
			attemptID, a.QuestionID, a.AnswerID,
		)
	}

	percent := 0
	if total > 0 {
		percent = (score * 100) / total
	}

	return &SubmitResult{
		Score:       score,
		Total:       total,
		Percent:     percent,
		Passed:      percent >= passThreshold,
		ShowAnswers: showAnswers,
		Details:     details,
	}, nil
}

// ===== ВСПОМОГАТЕЛЬНАЯ: получить квиз с вопросами =====

func getQuizByID(ctx context.Context, id string, withCorrect bool) (*QuizResponse, error) {
	var q Quiz
	err := db.QueryRow(ctx,
		`SELECT id, title, is_published, pass_threshold, one_attempt, show_answers, created_by FROM quizzes WHERE id=$1`, id,
	).Scan(&q.ID, &q.Title, &q.IsPublished, &q.PassThreshold, &q.OneAttempt, &q.ShowAnswers, &q.CreatedBy)
	if err != nil {
		return nil, errors.New("квиз не найден")
	}

	rows, err := db.Query(ctx,
		`SELECT id, text, order_index FROM questions WHERE quiz_id=$1 ORDER BY order_index`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var question Question
		rows.Scan(&question.ID, &question.Text, &question.OrderIndex)

		aRows, err := db.Query(ctx,
			`SELECT id, text, is_correct, order_index FROM answers WHERE question_id=$1 ORDER BY order_index`,
			question.ID,
		)
		if err != nil {
			return nil, err
		}

		for aRows.Next() {
			var a Answer
			aRows.Scan(&a.ID, &a.Text, &a.IsCorrect, &a.OrderIndex)
			if !withCorrect {
				a.IsCorrect = false // скрываем правильный ответ для юзера
			}
			question.Answers = append(question.Answers, a)
		}
		aRows.Close()

		q.Questions = append(q.Questions, question)
	}

	return &QuizResponse{Quiz: q}, nil
}
