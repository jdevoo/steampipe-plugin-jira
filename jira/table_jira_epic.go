package jira

import (
	"context"
	"fmt"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableEpic(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_epic",
		Description: "An epic is essentially a large user story that can be broken down into a number of smaller stories. An epic can span more than one project.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AnyColumn([]string{"id", "key"}),
			Hydrate:    getEpic,
		},
		List: &plugin.ListConfig{
			Hydrate: listEpics,
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The id of the epic.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "The name of the epic.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "key",
				Description: "The key of the epic.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "done",
				Description: "Indicates the status of the epic.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "self",
				Description: "The URL of the epic details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "summary",
				Description: "Description of the epic.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "color",
				Description: "Label colour details for the epic.",
				Type:        proto.ColumnType_JSON,
			},

			// Standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
		},
	}
}

//// LIST FUNCTION

func listEpics(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("listEpics")

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	maxResults := 1000
	for {
		apiEndpoint := fmt.Sprintf(
			"/rest/agile/1.0/epic/search?startAt=%d&maxResults=%d",
			last,
			maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			return nil, err
		}

		listResult := new(ListEpicResult)
		_, err = client.Do(req, listResult)
		if err != nil {
			logger.Error("listEpics", "Error", err)
			return nil, err
		}

		for _, epic := range listResult.Values {
			d.StreamListItem(ctx, epic)
		}

		last = listResult.StartAt + len(listResult.Values)
		if listResult.IsLast {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTION

func getEpic(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getEpic")

	epicId := d.KeyColumnQuals["id"].GetInt64Value()
	epicKey := d.KeyColumnQuals["key"].GetStringValue()
	var apiEndpoint string

	if epicKey != "" {
		apiEndpoint = fmt.Sprintf("/rest/agile/1.0/epic/%s", epicKey)
	} else {
		apiEndpoint = fmt.Sprintf("/rest/agile/1.0/epic/%d", epicId)
	}

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	epic := new(Epic)
	_, err = client.Do(req, epic)
	if err != nil {
		if isNotFoundError(err) || strings.Contains(err.Error(), "400") {
			return nil, nil
		}
		logger.Error("getEpic", "Error", err)
		return nil, err
	}

	return epic, err
}

//// Custom Structs

type ListEpicResult struct {
	MaxResults int    `json:"maxResults"`
	StartAt    int    `json:"startAt"`
	Total      int    `json:"total"`
	IsLast     bool   `json:"isLast"`
	Values     []Epic `json:"values"`
}

type Epic struct {
	Color   Color  `json:"color"`
	Done    bool   `json:"done"`
	Id      int64  `json:"id"`
	Key     string `json:"key"`
	Name    string `json:"name"`
	Self    string `json:"self"`
	Summary string `json:"summary"`
}

type Color struct {
	Key string `json:"key"`
}