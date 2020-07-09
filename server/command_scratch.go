package main

import (
	"fmt"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

func executeScratch(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	client := pluginapi.NewClient(p.API)

	db, err := client.Store.GetMasterDB()
	if err != nil {
		return p.responsef(cmdArgs, errors.Wrap(err, "failed to get a database connection").Error())
	}

	// qb := sq.StatementBuilderType{}

	// query := qb.
	// 	Select("COUNT(DISTINCT UserId)").
	// 	From("ChannelMemberHistory AS u").
	// 	Where(sq.Eq{"ChannelId": incidentID}).
	// 	Where(sq.Expr("u.UserId NOT IN (SELECT UserId FROM Bots)"))

	// query := qb.
	// 	Select("COUNT(Id)").
	// 	From("Posts")

	// queryStr, queryArgs, err := query.ToSql()
	// if err != nil {
	// 	return p.responsef(cmdArgs, errors.Wrap(err, "failed to build the query to retrieve all members in an incident").Error())
	// }

	// var numPosts int64
	// err = db.QueryRow(queryStr, queryArgs...).Scan(&numPosts)

	var numPosts int64
	err = db.QueryRow("SELECT COUNT(*) from Posts").Scan(&numPosts)
	if err != nil {
		return p.responsef(cmdArgs, errors.Wrap(err, "failed to query database").Error())
	}

	res := fmt.Sprintf("%d", numPosts)

	return p.responsef(cmdArgs, res)
}
