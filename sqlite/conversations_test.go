package sqlite_test

import (
	"testing"

	"maragu.dev/is"

	"app/model"
	"app/sqlitetest"
)

func TestDatabase_SaveTurn(t *testing.T) {
	t.Run("should save a new turn successfully", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		var conversationID model.ConversationID
		err := db.H.Get(t.Context(), &conversationID, `
			insert into conversations (topic) values ('Test topic')
			returning id`)
		is.NotError(t, err)

		// Create a speaker
		var speakerID model.SpeakerID
		err = db.H.Get(t.Context(), &speakerID, `
			insert into speakers (model_id, name) values ('mo_515bf0deb75982d78e99ccce48e21142', 'Test Speaker')
			returning id`)
		is.NotError(t, err)

		savedTurn, err := db.SaveTurn(t.Context(), model.Turn{
			ConversationID: conversationID,
			SpeakerID:      speakerID,
			Content:        "Hello, world!",
		})
		is.NotError(t, err)
		is.True(t, savedTurn.ID != "")
		is.Equal(t, conversationID, savedTurn.ConversationID)
		is.Equal(t, speakerID, savedTurn.SpeakerID)
		is.Equal(t, "Hello, world!", savedTurn.Content)
		is.True(t, !savedTurn.Created.T.IsZero())
		is.True(t, !savedTurn.Updated.T.IsZero())
	})

	t.Run("should upsert an existing turn successfully", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Create a conversation
		var conversationID model.ConversationID
		err := db.H.Get(t.Context(), &conversationID, `
			insert into conversations (topic) values ('Test topic')
			returning id`)
		is.NotError(t, err)

		// Create a speaker
		var speakerID model.SpeakerID
		err = db.H.Get(t.Context(), &speakerID, `
			insert into speakers (model_id, name) values ('mo_515bf0deb75982d78e99ccce48e21142', 'Test Speaker')
			returning id`)
		is.NotError(t, err)

		savedTurn, err := db.SaveTurn(t.Context(), model.Turn{
			ConversationID: conversationID,
			SpeakerID:      speakerID,
			Content:        "Initial content",
		})
		is.NotError(t, err)

		// Upsert the same turn with updated content
		updatedTurn := model.Turn{
			ID:             savedTurn.ID,
			ConversationID: conversationID,
			SpeakerID:      speakerID,
			Content:        "Updated content",
		}

		upsertedTurn, err := db.SaveTurn(t.Context(), updatedTurn)
		is.NotError(t, err)
		is.Equal(t, savedTurn.ID, upsertedTurn.ID)
		is.Equal(t, "Updated content", upsertedTurn.Content)
	})

	t.Run("should return error when conversation does not exist", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Create a speaker
		var speakerID model.SpeakerID
		err := db.H.Get(t.Context(), &speakerID, `
			insert into speakers (model_id, name) values ('mo_515bf0deb75982d78e99ccce48e21142', 'Test Speaker')
			returning id`)
		is.NotError(t, err)

		// Try to save a turn with non-existent conversation
		_, err = db.SaveTurn(t.Context(), model.Turn{
			ConversationID: "co_nonexistent",
			SpeakerID:      speakerID,
			Content:        "Hello, world!",
		})
		is.Error(t, model.ErrorConversationNotFound, err)
	})

	t.Run("should return error when speaker does not exist", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Create a conversation
		var conversationID model.ConversationID
		err := db.H.Get(t.Context(), &conversationID, `
			insert into conversations (topic) values ('Test topic')
			returning id`)
		is.NotError(t, err)

		// Try to save a turn with non-existent speaker
		_, err = db.SaveTurn(t.Context(), model.Turn{
			ConversationID: conversationID,
			SpeakerID:      "sp_nonexistent",
			Content:        "Hello, world!",
		})
		is.Error(t, model.ErrorSpeakerNotFound, err)
	})
}
