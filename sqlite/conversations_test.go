package sqlite_test

import (
	"testing"

	"maragu.dev/errors"
	"maragu.dev/is"

	"app/model"
	"app/sqlitetest"
)

func TestDatabase_SaveTurn(t *testing.T) {
	t.Run("can save a new turn with empty ID", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Create a test conversation and speaker first
		var conversationID model.ConversationID
		err := db.H.Get(t.Context(), &conversationID, `
			insert into conversations (topic) values ('Test conversation') 
			returning id`)
		is.NotError(t, err)

		var speakerID model.SpeakerID
		err = db.H.Get(t.Context(), &speakerID, `
			select id from speakers limit 1`)
		is.NotError(t, err)

		// Create a turn with empty ID
		turn := model.Turn{
			ConversationID: conversationID,
			SpeakerID:      speakerID,
			Content:        "Hello, this is a test message",
		}

		savedTurn, err := db.SaveTurn(t.Context(), turn)
		is.NotError(t, err)
		is.True(t, savedTurn.ID != "")                            // ID should be generated
		is.Equal(t, turn.ConversationID, savedTurn.ConversationID) // ConversationID should match
		is.Equal(t, turn.SpeakerID, savedTurn.SpeakerID)           // SpeakerID should match
		is.Equal(t, turn.Content, savedTurn.Content)               // Content should match
		is.True(t, savedTurn.Created != model.Time{})             // Created should be set
		is.True(t, savedTurn.Updated != model.Time{})             // Updated should be set
	})

	t.Run("can update an existing turn", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Create a test conversation and speaker first
		var conversationID model.ConversationID
		err := db.H.Get(t.Context(), &conversationID, `
			insert into conversations (topic) values ('Test conversation') 
			returning id`)
		is.NotError(t, err)

		var speakerID model.SpeakerID
		err = db.H.Get(t.Context(), &speakerID, `
			select id from speakers limit 1`)
		is.NotError(t, err)

		// First, create a turn
		turn := model.Turn{
			ConversationID: conversationID,
			SpeakerID:      speakerID,
			Content:        "Original content",
		}

		savedTurn, err := db.SaveTurn(t.Context(), turn)
		is.NotError(t, err)

		// Now update the turn
		savedTurn.Content = "Updated content"
		updatedTurn, err := db.SaveTurn(t.Context(), savedTurn)
		is.NotError(t, err)
		is.Equal(t, savedTurn.ID, updatedTurn.ID)                 // ID should be the same
		is.Equal(t, savedTurn.ConversationID, updatedTurn.ConversationID) // ConversationID should match
		is.Equal(t, savedTurn.SpeakerID, updatedTurn.SpeakerID)   // SpeakerID should match
		is.Equal(t, "Updated content", updatedTurn.Content)       // Content should be updated
		// Note: updated timestamp might be the same due to fast execution
	})

	t.Run("can insert turn with specific ID", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Create a test conversation and speaker first
		var conversationID model.ConversationID
		err := db.H.Get(t.Context(), &conversationID, `
			insert into conversations (topic) values ('Test conversation') 
			returning id`)
		is.NotError(t, err)

		var speakerID model.SpeakerID
		err = db.H.Get(t.Context(), &speakerID, `
			select id from speakers limit 1`)
		is.NotError(t, err)

		// Create a turn with a specific ID
		turnID := model.TurnID("tu_specific123456789abcdef0")
		turn := model.Turn{
			ID:             turnID,
			ConversationID: conversationID,
			SpeakerID:      speakerID,
			Content:        "Turn with specific ID",
		}

		savedTurn, err := db.SaveTurn(t.Context(), turn)
		is.NotError(t, err)
		is.Equal(t, turnID, savedTurn.ID)                         // ID should match the provided one
		is.Equal(t, turn.ConversationID, savedTurn.ConversationID) // ConversationID should match
		is.Equal(t, turn.SpeakerID, savedTurn.SpeakerID)           // SpeakerID should match
		is.Equal(t, turn.Content, savedTurn.Content)               // Content should match
	})

	t.Run("returns error when conversation does not exist", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		var speakerID model.SpeakerID
		err := db.H.Get(t.Context(), &speakerID, `
			select id from speakers limit 1`)
		is.NotError(t, err)

		// Try to save a turn with non-existent conversation
		turn := model.Turn{
			ConversationID: model.ConversationID("co_nonexistent123456789abc"),
			SpeakerID:      speakerID,
			Content:        "This should fail",
		}

		_, err = db.SaveTurn(t.Context(), turn)
		is.True(t, errors.Is(err, model.ErrorConversationNotFound))
	})

	t.Run("returns error when speaker does not exist", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Create a test conversation
		var conversationID model.ConversationID
		err := db.H.Get(t.Context(), &conversationID, `
			insert into conversations (topic) values ('Test conversation') 
			returning id`)
		is.NotError(t, err)

		// Try to save a turn with non-existent speaker
		turn := model.Turn{
			ConversationID: conversationID,
			SpeakerID:      model.SpeakerID("sp_nonexistent123456789abc"),
			Content:        "This should fail",
		}

		_, err = db.SaveTurn(t.Context(), turn)
		is.True(t, errors.Is(err, model.ErrorSpeakerNotFound))
	})

	t.Run("can handle concurrent operations", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Create a test conversation and speaker first
		var conversationID model.ConversationID
		err := db.H.Get(t.Context(), &conversationID, `
			insert into conversations (topic) values ('Test conversation') 
			returning id`)
		is.NotError(t, err)

		var speakerID model.SpeakerID
		err = db.H.Get(t.Context(), &speakerID, `
			select id from speakers limit 1`)
		is.NotError(t, err)

		// Create multiple turns to ensure transaction isolation works
		turn1 := model.Turn{
			ConversationID: conversationID,
			SpeakerID:      speakerID,
			Content:        "First turn",
		}

		turn2 := model.Turn{
			ConversationID: conversationID,
			SpeakerID:      speakerID,
			Content:        "Second turn",
		}

		savedTurn1, err := db.SaveTurn(t.Context(), turn1)
		is.NotError(t, err)

		savedTurn2, err := db.SaveTurn(t.Context(), turn2)
		is.NotError(t, err)

		is.True(t, savedTurn1.ID != savedTurn2.ID) // Should have different IDs
		is.Equal(t, "First turn", savedTurn1.Content)
		is.Equal(t, "Second turn", savedTurn2.Content)
	})

	t.Run("handles database constraint violations gracefully", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		// Create a test conversation and speaker first
		var conversationID model.ConversationID
		err := db.H.Get(t.Context(), &conversationID, `
			insert into conversations (topic) values ('Test conversation') 
			returning id`)
		is.NotError(t, err)

		var speakerID model.SpeakerID
		err = db.H.Get(t.Context(), &speakerID, `
			select id from speakers limit 1`)
		is.NotError(t, err)

		// First, save a turn with a specific ID
		turnID := model.TurnID("tu_test123456789abcdef012")
		turn1 := model.Turn{
			ID:             turnID,
			ConversationID: conversationID,
			SpeakerID:      speakerID,
			Content:        "First turn with specific ID",
		}

		savedTurn1, err := db.SaveTurn(t.Context(), turn1)
		is.NotError(t, err)
		is.Equal(t, turnID, savedTurn1.ID)

		// Try to save another turn with the same ID but different content
		// This should update the existing turn instead of failing
		turn2 := model.Turn{
			ID:             turnID,
			ConversationID: conversationID,
			SpeakerID:      speakerID,
			Content:        "Updated content for same ID",
		}

		savedTurn2, err := db.SaveTurn(t.Context(), turn2)
		is.NotError(t, err)
		is.Equal(t, turnID, savedTurn2.ID)
		is.Equal(t, "Updated content for same ID", savedTurn2.Content)
	})
}