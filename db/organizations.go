package db

type OrganizationModel struct {
	Id      string
	Name    string
	OwnerId string
}

type OrganizationGroupModel struct {
}

func CreateOrganization(o OrganizationModel) (*OrganizationModel, error) {
	const qsIns = "INSERT INTO organizations(name, owner_id) VALUES($1, $2)"
	const qsSel = "SELECT id FROM organizations WHERE name=$1 AND owner_id=$2"
	var err error

	// Get a connection from the pool and set it up to release
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	// Begin a transaction and set it up to rollback by default
	tx, err := conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Attempt to insert the new user
	if _, err = tx.Exec(qsIns, o.Name, o.OwnerId); err != nil {
		return nil, err
	}

	// Attempt to find the new user's id by username and email
	row := tx.QueryRow(qsSel, o.Name, o.OwnerId)
	var id string
	if err = row.Scan(&id); err != nil {
		return nil, err
	}
	o.Id = id
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &o, nil
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
