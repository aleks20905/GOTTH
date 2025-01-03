package handlers

import (
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
)

type getSubjectQuestion struct {
	qestionstore store.QuestionStorer
}

type GetgetSubjectQuestionParams struct {
	Qestionstore store.QuestionStorer
}

func NewSubjectQuestion(params GetgetSubjectQuestionParams) *getSubjectQuestion {
	return &getSubjectQuestion{
		qestionstore: params.Qestionstore,
	}
}

func (h *getSubjectQuestion) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	idk, err := h.qestionstore.GetAllSubjects()
	if err != nil {
		http.Error(w, "Error loading schedule", http.StatusInternalServerError)
		return
	}

	c := templates.GetQuestion(idk, store.SubjectQuestions{})
	err = templates.Layout(c, "My website").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
