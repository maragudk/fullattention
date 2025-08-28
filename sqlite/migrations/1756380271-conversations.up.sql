-- provider is an enum for the supported providers of models.
create table providers (
  name text primary key
) strict;

insert into providers (name) values ('brain'), ('llamacpp'), ('openai'), ('anthropic'), ('google'), ('fireworks');

-- models are llms.
-- They have names (how they're identified) and types (how they're communicated with),
-- as well as configuration (in JSON) which varies by type.
create table models (
  id text primary key default ('mo_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  provider text not null references providers (name) on delete restrict,
  name text not null,
  config text not null default '{}' check (json_valid(config))
) strict;

create trigger models_updated_timestamp after update on models begin
  update models set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = new.id;
end;

insert into models (id, provider, name, config) values
  ('mo_515bf0deb75982d78e99ccce48e21142', 'brain', 'human', '{"intelligence": true}'),
  ('mo_8b74dab2a7f360570be6e4898f944be3', 'openai', 'gpt-5', '{"reasoning": {"effort": "high"}}'),
  ('mo_8cc34e092637b06b9a61c3c254ef2133', 'anthropic', 'claude-opus-4-1-20250805', '{}'),
  ('mo_748b19edaa66505f81aa7725dfcd3e53', 'google', 'models/gemini-2.5-pro', '{}');

-- speakers are named models with an optional system prompt. Many speakers can use the same model.
-- Think of these as roles/personas for models.
create table speakers (
  id text primary key default ('sp_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  model_id text not null references models (id) on delete restrict,
  name text unique not null,
  system text not null default '',
  config text not null default '{}' check (json_valid(config))
) strict;

create trigger speakers_updated_timestamp after update on speakers begin
  update speakers set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = new.id;
end;

insert into speakers (id, model_id, name, config, system) values
  ('sp_71b9d186337e32d7039774e61afb956c', 'mo_515bf0deb75982d78e99ccce48e21142', 'Me', '{}', 'You do you.');

-- conversations have optional topics and tie turns together.
create table conversations (
  id text primary key default ('co_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  topic text not null default ''
) strict;

create trigger conversations_updated_timestamp after update on conversations begin
  update conversations set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = new.id;
end;

-- turns in a conversation, by a speaker.
create table turns (
  id text primary key default ('tu_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  conversation_id text not null references conversations (id) on delete cascade,
  speaker_id text not null references speakers (id) on delete restrict,
  content text not null default ''
) strict;

create trigger turns_updated_timestamp after update on turns begin
  update turns set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = new.id;
end;

create index turns_conversationID_created on turns (conversation_id, created);
