package sqlite

import (
	"context"
	"database/sql"

	"maragu.dev/errors"

	"app/model"
)

func (d *Database) GetLatestConversation(ctx context.Context) (model.Conversation, error) {
	var c model.Conversation
	err := d.H.Get(ctx, &c, "select * from conversations order by created desc limit 1")
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return c, model.ErrorConversationNotFound
	}
	return c, err
}

func (d *Database) GetConversationDocument(ctx context.Context, id model.ConversationID) (model.ConversationDocument, error) {
	var cd model.ConversationDocument
	cd.Speakers = map[model.SpeakerID]model.Speaker{}

	err := d.H.InTx(ctx, func(ctx context.Context, tx *Tx) error {
		if err := tx.Get(ctx, &cd.Conversation, `select * from conversations where id = ?`, id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.ErrorConversationNotFound
			}
			return err
		}
		if err := tx.Select(ctx, &cd.Turns, `select * from turns where conversation_id = ? order by created`, id); err != nil {
			return err
		}
		for _, t := range cd.Turns {
			s, ok := cd.Speakers[t.SpeakerID]
			if ok {
				continue
			}

			if err := tx.Get(ctx, &s, `select * from speakers where id = ?`, t.SpeakerID); err != nil {
				return err
			}
			cd.Speakers[s.ID] = s
		}

		return nil
	})

	return cd, err
}

func (d *Database) GetConversations(ctx context.Context) ([]model.Conversation, error) {
	var cs []model.Conversation
	err := d.H.Select(ctx, &cs, "select * from conversations order by created desc")
	return cs, err
}

// SaveTurn saves a turn to the database. If the turn's ID is empty, a new turn is created.
// Otherwise, the existing turn is updated (upserted). Before saving, it ensures both the
// conversation and speaker referenced by the turn exist in the database.
// All operations are performed in a single transaction for consistency.
func (d *Database) SaveTurn(ctx context.Context, turn model.Turn) (model.Turn, error) {
	var savedTurn model.Turn

	err := d.H.InTx(ctx, func(ctx context.Context, tx *Tx) error {
		// Validate that the conversation exists
		var conversationExists bool
		err := tx.Get(ctx, &conversationExists, `select 1 from conversations where id = ?`, turn.ConversationID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.ErrorConversationNotFound
			}
			return err
		}

		// Validate that the speaker exists
		var speakerExists bool
		err = tx.Get(ctx, &speakerExists, `select 1 from speakers where id = ?`, turn.SpeakerID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.ErrorSpeakerNotFound
			}
			return err
		}

		if turn.ID == "" {
			// Insert new turn - let database generate the ID
			err = tx.Exec(ctx, `
				insert into turns (conversation_id, speaker_id, content) 
				values (?, ?, ?)`,
				turn.ConversationID, turn.SpeakerID, turn.Content)
			if err != nil {
				return err
			}

			// Get the newly created turn
			err = tx.Get(ctx, &savedTurn, `
				select * from turns 
				where conversation_id = ? and speaker_id = ? and content = ? 
				order by created desc limit 1`,
				turn.ConversationID, turn.SpeakerID, turn.Content)
			if err != nil {
				return err
			}
		} else {
			// Check if turn exists
			var existingTurn model.Turn
			err = tx.Get(ctx, &existingTurn, `select * from turns where id = ?`, turn.ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}

			if errors.Is(err, sql.ErrNoRows) {
				// Turn doesn't exist, insert with the provided ID
				err = tx.Exec(ctx, `
					insert into turns (id, conversation_id, speaker_id, content) 
					values (?, ?, ?, ?)`,
					turn.ID, turn.ConversationID, turn.SpeakerID, turn.Content)
				if err != nil {
					return err
				}
			} else {
				// Turn exists, update it
				err = tx.Exec(ctx, `
					update turns 
					set conversation_id = ?, speaker_id = ?, content = ? 
					where id = ?`,
					turn.ConversationID, turn.SpeakerID, turn.Content, turn.ID)
				if err != nil {
					return err
				}
			}

			// Get the saved turn
			err = tx.Get(ctx, &savedTurn, `select * from turns where id = ?`, turn.ID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return savedTurn, err
}
