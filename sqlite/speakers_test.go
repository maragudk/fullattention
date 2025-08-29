package sqlite_test

import (
	"testing"

	"maragu.dev/is"

	"app/model"
	"app/sqlitetest"
)

var (
	modelGPT5       = model.ModelID("mo_8b74dab2a7f360570be6e4898f944be3")
	modelClaudeOpus = model.ModelID("mo_8cc34e092637b06b9a61c3c254ef2133")
	modelGemini     = model.ModelID("mo_748b19edaa66505f81aa7725dfcd3e53")
)

func TestDatabase_SaveSpeaker(t *testing.T) {
	t.Run("should save a new speaker successfully when ID is empty", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		savedSpeaker, err := db.SaveSpeaker(t.Context(), model.Speaker{
			ModelID: modelGPT5,
			Name:    "Test Assistant",
			System:  "You are a helpful assistant",
			Config:  `{"temperature": 0.7}`,
		})
		is.NotError(t, err)
		is.True(t, savedSpeaker.ID != "")
		is.Equal(t, modelGPT5, savedSpeaker.ModelID)
		is.Equal(t, "Test Assistant", savedSpeaker.Name)
		is.Equal(t, "You are a helpful assistant", savedSpeaker.System)
		is.Equal(t, model.JSON(`{"temperature": 0.7}`), savedSpeaker.Config)
		is.True(t, !savedSpeaker.Created.T.IsZero())
		is.True(t, !savedSpeaker.Updated.T.IsZero())
	})

	t.Run("should upsert an existing speaker successfully", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		savedSpeaker, err := db.SaveSpeaker(t.Context(), model.Speaker{
			ModelID: modelClaudeOpus,
			Name:    "Initial Name",
			System:  "Initial system prompt",
			Config:  `{"temperature": 0.5}`,
		})
		is.NotError(t, err)

		updatedSpeaker := model.Speaker{
			ID:      savedSpeaker.ID,
			ModelID: modelClaudeOpus,
			Name:    "Updated Name",
			System:  "Updated system prompt",
			Config:  `{"temperature": 0.9}`,
		}

		upsertedSpeaker, err := db.SaveSpeaker(t.Context(), updatedSpeaker)
		is.NotError(t, err)
		is.Equal(t, updatedSpeaker.ID, upsertedSpeaker.ID)
		is.Equal(t, updatedSpeaker.Name, upsertedSpeaker.Name)
		is.Equal(t, updatedSpeaker.System, upsertedSpeaker.System)
		is.Equal(t, updatedSpeaker.Config, upsertedSpeaker.Config)
	})

	t.Run("should return error when model does not exist", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		var initialCount int
		err := db.H.Get(t.Context(), &initialCount, `select count(*) from speakers`)
		is.NotError(t, err)

		speaker := model.Speaker{
			ModelID: "mo_nonexistent",
			Name:    "Test Speaker",
			System:  "System prompt",
			Config:  `{}`,
		}

		_, err = db.SaveSpeaker(t.Context(), speaker)
		is.Error(t, model.ErrorModelNotFound, err)

		// Verify no new speaker was saved
		var finalCount int
		err = db.H.Get(t.Context(), &finalCount, `select count(*) from speakers`)
		is.NotError(t, err)
		is.Equal(t, initialCount, finalCount)

		// Also verify the specific speaker name doesn't exist
		var exists bool
		err = db.H.Get(t.Context(), &exists, `select exists(select 1 from speakers where name = 'Test Speaker')`)
		is.NotError(t, err)
		is.True(t, !exists)
	})
}

func TestDatabase_GetSpeakers(t *testing.T) {
	t.Run("should return all speakers by name", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		speakers, err := db.GetSpeakers(t.Context())
		is.NotError(t, err)
		is.Equal(t, 2, len(speakers))

		is.Equal(t, "Me", speakers[0].Name)
		is.Equal(t, "The Caretaker", speakers[1].Name)
	})

	t.Run("should return speakers with empty tools by default", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		speakers, err := db.GetSpeakers(t.Context())
		is.NotError(t, err)
		is.Equal(t, 2, len(speakers))

		for _, speaker := range speakers {
			is.Equal(t, 0, len(speaker.Tools))
		}
	})

	t.Run("should return speakers with tools when assigned", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Create a speaker
		savedSpeaker, err := db.SaveSpeaker(t.Context(), model.Speaker{
			ModelID: modelGPT5,
			Name:    "Tool Speaker",
			System:  "Test system",
			Config:  `{}`,
		})
		is.NotError(t, err)

		// Assign the save_name tool to the speaker
		_, err = db.H.Exec(t.Context(), "insert into speakers_tools (speaker_id, tool_name) values (?, ?)", 
			savedSpeaker.ID, "save_name")
		is.NotError(t, err)

		speakers, err := db.GetSpeakers(t.Context())
		is.NotError(t, err)

		// Find our speaker in the results
		var toolSpeaker *model.Speaker
		for i, speaker := range speakers {
			if speaker.Name == "Tool Speaker" {
				toolSpeaker = &speakers[i]
				break
			}
		}

		is.True(t, toolSpeaker != nil)
		is.Equal(t, 1, len(toolSpeaker.Tools))
		is.Equal(t, "save_name", toolSpeaker.Tools[0])
	})
}

