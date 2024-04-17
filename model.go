package castle

type Event struct {
	EventType   EventType
	EventStatus EventStatus
}

// EventType is an enum defining types of event castle tracks.
type EventType string

// See https://docs.castle.io/docs/events
const (
	EventTypeLogin                EventType = "$login"
	EventTypeRegistration         EventType = "$registration"
	EventTypeProfileUpdate        EventType = "$profile_update"
	EventTypeProfileReset         EventType = "$profile_reset"
	EventTypePasswordResetRequest EventType = "$password_reset_request"
	EventTypeChallenge            EventType = "$challenge"
	EventTypeLogout               EventType = "&logout"
)

// EventStatus is an enum defining the statuses for a given event.
type EventStatus string

// See https://docs.castle.io/docs/events
const (
	EventStatusAttempted EventStatus = "$attempted"
	EventStatusSucceeded EventStatus = "$succeeded"
	EventStatusFailed    EventStatus = "$failed"
	EventStatusRequested EventStatus = "$requested"
)

// RecommendedAction encapsulates the 3 possible responses from auth call (allow, challenge, deny).
type RecommendedAction string

// See https://castle.io/docs/authentication
const (
	RecommendedActionNone      RecommendedAction = ""
	RecommendedActionAllow     RecommendedAction = "allow"
	RecommendedActionChallenge RecommendedAction = "challenge"
	RecommendedActionDeny      RecommendedAction = "deny"
)

// Request wraps the Castle required data for to the pkg's Risk and Filter methods.
type Request struct {
	Context    *Context
	Event      Event
	User       User
	Properties map[string]string
}

// Context captures data from HTTP request.
type Context struct {
	IP           string            `json:"ip"`
	Headers      map[string]string `json:"headers"`
	RequestToken string            `json:"request_token"`
}

type User struct {
	ID           string            `json:"id"`
	Email        string            `json:"email,omitempty"`
	Phone        string            `json:"phone,omitempty"`
	Name         string            `json:"name,omitempty"`
	RegisteredAt string            `json:"registered_at,omitempty"`
	Traits       map[string]string `json:"traits,omitempty"`
}

type castleAPIRequest struct {
	Type         EventType         `json:"type"`
	Status       EventStatus       `json:"status"`
	RequestToken string            `json:"request_token"`
	User         User              `json:"user"`
	Context      *Context          `json:"context"`
	Properties   map[string]string `json:"properties,omitempty"`
}

type castleAPIResponse struct {
	Type    string  `json:"type"`
	Message string  `json:"message"`
	Risk    float32 `json:"risk"`
	Policy  struct {
		Name       string `json:"name"`
		ID         string `json:"id"`
		RevisionID string `json:"revision_id"`
		Action     string `json:"action"`
	} `json:"policy"`
	Device struct {
		Token string `json:"token"`
	} `json:"device"`
}
