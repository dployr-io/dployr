package data

import "dployr.io/pkg/models"

func (d *DataService) GetDeployments(host, token, refresh, projectId string) ([]models.Deployment, error) {
	resp, err := d.makeRequest("GET", "api/deployments", host, token, map[string]string{"project_id": projectId}, nil)
	if err != nil {
		return nil, err
	}

	var result []models.Deployment
	return result, d.decodeResponse(resp, &result)
}
