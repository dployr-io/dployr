package data

import "dployr.io/pkg/models"

func (d *DataService) GetProjects(host, token string) ([]models.Project, error) {
	resp, err := d.makeRequest("GET", "api/projects", host, token, nil, nil)
	if err != nil {
		return nil, err
	}

	var result []models.Project
	return result, d.decodeResponse(resp, &result)
}

func (d *DataService) CreateProject(host, token string, payload map[string]string) (*models.Project, error) {
	resp, err := d.makeRequest("POST", "api/projects", host, token, nil, payload)
	if err != nil {
		return nil, err
	}

	var result *models.Project
	return result, d.decodeResponse(resp, &result)
}

func (d *DataService) UpdateProject(host, token string, payload map[string]interface{}) (*models.Project, error) {
	resp, err := d.makeRequest("PUT", "api/projects", host, token, nil, payload)
	if err != nil {
		return nil, err
	}

	var result *models.Project
	return result, d.decodeResponse(resp, &result)
}

func (d *DataService) DeleteProject(host, token string) (*models.MessageResponse, error) {
	resp, err := d.makeRequest("DELETE", "api/projects", host, token, nil, nil)
	if err != nil {
		return nil, err
	}

	var result *models.MessageResponse
	return result, d.decodeResponse(resp, &result)
}
