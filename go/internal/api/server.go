package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/flowci/flowci/internal/api/handlers"
	"github.com/flowci/flowci/internal/api/middleware"
	"github.com/flowci/flowci/internal/builder"
	"github.com/flowci/flowci/internal/config"
	"github.com/flowci/flowci/internal/deployer"
	"github.com/flowci/flowci/pkg/docker"
)

type Server struct {
	router      *chi.Mux
	buildHandler   *handlers.BuildHandler
	deployHandler  *handlers.DeployHandler
	dockerHandler  *handlers.DockerHandler
	projectHandler *handlers.ProjectHandler
	healthHandler  *handlers.HealthHandler
}

func NewServer(
	dc *docker.Client,
	b *builder.Builder,
	d *deployer.Deployer,
	cm *config.Manager,
	version string,
) *Server {
	s := &Server{
		router:         chi.NewRouter(),
		buildHandler:   handlers.NewBuildHandler(b, dc),
		deployHandler:  handlers.NewDeployHandler(d),
		dockerHandler:  handlers.NewDockerHandler(dc),
		projectHandler: handlers.NewProjectHandler(cm),
		healthHandler:  handlers.NewHealthHandler(dc, version),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.ErrorHandler)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
}

func (s *Server) setupRoutes() {
	s.router.Get("/health", s.healthHandler.Check)

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Route("/builds", func(r chi.Router) {
			r.Post("/", s.buildHandler.CreateBuild)
			r.Get("/{id}", s.buildHandler.GetBuild)
			r.Get("/{id}/logs", s.buildHandler.GetBuildLogs)
		})

		r.Route("/deploys", func(r chi.Router) {
			r.Post("/", s.deployHandler.CreateDeploy)
			r.Get("/{id}", s.deployHandler.GetDeployStatus)
			r.Post("/{id}/rollback", s.deployHandler.RollbackDeploy)
		})

		r.Route("/docker", func(r chi.Router) {
			r.Get("/check", s.dockerHandler.CheckConnection)
			r.Get("/images", s.dockerHandler.ListImages)
			r.Get("/containers", s.dockerHandler.ListContainers)
		})

		r.Route("/projects", func(r chi.Router) {
			r.Get("/", s.projectHandler.ListProjects)
			r.Post("/", s.projectHandler.CreateProject)
			r.Get("/{id}", s.projectHandler.GetProject)
			r.Put("/{id}", s.projectHandler.UpdateProject)
			r.Delete("/{id}", s.projectHandler.DeleteProject)
		})
	})
}

func (s *Server) Router() *chi.Mux {
	return s.router
}
