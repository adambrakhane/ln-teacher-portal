package api

type Class struct {
	ID        int64  `db:"id"`
	TeacherID int64  `db:"teacher_id"`
	Name      string `db:"name"`

	Schedule []WorkshopSchedule
}
