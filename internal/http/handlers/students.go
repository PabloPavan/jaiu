package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/PabloPavan/jaiu/internal/view"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) StudentsIndex(w http.ResponseWriter, r *http.Request) {
	data := h.buildStudentsData(r)
	if isHTMX(r) {
		h.renderComponent(w, r, view.StudentsList(data))
		return
	}

	h.renderPage(w, r, page("Alunos", view.StudentsPage(data)))
}

func (h *Handler) StudentsPreview(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	data := view.StudentsPreviewData{}

	if h.services.Students != nil {
		filter := ports.StudentFilter{
			Query:    query,
			Statuses: []domain.StudentStatus{domain.StudentActive},
			Limit:    5,
		}
		students, err := h.services.Students.Search(r.Context(), filter)
		if err != nil {
			log.Printf("preview students: %v", err)
		} else {
			data.Items = make([]view.StudentItem, 0, len(students))
			for _, student := range students {
				label, className := statusPresentation(student.Status)
				item := view.StudentItem{
					ID:          student.ID,
					FullName:    student.FullName,
					BirthDate:   formatDateBR(student.BirthDate),
					Phone:       student.Phone,
					Email:       student.Email,
					Status:      string(student.Status),
					StatusLabel: label,
					StatusClass: className,
				}
				data.Items = append(data.Items, item)
			}
		}
	}

	h.renderComponent(w, r, view.StudentsPreview(data))
}

func (h *Handler) StudentsNew(w http.ResponseWriter, r *http.Request) {
	data := studentFormCreateData()
	h.renderPage(w, r, page("Novo aluno", view.StudentFormPage(data)))
}

func (h *Handler) StudentsCreate(w http.ResponseWriter, r *http.Request) {
	data := studentFormCreateData()
	student, err := parseStudentForm(r, &data)
	if err != nil {
		data.Error = err.Error()
		if isHTMX(r) {
			h.renderComponent(w, r, view.StudentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.StudentFormPage(data)))
		return
	}

	if h.services.Students == nil {
		data.Error = "Servico de alunos indisponivel."
		if isHTMX(r) {
			h.renderComponent(w, r, view.StudentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.StudentFormPage(data)))
		return
	}

	if _, err := h.services.Students.Register(r.Context(), student); err != nil {
		data.Error = "Nao foi possivel salvar o aluno."
		if isHTMX(r) {
			h.renderComponent(w, r, view.StudentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.StudentFormPage(data)))
		return
	}

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/students")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

func (h *Handler) StudentsEdit(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "studentID")
	if h.services.Students == nil {
		http.NotFound(w, r)
		return
	}

	student, err := h.services.Students.FindByID(r.Context(), studentID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("load student: %v", err)
		http.Error(w, "Erro ao carregar aluno.", http.StatusInternalServerError)
		return
	}

	data := studentFormEditData(student)
	h.renderPage(w, r, page(data.Title, view.StudentFormPage(data)))
}

func (h *Handler) StudentsUpdate(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "studentID")
	data := studentFormEditData(domain.Student{ID: studentID})
	student, err := parseStudentForm(r, &data)
	if err != nil {
		data.Error = err.Error()
		if isHTMX(r) {
			h.renderComponent(w, r, view.StudentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.StudentFormPage(data)))
		return
	}

	if h.services.Students == nil {
		data.Error = "Servico de alunos indisponivel."
		if isHTMX(r) {
			h.renderComponent(w, r, view.StudentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.StudentFormPage(data)))
		return
	}

	student.ID = studentID
	if _, err := h.services.Students.Update(r.Context(), student); err != nil {
		data.Error = "Nao foi possivel atualizar o aluno."
		if isHTMX(r) {
			h.renderComponent(w, r, view.StudentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.StudentFormPage(data)))
		return
	}

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/students")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

