package handlers

import (
	"fmt"
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
	"slices"
	"strconv"
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
	subject := "komp_mreji"

	gotSubject := r.URL.Query().Get("subject")
	if gotSubject != "" {
		subject = gotSubject
	}

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

	c := templates.GetQuestion(subjectList, subject, *qestions)
	err = templates.Layout(c, "My website").Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

// TODO this houdent be hire probably
func (h *getSubjectQuestion) HandleSubmitQuestion(w http.ResponseWriter, r *http.Request) {
	DEBUG := true
	subject := r.FormValue("subject")
	userAnswer := r.FormValue("userAnswer")
	nQuestion, err := strconv.Atoi(r.FormValue("Nquestion"))
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}

	if DEBUG {
		fmt.Println("User selected:", userAnswer)
		fmt.Println("subject:", subject)
		fmt.Println("Nquestion :", nQuestion)
	}

	result := "<lable class=\"result wrong\"> Wrong Answer </lable>"
	answers, err := h.qestionstore.GetCorrectAnswers(subject, nQuestion)
	if err != nil {
		http.Error(w, "Error Getting the answer", http.StatusInternalServerError)
	}

	if slices.Contains(answers, userAnswer) {
		result = "<lable class=\"result corect\"> Corect Answer </lable>"
	}

	fmt.Println("Nquestion :", answers)
	w.Write([]byte(result))
}
