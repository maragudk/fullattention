package model

// Error is for errors in the business domain. See the constants below.
type Error string

const (
	ErrorConversationNotFound = Error("conversation not found")
	ErrorModelNotFound        = Error("model not found")
	ErrorSpeakerNotFound      = Error("speaker not found")
)

func (e Error) Error() string {
	return string(e)
}
