package model

import (
	"encoding/json"
	"fmt"
)

type JSON string

type Provider string

const (
	ProviderAnthropic = Provider("anthropic")
	ProviderBrain     = Provider("brain")
	ProviderFireworks = Provider("fireworks")
	ProviderGoogle    = Provider("google")
	ProviderLlamaCPP  = Provider("llamacpp")
	ProviderOpenAI    = Provider("openai")
)

type ModelID ID

func (i ModelID) String() string {
	return string(i)
}

var _ fmt.Stringer = ModelID("")

type Model struct {
	ID       ModelID
	Created  Time
	Updated  Time
	Provider Provider
	Name     string
	Config   JSON
}

func (m Model) URL() string {
	config := unmarshalConfig(m.Config)

	switch m.Provider {
	case ProviderAnthropic, ProviderGoogle, ProviderOpenAI:
		return ""
	case ProviderFireworks:
		return "https://api.fireworks.ai/inference/v1"
	case ProviderLlamaCPP:
		return fmt.Sprintf("http://%v/v1", config["address"])
	default:
		panic("unsupported model type")
	}
}

type SpeakerID ID

func (i SpeakerID) String() string {
	return string(i)
}

var _ fmt.Stringer = SpeakerID("")

type Speaker struct {
	ID      SpeakerID
	Created Time
	Updated Time
	ModelID ModelID `db:"model_id"`
	Name    string
	System  string
	Config  JSON
	Tools   []string `db:"-"`
}

type ConversationID ID

func (c ConversationID) String() string {
	return string(c)
}

var _ fmt.Stringer = ConversationID("")

type Conversation struct {
	ID      ConversationID
	Created Time
	Updated Time
	Topic   string
}

type TurnID ID

func (i TurnID) String() string {
	return string(i)
}

var _ fmt.Stringer = TurnID("")

type Turn struct {
	ID             TurnID
	Created        Time
	Updated        Time
	ConversationID ConversationID `db:"conversation_id"`
	SpeakerID      SpeakerID      `db:"speaker_id"`
	Content        string
}

type ConversationDocument struct {
	Conversation Conversation
	Speakers     map[SpeakerID]Speaker
	Turns        []Turn
}

func unmarshalConfig(s JSON) map[string]any {
	config := map[string]any{}
	if err := json.Unmarshal([]byte(s), &config); err != nil {
		panic(err)
	}
	return config
}
