package jsonstore

import (
	"encoding/json"
	"fmt"
	"goth/internal/store"
	"os"
	"path/filepath"
)

type QuestionStore struct {
	entry       map[string]store.SubjectQuestions
	subjectsDir string
}

type NewQuestionStoreParams struct {
	SubjectsDir string // Path to subjects directory
}

func NewQuestionStore(params NewQuestionStoreParams) *QuestionStore {
	if params.SubjectsDir == "" {
		params.SubjectsDir = "assets/subjects/" // Default path
	}

	qs := &QuestionStore{
		entry:       make(map[string]store.SubjectQuestions),
		subjectsDir: params.SubjectsDir,
	}

	if err := qs.loadAllSubjects(); err != nil {
		// Log error but don't fail - store will be empty but functional
		fmt.Printf("Failed to load subjects: %v\n", err)
	}

	return qs
}

func (s *QuestionStore) loadAllSubjects() error {
	// List all subject directories
	subjects, err := s.listSubjects()
	if err != nil {
		return fmt.Errorf("failed to list subjects: %w", err)
	}

	// Load questions for each subject
	for _, subject := range subjects {
		questions, err := s.loadSubjectQuestions(subject)
		if err != nil {
			return fmt.Errorf("failed to load questions for subject %s: %w", subject, err)
		}
		s.entry[subject] = *questions
	}

	return nil
}

func (s *QuestionStore) listSubjects() ([]string, error) {
	var subjects []string
	entries, err := os.ReadDir(s.subjectsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read subjects directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subjects = append(subjects, entry.Name())
		}
	}
	return subjects, nil
}

func (s *QuestionStore) loadSubjectQuestions(subject string) (*store.SubjectQuestions, error) {
	questionsFile := filepath.Join(s.subjectsDir, subject, "questions.json")
	openQuestionsFile := filepath.Join(s.subjectsDir, subject, "open_questions.json")

	var multipleChoice []store.Question
	var openEnded []store.OpenQuestion

	// Load multiple-choice questions
	if err := s.loadJSONFile(questionsFile, &multipleChoice); err != nil {
		return nil, fmt.Errorf("error loading multiple-choice questions: %w", err)
	}

	// Auto-generate IDs
	for i := range multipleChoice {
		multipleChoice[i].ID = i + 1
	}

	// Load open-ended questions
	if err := s.loadJSONFile(openQuestionsFile, &openEnded); err != nil {
		return nil, fmt.Errorf("error loading open-ended questions: %w", err)
	}

	// Auto-generate IDs
	for i := range openEnded {
		openEnded[i].ID = i + 1
	}

	return &store.SubjectQuestions{
		MultipleChoice: multipleChoice,
		OpenEnded:      openEnded,
	}, nil
}

func (s *QuestionStore) loadJSONFile(filePath string, target interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// GetSubjectQuestions returns questions for a specific subject
func (s *QuestionStore) GetSubjectQuestions(subject string) (*store.SubjectQuestions, bool) {
	questions, exists := s.entry[subject]
	if !exists {
		return nil, false
	}
	return &questions, true
}

// GetAllSubjects returns a list of all available subjects
func (s *QuestionStore) GetAllSubjects() ([]string, error) {
	subjects := make([]string, 0, len(s.entry))
	for subject := range s.entry {
		subjects = append(subjects, subject)
	}
	return subjects, nil
}
