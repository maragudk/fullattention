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
