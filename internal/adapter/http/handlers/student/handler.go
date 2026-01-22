package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	handler "github.com/PabloPavan/jaiu/internal/adapter/http/handlers"
	view "github.com/PabloPavan/jaiu/internal/adapter/http/handlers/student/views"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/observability"
	"github.com/PabloPavan/jaiu/internal/ports"
	studentapp "github.com/PabloPavan/jaiu/internal/student/app"
	"github.com/go-chi/chi/v5"
)

type StudentHandler struct {
	app     *studentapp.StudentApp
	handler *handler.Handler
}

func New(app *studentapp.StudentApp, handler *handler.Handler) *StudentHandler {
	return &StudentHandler{app: app, handler: handler}
}

const studentsPageSize = 5

func (h *StudentHandler) StudentsIndex(w http.ResponseWriter, r *http.Request) {
	data := h.buildStudentsData(r)
	h.handler.RenderHTMXOrPage(w, r, "Alunos", view.StudentsPage(data), view.StudentsContent(data))
}

func (h *StudentHandler) StudentsPreview(w http.ResponseWriter, r *http.Request) {
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
			observability.Logger(r.Context()).Error("failed to load students preview", "err", err)
		} else {
			data.Items = make([]view.StudentItem, 0, len(students))
			for _, student := range students {
				item := view.StudentItem{
					ID:          student.ID,
					FullName:    student.FullName,
					BirthDate:   formatDateBR(student.BirthDate),
					Phone:       student.Phone,
					Email:       student.Email,
					PhotoURL:    h.photoURLForVariant(student.PhotoObjectKey, "list"),
					Initials:    studentInitials(student.FullName),
					Status:      string(student.Status),
					StatusLabel: statusPresentation(student.Status),
				}
				data.Items = append(data.Items, item)
			}
		}
	}

	h.renderComponent(w, r, view.StudentsPreview(data))
}

func (h *StudentHandler) StudentsNew(w http.ResponseWriter, r *http.Request) {
	data := studentFormCreateData()
	h.renderPage(w, r, page("Novo aluno", view.StudentFormPage(data)))
}

func (h *StudentHandler) StudentsCreate(w http.ResponseWriter, r *http.Request) {
	data := studentFormCreateData()
	student, err := h.parseStudentForm(r, &data)
	if err != nil {
		data.Error = err.Error()
		h.renderFormError(w, r, data.Title, view.StudentFormPage(data))
		return
	}

	if h.services.Students == nil {
		data.Error = "Servico de alunos indisponivel."
		h.renderFormError(w, r, data.Title, view.StudentFormPage(data))
		return
	}

	_, err = h.services.Students.Register(r.Context(), student)
	if err != nil {
		data.Error = "Nao foi possivel salvar o aluno."
		h.renderFormError(w, r, data.Title, view.StudentFormPage(data))
		return
	}

	h.redirectHTMXOrRedirect(w, r, "/students")
}

func (h *StudentHandler) StudentsEdit(w http.ResponseWriter, r *http.Request) {
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
		observability.Logger(r.Context()).Error("failed to load student", "err", err)
		http.Error(w, "Erro ao carregar aluno.", http.StatusInternalServerError)
		return
	}

	data := studentFormEditData(student, h.photoURLForVariant(student.PhotoObjectKey, "preview"))
	h.renderPage(w, r, page(data.Title, view.StudentFormPage(data)))
}

func (h *StudentHandler) StudentsUpdate(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "studentID")
	data := studentFormEditData(domain.Student{ID: studentID}, "")
	student, err := h.parseStudentForm(r, &data)
	if err != nil {
		data.Error = err.Error()
		h.renderFormError(w, r, data.Title, view.StudentFormPage(data))
		return
	}

	if h.services.Students == nil {
		data.Error = "Servico de alunos indisponivel."
		h.renderFormError(w, r, data.Title, view.StudentFormPage(data))
		return
	}

	student.ID = studentID
	_, err = h.services.Students.Update(r.Context(), student)
	if err != nil {
		data.Error = "Nao foi possivel atualizar o aluno."
		h.renderFormError(w, r, data.Title, view.StudentFormPage(data))
		return
	}

	h.redirectHTMXOrRedirect(w, r, "/students")
}

