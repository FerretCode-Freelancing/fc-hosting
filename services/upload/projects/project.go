package projects

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ProjectResponse struct {
	SubscriptionId string `json:"subscription_id"`
	RamLimit       string `json:"ram_limit"`
}

func GetProject(projectId string) (ProjectResponse, error) {
	client := http.Client{}

	projects := fmt.Sprintf("%s:%s", os.Getenv("FC_PROJECTS_SERVICE_HOST"), os.Getenv("FC_PROJECTS_SERVICE_PORT"))

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/get?%s", projects, projectId),
		nil,
	)

	if err != nil {
		return ProjectResponse{}, err
	}

	res, err := client.Do(req)

	if err != nil {
		return ProjectResponse{}, err
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return ProjectResponse{}, err
	}

	projectResponse := ProjectResponse{}

	if unmarshalErr := json.Unmarshal(body, &projectResponse); unmarshalErr != nil {
		return ProjectResponse{}, unmarshalErr
	}

	return projectResponse, nil
}
