package web

import (
	"log"
	"net/http"

	"github.com/open-telemetry/opamp-go/internal/examples/server/data"
)

var (
	agentsStore *data.Agents
	httpServer  *http.Server
)

// View models expected by ui.go.

type AgentView struct {
	Name       string
	Type       string
	InstanceUID string
	Connected  bool
	Healthy    bool
	Status     string
	Version    string
	Attributes map[string]string
	Deployment DeploymentView
}

type DeploymentView struct {
	Runtime string
}

type AgentItemView struct {
	Agent          AgentView
	Groups         []GroupView
	LastDeployment *DeploymentViewModel
}

type GroupView struct {
	ID   string
	Name string
}

type DeploymentViewModel struct {
	ConfigurationName    string
	ConfigurationVersion string
	Status               string
}

type FleetPageView struct {
	Page string

	Known   int
	Active  int
	Offline int
	Retired int

	HealthyPercent int
	Healthy        int
	Attention      int

	LastConnectedAgent string
	LastConnectedGroup string

	StatusFilter string
	Items        []AgentItemView
}

func Start(store *data.Agents) {
	agentsStore = store

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/agents", http.StatusFound)
	})

	mux.HandleFunc("/agents", agentsHandler)

	httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("[FleetAMP] UI listening on :8080")

		if err := httpServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("[FleetAMP] web server error: %v", err)
		}
	}()
}

func agentsHandler(w http.ResponseWriter, r *http.Request) {
	agents := agentsStore.GetAllAgentsReadonlyClone()

	view := FleetPageView{
		Page:         "fleet",
		StatusFilter: "all",
		Items:        make([]AgentItemView, 0, len(agents)),
	}

	for _, agent := range agents {
		healthy := false

		if agent.Status != nil &&
			agent.Status.Health != nil {
			healthy = agent.Status.Health.Healthy
		}

		name := agent.ServiceName()

		if name == "" || name == "Unknown" {
			name = agent.InstanceIdStr
		}

		item := AgentItemView{
			Agent: AgentView{
				Name:        name,
				Type:        "OpenTelemetry Collector",
				InstanceUID: agent.InstanceIdStr,
				Connected:   true,
				Healthy:     healthy,
				Status:      agent.HealthStatus(),
				Version:     agent.ServiceVersion(),

				Attributes: map[string]string{},

				Deployment: DeploymentView{
					Runtime: "OpenTelemetry",
				},
			},
			Groups:         nil,
			LastDeployment: nil,
		}

		view.Items = append(view.Items, item)

		view.Known++
		view.Active++

		if healthy {
			view.Healthy++
		} else {
			view.Attention++
		}

		view.LastConnectedAgent = name
	}

	if view.Known > 0 {
		view.HealthyPercent = (view.Healthy * 100) / view.Known
	}

	if err := agentsPage.Execute(w, view); err != nil {
		log.Printf("[FleetAMP] template error: %v", err)
		http.Error(
			w,
			"FleetAMP template rendering error",
			http.StatusInternalServerError,
		)
	}
}
