CREATE TABLE IF NOT EXISTS organizations (
  id       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name     VARCHAR(31) UNIQUE                                  NOT NULL,
  owner_id UUID REFERENCES users (id) ON DELETE CASCADE        NOT NULL
);


CREATE TABLE IF NOT EXISTS organization_groups (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name            VARCHAR(31)                                          NOT NULL,
  organization_id UUID REFERENCES organizations (id) ON DELETE CASCADE NOT NULL,
  UNIQUE (name, organization_id)
);


CREATE TABLE IF NOT EXISTS organization_group_users (
  id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id               UUID REFERENCES users (id) ON DELETE CASCADE               NOT NULL,
  organization_group_id UUID REFERENCES organization_groups (id) ON DELETE CASCADE NOT NULL,
  UNIQUE (user_id, organization_group_id)
);


CREATE TABLE IF NOT EXISTS organization_group_permission_types (
  id   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR(31)
);


CREATE TABLE IF NOT EXISTS organization_group_permissions (
  if                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  group_id           UUID REFERENCES organization_groups (id) ON DELETE CASCADE                 NOT NULL,
  permission_type_id UUID REFERENCES organization_group_permission_types (id) ON DELETE CASCADE NOT NULL,
  UNIQUE (group_id, permission_type_id)
);