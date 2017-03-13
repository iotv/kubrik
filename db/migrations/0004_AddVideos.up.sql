CREATE TABLE videos (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title           TEXT                                                 NOT NULL,
  organization_id UUID REFERENCES organizations (id) ON DELETE CASCADE NOT NULL
);


CREATE TABLE video_segments (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  s3_url       TEXT                                          NOT NULL,
  start_offset DOUBLE PRECISION                              NOT NULL CHECK (start_offset > 0.0),
  end_offset   DOUBLE PRECISION                              NOT NULL,
  video_id     UUID REFERENCES videos (id) ON DELETE CASCADE NOT NULL,
  CHECK (end_offset > start_offset)
);
-- TODO: consider adding a constraint per video_id to prevent segments within

CREATE INDEX video_segments_start_offsets
  ON video_segments (start_offset);
CREATE INDEX video_segments_end_offsets
  ON video_segments (end_offset);