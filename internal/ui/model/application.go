package model

import "github.com/mokiat/lacking/ui/mvc"

const (
	ViewNameIntro   ViewName = "intro"
	ViewNameLoading ViewName = "loading"
	ViewNamePlay    ViewName = "play"
)

type ViewName = string

func NewApplication(eventBus *mvc.EventBus) *Application {
	return &Application{
		eventBus:   eventBus,
		activeView: ViewNameIntro,
	}
}

type Application struct {
	eventBus   *mvc.EventBus
	activeView ViewName
}

func (a *Application) ActiveView() ViewName {
	return a.activeView
}

func (a *Application) SetActiveView(view ViewName) {
	a.activeView = view
	a.eventBus.Notify(&AppActiveViewSetEvent{
		ActiveView: view,
	})
}

type AppActiveViewSetEvent struct {
	ActiveView ViewName
}
