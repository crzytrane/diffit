package main

import (
	"bufio"
	"cmp"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/crzytrane/diffit/internal/archive"
	"github.com/crzytrane/diffit/internal/config"
	"github.com/crzytrane/diffit/internal/database"
	"github.com/crzytrane/diffit/internal/diffimage"
	"github.com/crzytrane/diffit/internal/handlers"
	"github.com/crzytrane/diffit/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/n7olkachev/imgdiff/pkg/imgdiff"
)

func corsMiddleware(allowOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			isDev := strings.HasPrefix(r.Host, "localhost")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			// In dev mode, allow localhost origins
			if isDev && strings.Contains(origin, "localhost") {
				allowed = true
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if isDev {
				w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database,omitempty"`
}

func main() {
	cfg := config.Load()

	envOrDefaultPort := cmp.Or(os.Getenv("PORT"), cfg.Port)
	envOrDefaultPortInt, err := strconv.Atoi(envOrDefaultPort)
	if err != nil {
		panic("Couldn't convert port number, check that it is a valid value")
	}
	port := flag.Int("port", envOrDefaultPortInt, "port to run on")
	flag.Parse()

	// Initialize database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
		log.Printf("Running without database - only legacy /files endpoint will work")
	} else {
		defer db.Close()

		// Run migrations
		if err := db.Migrate(context.Background()); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Database connected and migrations complete")
	}

	// Initialize storage
	store, err := storage.New(cfg.StoragePath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	log.Printf("Storage initialized at %s", cfg.StoragePath)

	// Initialize handlers
	var h *handlers.Handlers
	if db != nil {
		h = handlers.New(db.Pool, store)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(corsMiddleware(cfg.AllowOrigins))

	// Health check
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!\n"))
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		resp := healthResponse{Status: "ok"}
		if db != nil {
			if err := db.Pool.Ping(context.Background()); err != nil {
				resp.Database = "disconnected"
			} else {
				resp.Database = "connected"
			}
		} else {
			resp.Database = "not configured"
		}

		json.NewEncoder(w).Encode(resp)
	})

	// API routes (only if database is connected)
	if h != nil {
		r.Route("/api", func(r chi.Router) {
			// Projects
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", h.Projects.List)
				r.Post("/", h.Projects.Create)
				r.Get("/slug/{slug}", h.Projects.GetBySlug)
				r.Route("/{projectID}", func(r chi.Router) {
					r.Get("/", h.Projects.Get)
					r.Put("/", h.Projects.Update)
					r.Delete("/", h.Projects.Delete)

					// Nested builds
					r.Get("/builds", h.Builds.ListByProject)
					r.Get("/builds/latest", h.Builds.GetLatest)

					// Nested baselines
					r.Get("/baselines", h.Baselines.ListByProject)
				})
			})

			// Builds
			r.Route("/builds", func(r chi.Router) {
				r.Post("/", h.Builds.Create)
				r.Route("/{buildID}", func(r chi.Router) {
					r.Get("/", h.Builds.Get)
					r.Delete("/", h.Builds.Delete)
					r.Patch("/status", h.Builds.UpdateStatus)
					r.Post("/finalize", h.Builds.Finalize)

					// Nested snapshots
					r.Get("/snapshots", h.Snapshots.ListByBuild)
					r.Get("/snapshots/changed", h.Snapshots.GetChanged)
				})
			})

			// Snapshots
			r.Route("/snapshots", func(r chi.Router) {
				r.Post("/", h.Snapshots.Create)
				r.Post("/batch-review", h.Snapshots.BatchReview)
				r.Route("/{snapshotID}", func(r chi.Router) {
					r.Get("/", h.Snapshots.Get)
					r.Delete("/", h.Snapshots.Delete)
					r.Post("/review", h.Snapshots.Review)
					r.Get("/image/{imageType}", h.Snapshots.GetImage)
				})
			})

			// Baselines
			r.Route("/baselines", func(r chi.Router) {
				r.Post("/", h.Baselines.Create)
				r.Post("/from-snapshot", h.Baselines.CreateFromSnapshot)
				r.Route("/{baselineID}", func(r chi.Router) {
					r.Get("/", h.Baselines.Get)
					r.Delete("/", h.Baselines.Delete)
					r.Get("/image", h.Baselines.GetImage)
				})
			})
		})
	}

	// Legacy /files endpoint for backwards compatibility
	r.Post("/files", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(32 << 20) // 32 MB max
		if err != nil {
			fmt.Printf("error processing multipart form, err %s\n", err.Error())
			return
		}

		baseUpload, _, err := r.FormFile("file-base")
		if err != nil {
			fmt.Printf("err uploading base file\n")
			return
		}
		defer baseUpload.Close()

		otherUpload, _, err := r.FormFile("file-other")
		if err != nil {
			fmt.Printf("err uploading feature file\n")
			return
		}
		defer otherUpload.Close()

		dst, err := os.MkdirTemp("", "extracted-")
		if err != nil {
			return
		}

		baseFile, err := os.CreateTemp(dst, "base-*.png")
		if err != nil {
			return
		}
		defer baseFile.Close()
		otherFile, err := os.CreateTemp(dst, "other-*.png")
		if err != nil {
			return
		}
		defer otherFile.Close()
		diffFile, err := os.CreateTemp(dst, "diff-*.png")
		if err != nil {
			return
		}
		defer diffFile.Close()

		_, err = io.Copy(baseFile, baseUpload)
		if err != nil {
			fmt.Print("failed to copy base file\n")
			return
		}
		_, err = io.Copy(otherFile, otherUpload)
		if err != nil {
			fmt.Print("failed to copy other file\n")
			return
		}

		_, err = baseFile.Seek(0, io.SeekStart)
		if err != nil {
			fmt.Print("failed to seek files\n")
			return
		}
		_, err = otherFile.Seek(0, io.SeekStart)
		if err != nil {
			fmt.Print("failed to seek files\n")
			return
		}

		image1, _, err := image.Decode(baseFile)
		if err != nil {
			fmt.Print("failed to decode base file\n")
			return
		}

		image2, _, err := image.Decode(otherFile)
		if err != nil {
			fmt.Print("failed to decode other file\n")
			return
		}

		threshold := float64(0.1)
		resultDiff := *imgdiff.Diff(image1, image2, &imgdiff.Options{
			Threshold: threshold,
			DiffImage: false,
		})

		writer := bufio.NewWriter(diffFile)

		enc := &png.Encoder{
			CompressionLevel: png.BestSpeed,
		}

		w.Header().Add("Content-Type", "image/png")

		w.WriteHeader(http.StatusOK)

		err = enc.Encode(w, resultDiff.Image)

		if err != nil {
			fmt.Print("failed to encode diff file\n")
		}
		err = writer.Flush()
		if err != nil {
			fmt.Print("failed to flush diff file\n")
		}

		log.Printf("diff path is %s, isEqual: %v", diffFile.Name(), resultDiff.Equal)
	})

	// Legacy /archive endpoint
	r.Post("/archive", func(w http.ResponseWriter, r *http.Request) {
		baseFilePath, err := archive.UnpackArchiveFromRequest(r)

		baseDir := fmt.Sprintf("%s/base/", baseFilePath)
		featureDir := fmt.Sprintf("%s/feature/", baseFilePath)
		diffDir := fmt.Sprintf("%s/diff/", baseFilePath)

		fmt.Printf("baseFilePath is: %s\n", baseFilePath)

		fromDirectoryOptions := diffimage.FromDirectoryOptions{
			BaseDir:    baseDir,
			FeatureDir: featureDir,
			DiffDir:    diffDir,
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		files, err := diffimage.GetDiffsFromDirectory(fromDirectoryOptions)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		diffs := make([]diffimage.DiffResult, len(files))
		for index, diff := range files {
			diffResult, err := diffimage.DiffImage(diff, diffimage.DiffOptions{Threshold: 0.1})
			fmt.Printf("Diff %d\n\t- %s\n\t- %s\n\t- %s\n", index, diffResult.Input.BasePath, diffResult.Input.FeaturePath, diffResult.Input.DiffPath)
			diffs[index] = diffResult
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		diffTest := diffs[0].Input
		fmt.Printf("diff count %d\n\t- %s\n\t- %s\n\t- %s\n", len(diffs), diffTest.BasePath, diffTest.FeaturePath, diffTest.DiffPath)

		archive.ArchiveData(baseDir, featureDir, diffDir)

		http.Redirect(w, r, "http://localhost:5173/", http.StatusFound)
	})

	fmt.Printf("Listening on port %v\n", *port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", *port), r)

	fmt.Printf("Finished: %v", err)
}
