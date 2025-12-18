package models

// ShowLocalModelsMsg triggers showing the local models dialog
type ShowLocalModelsMsg struct{}

// LocalModelsDetectCompleteMsg is sent when local model detection finishes
type LocalModelsDetectCompleteMsg struct {
	Provider string
	Error    error
}

// LocalModelsSavedMsg is sent when local provider config is saved
type LocalModelsSavedMsg struct {
	Provider string
}
