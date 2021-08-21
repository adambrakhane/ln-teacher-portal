package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

var hmacSampleSecret = "be curious, not judgemental"

func Serve() {
	var err error
	dsn := os.Getenv("dsn")
	if dsn == "" {
		dsn = "host=localhost user=local_access password=man_wearing_coat dbname=adam port=5433 sslmode=disable"
	}

	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Could not open database connection:\n%v\n===\n", err)
	}

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Utility endpoints
	r.Post("/login", login)
	// r.Post("/heartbeat", login)
	// r.Post("/version", login)
	// r.Post("/version", login)

	// This section would contain all of the teacher resources
	r.Route("/teachers", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Route("/{teacherID}", func(r chi.Router) {
			r.Use(TeacherCtx)      // This allows us to run middleware for just this sub-router. This function will load the teacher data for use by any of the children endpoints.
			r.Get("/", getTeacher) // GET /teachers/135

		})
	})

	// Class data
	r.Route("/classes", func(r chi.Router) { // /classes/1
		r.Use(authMiddleware)
		r.Route("/{classID}", func(r chi.Router) {
			r.Get("/schedule", getClassWorkshopSchedule) // /classes/1/schedule
		})
	})

	// Serve static files too
	FileServer(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.ListenAndServe(fmt.Sprintf(":%v", port), r)
}

// Helper to write pretty JSON errors
func writeError(w http.ResponseWriter, err string, status int) {
	var response struct {
		Message string `json:"message"`
		Error   bool   `json:"error"`
	}

	response.Error = true
	response.Message = err

	responseBytes, _ := json.Marshal(response)

	http.Error(w, string(responseBytes), status)
}

// We'll use this to serve the static HTML
func FileServer(router *chi.Mux) {
	root := "./web"
	fs := http.FileServer(http.Dir(root))

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(root + r.RequestURI); os.IsNotExist(err) {
			http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		} else {
			fs.ServeHTTP(w, r)
		}
	})
}
