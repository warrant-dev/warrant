package adaptor

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	grpcClients "github.com/warrant-dev/warrant/pkg/grpc/client"
	"github.com/warrant-dev/warrant/pkg/grpc/pb"
	"google.golang.org/protobuf/proto"
)

func GetUserIds(userId string, includeOrgId bool, includeImGroupsId bool) (orgId string, imGroupIds []string, err error) {
	if includeOrgId {
		queryUserOrgRequest := &pb.QueryUserOrgRequest{
			Uid: &userId,
		}
		resp, err := grpcClients.InternalOrgServiceClient.QueryUserOrg(context.Background(), queryUserOrgRequest)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to query user org, userId: %s", userId)
			return "", nil, err
		}
		if resp.GetStatus() != 0 || resp.Data == nil {
			log.Error().Err(err).Msgf("Failed to query user org, userId: %s,status:%d,reason:%s", userId, resp.GetStatus(), resp.GetReason())
			return "", nil, errors.New("Failed to query user org, userId: " + userId)
		}
		orgAny := resp.GetData()
		orgResponse := pb.QueryUserOrgResponse{}
		err = proto.Unmarshal(orgAny.Value, &orgResponse)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to unmarshal user org, userId: %s", userId)
			return "", nil, err
		}
		if orgResponse.GetOrg() == nil {
			log.Error().Msgf("User org not found, userId: %s", userId)
			return "", nil, errors.New("User org not found, userId: " + userId)
		} else {
			orgId = orgResponse.GetOrg().GetId()
		}
	}
	if includeImGroupsId {
		memberGroupRequest := &pb.MemberGroupRequest{
			Number: &userId,
		}
		resp, err := grpcClients.GroupServiceClient.GetMemberGroups(context.Background(), memberGroupRequest)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get groups for user, userId: %s", userId)
			return orgId, nil, err
		}
		if resp.GetStatus() != 0 || resp.Data == nil {
			log.Error().Err(err).Msgf("Failed to get groups for user, userId: %s,status:%d,reason:%s", userId, resp.GetStatus(), resp.GetReason())
			return "", nil, errors.New("Failed to get groups for user, userId: " + userId)
		}
		groupsAny := resp.GetData()
		baseAnyResp := pb.BaseAnyResponse{}
		err = proto.Unmarshal(groupsAny.Value, &baseAnyResp)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to unmarshal groups for user, userId: %s", userId)
			return orgId, nil, err
		}
		err = json.Unmarshal([]byte(baseAnyResp.GetValue()), &imGroupIds)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to unmarshal groups for user, userId: %s", userId)
			return orgId, nil, err
		}
	}
	return orgId, imGroupIds, nil
}
