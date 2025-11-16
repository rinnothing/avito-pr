package test

import (
	"context"
	"fmt"

	"github.com/rinnothing/avito-pr/api/gen"
)

type testClient struct {
	client gen.ClientWithResponsesInterface
}

func newTestClient(port string) (*testClient, error) {
	client, err := gen.NewClientWithResponses(fmt.Sprintf("http://localhost:%s", port))
	if err != nil {
		return nil, err
	}
	return &testClient{client: client}, nil
}

func (c *testClient) AddTeam(ctx context.Context, team gen.Team) (*gen.PostTeamAddResponse, error) {
	return c.client.PostTeamAddWithResponse(ctx, team)
}

func (c *testClient) GetTeam(ctx context.Context, teamname string) (*gen.GetTeamGetResponse, error) {
	return c.client.GetTeamGetWithResponse(ctx, &gen.GetTeamGetParams{TeamName: teamname})
}

func (c *testClient) SetActive(ctx context.Context, userID string, isActive bool) (*gen.PostUsersSetIsActiveResponse, error) {
	return c.client.PostUsersSetIsActiveWithResponse(ctx, gen.PostUsersSetIsActiveJSONRequestBody{UserId: userID, IsActive: isActive})
}

func (c *testClient) AddPullRequest(ctx context.Context, pullRequestID, pullRequestName, authorID string) (*gen.PostPullRequestCreateResponse, error) {
	return c.client.PostPullRequestCreateWithResponse(ctx, gen.PostPullRequestCreateJSONRequestBody{PullRequestId: pullRequestID, PullRequestName: pullRequestName, AuthorId: authorID})
}

func (c *testClient) MergePullRequest(ctx context.Context, pullRequestID string) (*gen.PostPullRequestMergeResponse, error) {
	return c.client.PostPullRequestMergeWithResponse(ctx, gen.PostPullRequestMergeJSONRequestBody{PullRequestId: pullRequestID})
}

func (c *testClient) ReassignPullRequest(ctx context.Context, pullrequestID, oldUserID string) (*gen.PostPullRequestReassignResponse, error) {
	return c.client.PostPullRequestReassignWithResponse(ctx, gen.PostPullRequestReassignJSONRequestBody{PullRequestId: pullrequestID, OldUserId: oldUserID})
}

func (c *testClient) GetUserReviews(ctx context.Context, userID string) (*gen.GetUsersGetReviewResponse, error) {
	return c.client.GetUsersGetReviewWithResponse(ctx, &gen.GetUsersGetReviewParams{UserId: userID})
}