func (h *StudentHandler) StudentsDelete(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "studentID")
	if h.services.Students == nil {
		http.NotFound(w, r)
		return
	}

	_, err := h.services.Students.Deactivate(r.Context(), studentID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		observability.Logger(r.Context()).Error("failed to deactivate student", "err", err)
		http.Error(w, "Erro ao excluir aluno.", http.StatusInternalServerError)
		return
	}

	h.renderHTMXOrRedirect(w, r, "/students", func() {
		data := h.buildStudentsData(r)
		h.renderComponent(w, r, view.StudentsList(data))
	})
}

func studentFormCreateData() view.StudentFormData {
	return view.StudentFormData{
		Title:       "Novo aluno",
		Action:      "/students",
		SubmitLabel: "Criar aluno",
		Status:      string(domain.StudentActive),
	}
}

func studentFormEditData(student domain.Student, photoURL string) view.StudentFormData {
	return view.StudentFormData{
		Title:          "Editar aluno",
		Action:         "/students/" + student.ID,
		SubmitLabel:    "Salvar",
		DeleteAction:   "/students/" + student.ID + "/delete",
		ShowDelete:     student.ID != "",
		FullName:       student.FullName,
		BirthDate:      formatDateInputBR(student.BirthDate),
		Gender:         student.Gender,
		Phone:          student.Phone,
		Email:          student.Email,
		CPF:            student.CPF,
		Address:        student.Address,
		Notes:          student.Notes,
		PhotoObjectKey: student.PhotoObjectKey,
		PhotoURL:       photoURL,
		Status:         string(student.Status),
	}
}

func (h *StudentHandler) parseStudentForm(r *http.Request, data *view.StudentFormData) (domain.Student, error) {
	maxMemory := h.images.MaxMemory
	if maxMemory == 0 {
		maxMemory = 32 << 20
	}
	if err := r.ParseMultipartForm(maxMemory); err != nil {
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
		return domain.Student{}, errors.New("Data de nascimento invalida. Use o formato dd/mm/aaaa.")
	}

	gender := strings.TrimSpace(r.FormValue("gender"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	email := strings.TrimSpace(r.FormValue("email"))
	cpf := strings.TrimSpace(r.FormValue("cpf"))
	address := strings.TrimSpace(r.FormValue("address"))
	notes := strings.TrimSpace(r.FormValue("notes"))
	photoObjectKey := strings.TrimSpace(r.FormValue("photo_object_key"))
	uploadedObjectKey := ""
	file, header, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		if h.images.ImageService == nil {
			return domain.Student{}, errors.New("Upload de foto indisponivel.")
		}
		uploadedObjectKey, err = h.images.ImageService.UploadImage(r.Context(), file, header)
		if err != nil {
			return domain.Student{}, errors.New("Nao foi possivel salvar a foto.")
		}
	} else if !errors.Is(err, http.ErrMissingFile) {
		return domain.Student{}, errors.New("Nao foi possivel ler a foto.")
	}
	if uploadedObjectKey != "" {
		photoObjectKey = uploadedObjectKey
	}

	data.Gender = gender
	data.Phone = phone
	data.Email = email
	data.CPF = cpf
	data.Address = address
	data.Notes = notes
	data.PhotoObjectKey = photoObjectKey
	data.PhotoURL = h.photoURLForVariant(photoObjectKey, "preview")

	return domain.Student{
		FullName:       fullName,
		BirthDate:      birthDate,
		Gender:         gender,
		Phone:          phone,
		Email:          email,
		CPF:            cpf,
		Address:        address,
		Notes:          notes,
		PhotoObjectKey: photoObjectKey,
		Status:         status,
	}, nil
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
	case string(domain.StudentActive):
		return string(domain.StudentActive)
	case string(domain.StudentInactive):
		return string(domain.StudentInactive)
	case string(domain.StudentSuspended):
		return string(domain.StudentSuspended)
	default:
		return "all"
	}
}

func statusPresentation(status domain.StudentStatus) string {
	switch status {
	case domain.StudentInactive:
		return "Inativo"
	case domain.StudentSuspended:
		return "Suspenso"
	default:
		return "Ativo"
	}
}

