package db

type OrganizationModel struct {
	Id      string
	Name    string
	OwnerId string
}

type OrganizationGroupModel struct {
}

func CreateOrganization(name, ownerId string) (*OrganizationModel, error) {
	const qsIns = "INSERT INTO organizations(name, owner_id) VALUES($1, $2)"
	const qsSel = "SELECT id FROM organizations WHERE name=$1 AND owner_id=$2"
	var err error

	// Get a connection from the pool and set it up to release
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	// Attempt to insert the new user
	if _, err = conn.Exec(qsIns, name, ownerId); err != nil {
		return nil, err
	}

	// Attempt to find the new org's id by name and owner id
	row := conn.QueryRow(qsSel, name, ownerId)
	var id string
	if err = row.Scan(&id); err != nil {
		return nil, err
	}
	return &OrganizationModel{
		Id:      id,
		Name:    name,
		OwnerId: ownerId,
	}, nil
}

func GetOrganizationById(id string) (*OrganizationModel, error) {
	const qs = "SELECT name, owner_id FROM organizations WHERE id=$1"
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
