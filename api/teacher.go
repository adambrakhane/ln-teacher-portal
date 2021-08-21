package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Teacher struct {
	ID       int64  `db:"id"`
	FullName string `db:"full_name"`
	Password string `db:"password" json:"-"`

	Classes []Class
}

// Middleware handler that 1) validates the teacher token and 2) loads all that teacher's data into the request context
func TeacherCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		teacherID, _ := strconv.ParseInt(chi.URLParam(r, "teacherID"), 10, 64)

		// Get teacher data and store it in the context
		teacher, err := getTeacherByID(teacherID)
		if err != nil {
			log.Println(err)
			writeError(w, "TeacherNotFound", http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), "teacher", teacher)
		fmt.Println(teacher)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Get teacher data from the context and return it to the user
func getTeacher(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teacher, ok := ctx.Value("teacher").(Teacher)
	if !ok {
		writeError(w, "TeacherNotFound", http.StatusNotFound)

		return
	}

	response, _ := json.Marshal(teacher)
	w.Write(response)
}

// DB HELPERS

// This gets the teacher data & all corresponding class data
// This would be good enough to start as the amount of data is small, but of course
// this function would be broken into sub functions as required
func getTeacherByID(id int64) (teacher Teacher, err error) {
	err = db.Get(&teacher, `
		SELECT 
			id,
			full_name,
			password
		FROM teachers WHERE id = $1
	`, id)

	if err != nil {
		return
	}

	// Get classes too
	err = db.Select(&teacher.Classes, `
		SELECT
			id,
			teacher_id,
			name
		FROM classes WHERE teacher_id = $1
		`, teacher.ID)

	if err != nil {
		return
	}

	// Load the workshop schedule for each class
	for i, class := range teacher.Classes {
		schedule, err := getWorkshopScheduleForClass(class.ID)
		if err != nil {
			return teacher, err
		}
		teacher.Classes[i].Schedule = schedule
	}

	return
}
