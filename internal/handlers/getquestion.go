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
	subject := "komp_arhitekturi"
	subjectList, err := h.qestionstore.GetAllSubjects()
	if err != nil {
		http.Error(w, "Error getting subject there is no way to see this error !!! no way !!!", http.StatusInternalServerError) // if u see this error just pray
		return
	}

	qestions, err := h.qestionstore.GetSubjectQuestions(subject)
	if err != nil {
		http.Error(w, "Error getting subjectsQestion", http.StatusInternalServerError)
		return
	}

	c := templates.GetQuestion(subjectList, *qestions)
	err = templates.Layout(c, "My website").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
