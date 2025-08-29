-- tools is an enum for the supported tools that speakers can use.
create table tools (
  name text primary key
) strict;

insert into tools (name) values ('save_name');

-- speakers_tools is a many-to-many mapping table between speakers and their default tools.
create table speakers_tools (
  speaker_id text not null references speakers (id) on delete cascade,
  tool_name text not null references tools (name) on delete restrict,
  primary key (speaker_id, tool_name)
) strict;