func TestDatabase_GetSpeaker(t *testing.T) {
	t.Run("should get speaker by ID", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		savedSpeaker, err := db.SaveSpeaker(t.Context(), model.Speaker{
			ModelID: modelClaudeOpus,
			Name:    "Test Speaker",
			System:  "Test system",
			Config:  `{"test": true}`,
		})
		is.NotError(t, err)

		retrievedSpeaker, err := db.GetSpeaker(t.Context(), model.GetSpeakerFilter{ID: savedSpeaker.ID})
		is.NotError(t, err)
		is.Equal(t, savedSpeaker.ID, retrievedSpeaker.ID)
	})

	t.Run("should get speaker by name", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		savedSpeaker, err := db.SaveSpeaker(t.Context(), model.Speaker{
			ModelID: modelGPT5,
			Name:    "Unique Name",
			System:  "Test system",
			Config:  `{}`,
		})
		is.NotError(t, err)

		retrievedSpeaker, err := db.GetSpeaker(t.Context(), model.GetSpeakerFilter{Name: "Unique Name"})
		is.NotError(t, err)
		is.Equal(t, savedSpeaker.ID, retrievedSpeaker.ID)
		is.Equal(t, savedSpeaker.Name, retrievedSpeaker.Name)
	})

	t.Run("should prefer ID over name when both are set", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		savedSpeaker1, err := db.SaveSpeaker(t.Context(), model.Speaker{
			ModelID: modelGemini,
			Name:    "Speaker One",
			System:  "System 1",
			Config:  `{}`,
		})
		is.NotError(t, err)

		_, err = db.SaveSpeaker(t.Context(), model.Speaker{
			ModelID: modelGemini,
			Name:    "Speaker Two",
			System:  "System 2",
			Config:  `{}`,
		})
		is.NotError(t, err)

		// Get speaker with both ID and name set (ID should take precedence)
		retrievedSpeaker, err := db.GetSpeaker(t.Context(), model.GetSpeakerFilter{
			ID:   savedSpeaker1.ID,
			Name: "Speaker Two", // Different name
		})
		is.NotError(t, err)
		is.Equal(t, savedSpeaker1.ID, retrievedSpeaker.ID)
	})

	t.Run("should return ErrorSpeakerNotFound when speaker does not exist by ID", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		_, err := db.GetSpeaker(t.Context(), model.GetSpeakerFilter{ID: "sp_nonexistent"})
		is.Error(t, model.ErrorSpeakerNotFound, err)
	})

	t.Run("should return ErrorSpeakerNotFound when speaker does not exist by name", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		_, err := db.GetSpeaker(t.Context(), model.GetSpeakerFilter{Name: "Nonexistent Speaker"})
		is.Error(t, model.ErrorSpeakerNotFound, err)
	})

	t.Run("should panic when neither ID nor name is set", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		defer func() {
			r := recover()
			is.True(t, r != nil)
			is.Equal(t, "either ID or name must be set to get speaker", r)
		}()

		filter := model.GetSpeakerFilter{}
		_, _ = db.GetSpeaker(t.Context(), filter)
	})

	t.Run("should get speaker with empty tools by default", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		savedSpeaker, err := db.SaveSpeaker(t.Context(), model.Speaker{
			ModelID: modelClaudeOpus,
			Name:    "Test Speaker",
			System:  "Test system",
			Config:  `{"test": true}`,
		})
		is.NotError(t, err)

		retrievedSpeaker, err := db.GetSpeaker(t.Context(), model.GetSpeakerFilter{ID: savedSpeaker.ID})
		is.NotError(t, err)
		is.Equal(t, savedSpeaker.ID, retrievedSpeaker.ID)
		is.Equal(t, 0, len(retrievedSpeaker.Tools))
	})

	t.Run("should get speaker with tools when assigned", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		savedSpeaker, err := db.SaveSpeaker(t.Context(), model.Speaker{
			ModelID: modelGPT5,
			Name:    "Tool Test Speaker",
			System:  "Test system",
			Config:  `{}`,
		})
		is.NotError(t, err)

		// Assign the save_name tool to the speaker
		_, err = db.H.Exec(t.Context(), "insert into speakers_tools (speaker_id, tool_name) values (?, ?)", 
			savedSpeaker.ID, "save_name")
		is.NotError(t, err)

		retrievedSpeaker, err := db.GetSpeaker(t.Context(), model.GetSpeakerFilter{ID: savedSpeaker.ID})
		is.NotError(t, err)
		is.Equal(t, savedSpeaker.ID, retrievedSpeaker.ID)
		is.Equal(t, 1, len(retrievedSpeaker.Tools))
		is.Equal(t, "save_name", retrievedSpeaker.Tools[0])
	})

	t.Run("should get speaker with multiple tools in sorted order", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Add a test tool for this test
		_, err := db.H.Exec(t.Context(), "insert into tools (name) values (?)", "test_tool")
		is.NotError(t, err)

		savedSpeaker, err := db.SaveSpeaker(t.Context(), model.Speaker{
			ModelID: modelGemini,
			Name:    "Multi Tool Speaker",
			System:  "Test system",
			Config:  `{}`,
		})
		is.NotError(t, err)

		// Note: Inserting in reverse alphabetical order to test sorting
		_, err = db.H.Exec(t.Context(), "insert into speakers_tools (speaker_id, tool_name) values (?, ?)", 
			savedSpeaker.ID, "test_tool")
		is.NotError(t, err)
		_, err = db.H.Exec(t.Context(), "insert into speakers_tools (speaker_id, tool_name) values (?, ?)", 
			savedSpeaker.ID, "save_name")
		is.NotError(t, err)

		retrievedSpeaker, err := db.GetSpeaker(t.Context(), model.GetSpeakerFilter{ID: savedSpeaker.ID})
		is.NotError(t, err)
		is.Equal(t, savedSpeaker.ID, retrievedSpeaker.ID)
		is.Equal(t, 2, len(retrievedSpeaker.Tools))
		is.Equal(t, "save_name", retrievedSpeaker.Tools[0])  // Alphabetically first
		is.Equal(t, "test_tool", retrievedSpeaker.Tools[1]) // Alphabetically second
	})
}
