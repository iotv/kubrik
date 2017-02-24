package db

type VideoModel struct {
	Id             string
	Title          string
	OrganizationId string
}

func ListVideos(number int) (*[]VideoModel, error) {
	const qs = "SELECT id, title, organization_id FROM users LIMIT $1"
	return nil, nil
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	rows, err := conn.Query(qs, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var response []VideoModel
	for rows.Next() {
		var id string
		var title string
		var organizationId string
		err = rows.Scan(&id, &title, &organizationId)
		if err != nil {
			return nil, err
		}

		response = append(response, VideoModel{
			Id:             id,
			Title:          title,
			OrganizationId: organizationId,
		})
	}

	return &response, nil

}

func CreateVideo(title, organizationId string) (*VideoModel, error) {
	const qsIns = "INSERT INTO videos(title, organization_id) VALUES($1, $2) RETURNING id"
	var err error

	// Get a connection from the pool and set it up to release
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	// Attempt to insert the new user
	row := conn.QueryRow(qsIns, title, organizationId)
	var id string
	if err = row.Scan(&id); err != nil {
		return nil, err
	}

	return &VideoModel{
		Id:             id,
		Title:          title,
		OrganizationId: organizationId,
	}, nil
}