func (h *StudentHandler) fetchStudentCount(ctx context.Context, statuses []domain.StudentStatus) (int, error) {
	if h.services.Students == nil {
		return 0, nil
	}
	count, err := h.services.Students.Count(ctx, ports.StudentFilter{
		Statuses: statuses,
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (h *StudentHandler) buildStudentsData(r *http.Request) view.StudentsPageData {
	query := strings.TrimSpace(r.FormValue("q"))
	status := normalizeStudentStatusValue(strings.TrimSpace(r.FormValue("status")))

	page := 1
	if value := strings.TrimSpace(r.FormValue("page")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			page = parsed
		}
	}

	statuses := statusFilter(status)

	data := view.StudentsPageData{
		Query:      query,
		Status:     status,
		PageSize:   studentsPageSize,
		Page:       page,
		TotalPages: 1,
	}

	if h.services.Students != nil {
		if total, err := h.fetchStudentCount(r.Context(), nil); err != nil {
			observability.Logger(r.Context()).Error("failed to count students", "err", err)
		} else {
			data.TotalStudents = total
		}
		if total, err := h.fetchStudentCount(r.Context(), []domain.StudentStatus{domain.StudentActive}); err == nil {
			data.ActiveStudents = total
		}
		if total, err := h.fetchStudentCount(r.Context(), []domain.StudentStatus{domain.StudentInactive}); err == nil {
			data.InactiveStudents = total
		}
		if total, err := h.fetchStudentCount(r.Context(), []domain.StudentStatus{domain.StudentSuspended}); err == nil {
			data.SuspendedStudents = total
		}

		count, err := h.fetchStudentCount(r.Context(), statuses)
		if err != nil {
			observability.Logger(r.Context()).Error("failed to count students", "err", err)
		} else {
			data.TotalItems = count
		}

		totalPages := 1
		if data.TotalItems > 0 {
			totalPages = (data.TotalItems + studentsPageSize - 1) / studentsPageSize
		}
		if totalPages == 0 {
			totalPages = 1
		}
		if page < 1 {
			page = 1
		}
		if page > totalPages {
			page = totalPages
		}
		data.Page = page
		data.TotalPages = totalPages

		offset := (page - 1) * studentsPageSize
		filter := ports.StudentFilter{
			Query:    query,
			Statuses: statuses,
			Limit:    studentsPageSize,
			Offset:   offset,
		}
		students, err := h.services.Students.Search(r.Context(), filter)
		if err != nil {
			observability.Logger(r.Context()).Error("failed to list students", "err", err)
		} else {
			data.Items = make([]view.StudentItem, 0, len(students))
			for _, student := range students {
				item := view.StudentItem{
					ID:              student.ID,
					FullName:        student.FullName,
					BirthDate:       formatDateBR(student.BirthDate),
					Phone:           student.Phone,
					Email:           student.Email,
					PhotoURL:        h.photoURLForVariant(student.PhotoObjectKey, "list"),
					Initials:        studentInitials(student.FullName),
					Status:          string(student.Status),
					StatusLabel:     statusPresentation(student.Status),
					PlanName:        "",
					LastPaymentDate: "",
					LastPaymentInfo: "",
				}
				data.Items = append(data.Items, item)
			}
			if len(data.Items) > 0 {
				data.StartIndex = offset + 1
				data.EndIndex = offset + len(data.Items)
				if data.EndIndex > data.TotalItems {
					data.EndIndex = data.TotalItems
				}
			}
		}
	}

	return data
}

func (h *StudentHandler) photoURLForVariant(objectKey, variant string) string {
	if objectKey == "" || variant == "" {
		return ""
	}
	baseURL := strings.TrimRight(h.images.BaseURL, "/")
	if baseURL == "" {
		baseURL = "/images"
	}
	return baseURL + "/" + strings.TrimSpace(objectKey) + "/" + variant
}

func studentInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return ""
	}
	first := initialFrom(parts[0])
	if len(parts) == 1 {
		return first
	}
	last := initialFrom(parts[len(parts)-1])
	return first + last
}

func initialFrom(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) == 0 {
		return ""
	}
	return strings.ToUpper(string(runes[0]))
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
