package db

type OrganizationModel struct {
	Id        string
	Name      string
	IsUserOrg bool
	OwnerId   string
}

type OrganizationGroupModel struct {
}

func CreateOrganization(name, ownerId string, isUserOrg bool) (*OrganizationModel, error) {
	const qsIns = "INSERT INTO organizations(name, owner_id, is_user_org) VALUES($1, $2, $3) RETURNING id"
	var err error

	// Get a connection from the pool and set it up to release
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	// Attempt to insert the new user
	row := conn.QueryRow(qsIns, name, ownerId, isUserOrg)
	var id string
	if err = row.Scan(&id); err != nil {
		return nil, err
	}
	return &OrganizationModel{
		Id:        id,
		Name:      name,
		IsUserOrg: isUserOrg,
		OwnerId:   ownerId,
	}, nil
}

func GetOrganizationById(id string) (*OrganizationModel, error) {
	const qs = "SELECT name, owner_id FROM organizations WHERE id=$1"
	const q = `SELECT o.id, o.name, o.is_user_org, o.owner_id,
  g.id as group_id, g.name as group_name, g.is_public as group_is_public,
  p.id as permission_id, p.permission_type_id,
  t.name as permission_type_name
FROM organizations o
  LEFT JOIN organization_groups g
    ON o.id = g.organization_id
  LEFT JOIN organization_group_permissions p
    ON g.id = p.group_id
  LEFT JOIN organization_group_permission_types t
    ON p.permission_type_id = t.id
WHERE o.id = $1`
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	var name string
	var ownerId string
	row := conn.QueryRow(qs, id)
	err = row.Scan(&name, &ownerId)
	if err != nil {
		return nil, err
	}
	return &OrganizationModel{
		Id:      id,
		Name:    name,
		OwnerId: ownerId,
	}, nil
}

func GetOrganizationByName(name string) (*OrganizationModel, error) {
	const qs = "SELECT id, owner_id FROM organizations WHERE name=$1"
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	var id string
	var ownerId string
	row := conn.QueryRow(qs, name)
	err = row.Scan(&id, &ownerId)
	if err != nil {
		return nil, err
	}
	return &OrganizationModel{
		Id:      id,
		Name:    name,
		OwnerId: ownerId,
	}, nil
}
