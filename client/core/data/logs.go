package data

import "dployr.io/pkg/models"

func (d *DataService) GetLogs(host, token, refresh, projectId string) ([]models.LogEntry, error) {
	resp, err := d.makeRequest("GET", "logs", host, token, map[string]string{"project_id": projectId}, nil)
	if err != nil {
		return nil, err
	}

	var result []models.LogEntry
	return result, d.decodeResponse(resp, &result)
}