package db

type OrganizationModel struct {
	Id      string
	Name    string
	OwnerId string
}

type OrganizationGroupModel struct {
}

func CreateOrganization(o OrganizationModel) (*OrganizationModel, error) {
	return nil, nil
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
