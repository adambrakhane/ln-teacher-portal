package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type WorkshopResource struct {
	ID          int64  `db:"id"`
	Description string `db:"description"`
	Name        string `db:"name"`
}

type Workshop struct {
	ID          int64  `db:"id"`
	Description string `db:"description"`
	Name        string `db:"name"`
	CourseID    int64  `db:"course_id"`
}

type WorkshopSchedule struct {
	ID          int64 `db:"id"`
	Workshop    Workshop
	ScheduledAt time.Time `db:"scheduled_at"`
	Resources   []WorkshopResource
}

func getClassWorkshopSchedule(w http.ResponseWriter, r *http.Request) {
	classID, _ := strconv.ParseInt(chi.URLParam(r, "classID"), 10, 64)

	schedule, err := getWorkshopScheduleForClass(classID)
	if err != nil {
		writeError(w, "Error getting workshop schedule for class", http.StatusInternalServerError)
		return
	}

	response, _ := json.Marshal(schedule)
	w.Write(response)

}

func getWorkshopScheduleForClass(classID int64) (schedule []WorkshopSchedule, err error) {
	schedule = make([]WorkshopSchedule, 0)
	// Get workshops scheduled
	rows, err := db.Query(`
	SELECT
		ws.id,
		ws.scheduled_at,

		w.name,
		w.description,
		w.course_id,
		w.id,

		(select json_agg(wr) from workshop_resources wr where wr.workshop_id = w.id)
	FROM workshop_schedule ws
	LEFT JOIN workshops w on ws.workshop_id = w.id
	WHERE
		ws.class_id = $1
	`, classID)
	if err != nil {
		return
	}

	// Parse each response row
	for rows.Next() {
		fmt.Println("Schedule item!")
		scheduleItem := WorkshopSchedule{}
		workshop := Workshop{}
		resources := []WorkshopResource{}

		var resourceJSON string

		rows.Scan(
			&scheduleItem.ID,
			&scheduleItem.ScheduledAt,

			&workshop.Name,
			&workshop.Description,
			&workshop.CourseID,
			&workshop.ID,

			&resourceJSON,
		)

		fmt.Println(workshop)
		// Parse resource data
		fmt.Println("Resources")
		fmt.Println(string(resourceJSON))
		json.Unmarshal([]byte(resourceJSON), &resources)

		scheduleItem.Workshop = workshop
		scheduleItem.Resources = resources

		schedule = append(schedule, scheduleItem)
	}
	return schedule, nil
}
