CREATE TABLE videos (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title           TEXT                                                 NOT NULL,
  organization_id UUID REFERENCES organizations (id) ON DELETE CASCADE NOT NULL
);


CREATE TABLE video_segments (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  s3_url       TEXT                                          NOT NULL,
  start_offset NUMERIC(3)                                    NOT NULL,
  end_offset   NUMERIC(3)                                    NOT NULL,
  video_id     UUID REFERENCES videos (id) ON DELETE CASCADE NOT NULL
);