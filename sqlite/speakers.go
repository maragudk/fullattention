package sqlite

import (
	"context"
	"database/sql"

	"maragu.dev/errors"

	"app/model"
)

// SaveSpeaker via upsert.
// If the speaker's ID is empty, a new speaker is created.
// Otherwise, the existing speaker is updated.
func (d *Database) SaveSpeaker(ctx context.Context, s model.Speaker) (model.Speaker, error) {
	err := d.H.InTx(ctx, func(ctx context.Context, tx *Tx) error {
		var modelExists bool
		if err := tx.Get(ctx, &modelExists, `select exists (select 1 from models where id = ?)`, s.ModelID); err != nil {
			return err
		}
		if !modelExists {
			return model.ErrorModelNotFound
		}

		if s.ID == "" {
			const query = `
				insert into speakers (model_id, name, system, config)
				values (?, ?, ?, ?)
				returning *`
			if err := tx.Get(ctx, &s, query, s.ModelID, s.Name, s.System, s.Config); err != nil {
				return err
			}
			return nil
		}

		const query = `
			insert into speakers (id, model_id, name, system, config)
			values (?, ?, ?, ?, ?)
			on conflict (id) do update set
				model_id = excluded.model_id,
				name = excluded.name,
				system = excluded.system,
				config = excluded.config
			returning *`

		if err := tx.Get(ctx, &s, query, s.ID, s.ModelID, s.Name, s.System, s.Config); err != nil {
			return err
		}

		return nil
	})

	return s, err
}

// GetSpeakers by name.
func (d *Database) GetSpeakers(ctx context.Context) ([]model.Speaker, error) {
	var speakers []model.Speaker
	err := d.H.Select(ctx, &speakers, "select * from speakers order by name")
	if err != nil {
		return speakers, err
	}

	// Populate tools for each speaker
	for i := range speakers {
		tools, err := d.getSpeakerTools(ctx, speakers[i].ID)
		if err != nil {
			return speakers, err
		}
		speakers[i].Tools = tools
	}

	return speakers, nil
}

// GetSpeaker by ID or name.
// ID has precedence over name.
func (d *Database) GetSpeaker(ctx context.Context, f model.GetSpeakerFilter) (model.Speaker, error) {
	var s model.Speaker
	var err error

	if f.ID == "" && f.Name == "" {
		panic("either ID or name must be set to get speaker")
	}

	if f.ID != "" {
		err = d.H.Get(ctx, &s, "select * from speakers where id = ?", f.ID)
	} else {
		err = d.H.Get(ctx, &s, "select * from speakers where name = ?", f.Name)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return s, model.ErrorSpeakerNotFound
	}
	if err != nil {
		return s, err
	}

	// Populate tools for the speaker
	tools, err := d.getSpeakerTools(ctx, s.ID)
	if err != nil {
		return s, err
	}
	s.Tools = tools

	return s, nil
}

// getSpeakerTools returns the tools associated with a speaker.
func (d *Database) getSpeakerTools(ctx context.Context, speakerID model.SpeakerID) ([]string, error) {
	var tools []string
	err := d.H.Select(ctx, &tools, "select tool_name from speakers_tools where speaker_id = ? order by tool_name", speakerID)
	return tools, err
}
