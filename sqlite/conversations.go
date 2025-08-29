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

// SaveTurn via upsert.
// If the turn's ID is empty, a new turn is created.
// Otherwise, the existing turn is updated.
// The conversation and speaker referenced by the turn must exist.
func (d *Database) SaveTurn(ctx context.Context, t model.Turn) (model.Turn, error) {
	err := d.H.InTx(ctx, func(ctx context.Context, tx *Tx) error {
		var conversationExists bool
		if err := tx.Get(ctx, &conversationExists, `select exists (select 1 from conversations where id = ?)`, t.ConversationID); err != nil {
			return err
		}
		if !conversationExists {
			return model.ErrorConversationNotFound
		}

		var speakerExists bool
		if err := tx.Get(ctx, &speakerExists, `select exists (select 1 from speakers where id = ?)`, t.SpeakerID); err != nil {
			return err
		}
		if !speakerExists {
			return model.ErrorSpeakerNotFound
		}

		if t.ID == "" {
			const query = `
				insert into turns (conversation_id, speaker_id, content)
				values (?, ?, ?)
				returning *`
			if err := tx.Get(ctx, &t, query, t.ConversationID, t.SpeakerID, t.Content); err != nil {
				return err
			}

			return nil
		}

		const query = `
			insert into turns (id, conversation_id, speaker_id, content)
			values (?, ?, ?, ?)
			on conflict (id) do update set
				conversation_id = excluded.conversation_id,
				speaker_id = excluded.speaker_id,
				content = excluded.content
			returning *`

		if err := tx.Get(ctx, &t, query, t.ID, t.ConversationID, t.SpeakerID, t.Content); err != nil {
			return err
		}

		return nil
	})

	return t, err
}
