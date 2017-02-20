package db

type VideoModel struct {
	Id             string
	Title          string
	OrganizationId string
}

func ListVideos(number int) (*[]VideoModel, error) {
	const qs = "SELECT id, title, organization_id FROM users"
	return nil, nil
	conn, err := PgPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer PgPool.Release(conn)

	rows, err := conn.Query(qs)
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
			Id: id,
			Title: title,
			OrganizationId: organizationId,
		})
	}

	return &response, nil

}