func (h *Handler) StudentsDelete(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "studentID")
	if h.services.Students == nil {
		http.NotFound(w, r)
		return
	}

	if _, err := h.services.Students.Deactivate(r.Context(), studentID); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("deactivate student: %v", err)
		http.Error(w, "Erro ao excluir aluno.", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		data := h.buildStudentsData(r)
		h.renderComponent(w, r, view.StudentsList(data))
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

func studentFormCreateData() view.StudentFormData {
	return view.StudentFormData{
		Title:       "Novo aluno",
		Action:      "/students",
		SubmitLabel: "Criar aluno",
		Status:      string(domain.StudentActive),
	}
}

func studentFormEditData(student domain.Student) view.StudentFormData {
	return view.StudentFormData{
		Title:        "Editar aluno",
		Action:       "/students/" + student.ID,
		SubmitLabel:  "Salvar",
		DeleteAction: "/students/" + student.ID + "/delete",
		ShowDelete:   student.ID != "",
		FullName:     student.FullName,
		BirthDate:    formatDateInputBR(student.BirthDate),
		Gender:       student.Gender,
		Phone:        student.Phone,
		Email:        student.Email,
		CPF:          student.CPF,
		Address:      student.Address,
		Notes:        student.Notes,
		PhotoURL:     student.PhotoURL,
		Status:       string(student.Status),
	}
}

func parseStudentForm(r *http.Request, data *view.StudentFormData) (domain.Student, error) {
	if err := r.ParseForm(); err != nil {
		return domain.Student{}, errors.New("Nao foi possivel ler o formulario.")
	}

	fullName := strings.TrimSpace(r.FormValue("full_name"))
	data.FullName = fullName
	if fullName == "" {
		return domain.Student{}, errors.New("Nome completo e obrigatorio.")
	}

	statusValue := strings.TrimSpace(r.FormValue("status"))
	status, err := parseStudentStatus(statusValue)
	if err != nil {
		return domain.Student{}, err
	}
	data.Status = string(status)

	birthRaw := strings.TrimSpace(r.FormValue("birth_date"))
	data.BirthDate = birthRaw
	birthDate, err := parseDateInput(birthRaw)
	if err != nil {
		return domain.Student{}, errors.New("Data de nascimento invalida. Use dd/mm/aaaa.")
	}

	gender := strings.TrimSpace(r.FormValue("gender"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	email := strings.TrimSpace(r.FormValue("email"))
	cpf := strings.TrimSpace(r.FormValue("cpf"))
	address := strings.TrimSpace(r.FormValue("address"))
	notes := strings.TrimSpace(r.FormValue("notes"))
	photoURL := strings.TrimSpace(r.FormValue("photo_url"))

	data.Gender = gender
	data.Phone = phone
	data.Email = email
	data.CPF = cpf
	data.Address = address
	data.Notes = notes
	data.PhotoURL = photoURL

	return domain.Student{
		FullName:  fullName,
		BirthDate: birthDate,
		Gender:    gender,
		Phone:     phone,
		Email:     email,
		CPF:       cpf,
		Address:   address,
		Notes:     notes,
		PhotoURL:  photoURL,
		Status:    status,
	}, nil
}

func parseStudentStatus(value string) (domain.StudentStatus, error) {
	switch strings.ToLower(value) {
	case "", string(domain.StudentActive):
		return domain.StudentActive, nil
	case string(domain.StudentInactive):
		return domain.StudentInactive, nil
	case string(domain.StudentSuspended):
		return domain.StudentSuspended, nil
	default:
		return "", errors.New("Status invalido.")
	}
}

func statusFilter(value string) []domain.StudentStatus {
	switch strings.ToLower(value) {
	case "", string(domain.StudentActive):
		return []domain.StudentStatus{domain.StudentActive}
	case string(domain.StudentInactive):
		return []domain.StudentStatus{domain.StudentInactive}
	case string(domain.StudentSuspended):
		return []domain.StudentStatus{domain.StudentSuspended}
	case "all":
		return nil
	default:
		return []domain.StudentStatus{domain.StudentActive}
	}
}

func normalizeStudentStatusValue(value string) string {
	switch strings.ToLower(value) {
	case "all":
		return "all"
	case string(domain.StudentInactive):
		return string(domain.StudentInactive)
	case string(domain.StudentSuspended):
		return string(domain.StudentSuspended)
	default:
		return string(domain.StudentActive)
	}
}

func statusPresentation(status domain.StudentStatus) (string, string) {
	switch status {
	case domain.StudentInactive:
		return "Inativo", "rounded-full bg-slate-700/50 px-3 py-1 text-slate-300"
	case domain.StudentSuspended:
		return "Suspenso", "rounded-full bg-amber-400/10 px-3 py-1 text-amber-200"
	default:
		return "Ativo", "rounded-full bg-emerald-400/10 px-3 py-1 text-emerald-200"
	}
}

func (h *Handler) buildStudentsData(r *http.Request) view.StudentsPageData {
	query := strings.TrimSpace(r.FormValue("q"))
	status := normalizeStudentStatusValue(strings.TrimSpace(r.FormValue("status")))

	data := view.StudentsPageData{
		Query:  query,
		Status: status,
	}

	if h.services.Students != nil {
		filter := ports.StudentFilter{
			Query:    query,
			Statuses: statusFilter(status),
			Limit:    50,
		}
		students, err := h.services.Students.Search(r.Context(), filter)
		if err != nil {
			log.Printf("list students: %v", err)
		} else {
			data.Items = make([]view.StudentItem, 0, len(students))
			for _, student := range students {
				label, className := statusPresentation(student.Status)
				item := view.StudentItem{
					ID:          student.ID,
					FullName:    student.FullName,
					BirthDate:   formatDateBR(student.BirthDate),
					Phone:       student.Phone,
					Email:       student.Email,
					Status:      string(student.Status),
					StatusLabel: label,
					StatusClass: className,
				}
				data.Items = append(data.Items, item)
			}
		}
	}

	return data
}

func formatDateBR(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("02/01/2006")
}

func parseDateInput(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	if strings.Contains(value, "-") {
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			return nil, err
		}
		return &parsed, nil
	}
	parsed, err := time.Parse("02/01/2006", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func formatDateInputBR(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("02/01/2006")
}
