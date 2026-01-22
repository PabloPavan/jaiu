package app

import (
	"github.com/PabloPavan/jaiu/internal/student/app/commands"
	"github.com/PabloPavan/jaiu/internal/student/app/queries"
)

type StudentApp struct {
	// Commands (write)
	Create *commands.RegisterStudent
	// Update *commands.UpdateStudent
	// Delete *commands.DeleteStudent

	// Queries (read)
	GetGrid *queries.GetIndexPageStudent
	//List *queries.ListStudents
}
