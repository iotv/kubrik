package db

type VideoModel struct {
	Id             string
	Title          string
	OrganizationId string
	VideoSegments  []VideoSegmentModel
}

type VideoSegmentModel struct {
	Id          string
	S3URL       string
	StartOffset float64
	EndOffset   float64
	Duration    float64
}


func GetVideoById(id string) (*VideoModel, error) {
	const qs = `SELECT v.title, v.organization_id,
	vs.id as segment_id, vs.s3_url as segment_s3_url,
	vs.start_offset as segment_start_offset, vs.end_offset as segment_end_offset
FROM videos v
	LEFT JOIN video_segments vs
		ON v.id = vs.video_id
WHERE v.id = $1`

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

	response := VideoModel{}
	for rows.Next() {
		var title string
		var organizationId string
		var segmentId string
		var segmentS3URL string
		var segmentStartOffset float64
		var segmentEndOffset float64
		var duration float64

		err = rows.Scan(
			&title, &organizationId,
			&segmentId, &segmentS3URL,
			&segmentStartOffset, &segmentEndOffset)
		if err != nil {
			return nil, err
		}
		duration = segmentEndOffset - segmentStartOffset

		// TODO: don't set this every time?
		response.Id = id
		response.Title = title
		response.OrganizationId = organizationId
		response.VideoSegments = append(response.VideoSegments, VideoSegmentModel{
			Id: segmentId,
			S3URL: segmentS3URL,
			StartOffset: segmentStartOffset,
			EndOffset: segmentEndOffset,
			Duration: duration,
		})

	}
	return nil, nil
}


func ListVideos(number int) (*[]VideoModel, error) {
	const qs = "SELECT id, title, organization_id FROM videos LIMIT $1"
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
