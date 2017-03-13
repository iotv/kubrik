package db

type PermissionModel struct {
	Id                 string
	PermissionTypeId   string
	PermissionTypeName string
}

type GroupModel struct {
	Id          string
	Name        string
	IsPublic    bool
	Permissions []PermissionModel
}

type OrganizationModel struct {
	Id        string
	Name      string
	IsUserOrg bool
	OwnerId   string
	Groups    []GroupModel
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
	const qs = `SELECT o.name, o.is_user_org, o.owner_id,
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

	rows, err := conn.Query(qs, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	response := OrganizationModel{}
	for rows.Next() {
		var name string
		var isUserOrg bool
		var ownerId string
		var groupId string
		var groupName string
		var groupIsPublic bool
		var permissionId string
		var permissionTypeId string
		var permissionTypeName string

		err = rows.Scan(
			&name, &isUserOrg, &ownerId,
			&groupId, &groupName, &groupIsPublic,
			&permissionId, &permissionTypeId, &permissionTypeName)
		if err != nil {
			return nil, err
		}

		// TODO: maybe don't do these more than once
		response.Id = id
		response.Name = name
		response.IsUserOrg = isUserOrg
		response.OwnerId = ownerId

		// Find group if it exists
		groupExists := false
		groupIndex := 0
		for index, group := range response.Groups {
			if group.Id == groupId {
				groupExists = true
				groupIndex = index
				break
			}
		}

		// Create group if it doesn't exist
		if !groupExists && groupId != "" {
			groupIndex = len(response.Groups)
			response.Groups = append(response.Groups, GroupModel{
				Id:          groupId,
				Name:        groupName,
				IsPublic:    groupIsPublic,
				Permissions: []PermissionModel{},
			})
		}

		// Create permissions on group
		if permissionId != "" {
			response.Groups[groupIndex].Permissions = append(response.Groups[groupIndex].Permissions, PermissionModel{
				Id:                 permissionId,
				PermissionTypeId:   permissionTypeId,
				PermissionTypeName: permissionTypeName,
			})
		}
	}
	return &response, nil
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
