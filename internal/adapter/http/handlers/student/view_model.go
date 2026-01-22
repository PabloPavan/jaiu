package handlers

type StudentItem struct {
	ID              string
	FullName        string
	BirthDate       string
	Phone           string
	Email           string
	PhotoURL        string
	Initials        string
	Status          string
	StatusLabel     string
	PlanName        string
	LastPaymentDate string
	LastPaymentInfo string
}

type StudentsPageData struct {
	Query                string
	Status               string
	Page                 int
	PageSize             int
	TotalItems           int
	TotalPages           int
	StartIndex           int
	EndIndex             int
	TotalStudents        int
	ActiveStudents       int
	InactiveStudents     int
	SuspendedStudents    int
	OverduePayments      int
	NewStudentsThisMonth int
	Items                []StudentItem
}

type StudentFormData struct {
	Title          string
	Action         string
	SubmitLabel    string
	DeleteAction   string
	ShowDelete     bool
	FullName       string
	BirthDate      string
	Gender         string
	Phone          string
	Email          string
	CPF            string
	Address        string
	Notes          string
	PhotoObjectKey string
	PhotoURL       string
	Status         string
	Error          string
}

type StudentOption struct {
	ID   string
	Name string
}
