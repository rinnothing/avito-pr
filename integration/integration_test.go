package test

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"testing"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/config"
	"github.com/stretchr/testify/require"
)

var cfg *config.Config

func TestMain(m *testing.M) {
	var err error
	cfg, err = config.New("../config/prod.yaml")
	if err != nil {
		log.Fatalf("can't read config: %s", err.Error())
	}

	m.Run()
}

func randomName(pref string) string {
	return fmt.Sprintf("%s-%d", pref, rand.Int32N(1000))
}

func randomNameID(pref1, pref2 string) (string, string) {
	val := rand.Int32N(1000)
	return fmt.Sprintf("%s-%d", pref1, val), fmt.Sprintf("%s-%d", pref2, val)
}

func randomTeam(pref string, size int) (gen.Team, []string) {
	members := make([]gen.TeamMember, 0, size)
	membersIDs := make([]string, 0, size)
	for i := 0; i < size; i++ {
		id, name := randomNameID("id", "name")
		members = append(members, gen.TeamMember{IsActive: true, UserId: id, Username: name})
		membersIDs = append(membersIDs, id)
	}

	return gen.Team{TeamName: randomName("team-" + pref), Members: members}, membersIDs
}

func randomPR(pref string) (string, string) {
	return randomNameID("id-"+pref, "name-"+pref)
}

func TestHappyPath(t *testing.T) {
	ctx := t.Context()

	client, err := newTestClient(cfg.HTTP.Port)
	require.NoError(t, err)

	// add team
	fstTeam, fstIDs := randomTeam("fst", 4)
	fstResp, err := client.AddTeam(ctx, fstTeam)
	require.NoError(t, err)

	require.Equal(t, fstResp.StatusCode(), http.StatusCreated)
	require.Equal(t, fstTeam, *fstResp.JSON201.Team)

	// get team
	fstGetTeam, err := client.GetTeam(ctx, fstTeam.TeamName)
	require.NoError(t, err)

	require.Equal(t, fstGetTeam.StatusCode(), http.StatusOK)
	require.Equal(t, fstTeam, *fstGetTeam.JSON200)

	// add pr
	prId, prName := randomPR("pr")
	prResp, err := client.AddPullRequest(ctx, prId, prName, fstIDs[0])
	require.NoError(t, err)

	require.Equal(t, prResp.StatusCode(), http.StatusCreated)
	fstAssigned := prResp.JSON201.Pr.AssignedReviewers
	require.Contains(t, fstIDs, fstAssigned[0])
	require.Contains(t, fstIDs, fstAssigned[1])

	// reassign pr
	userToRemove := fstAssigned[0]
	prReassResp, err := client.ReassignPullRequest(ctx, prId, userToRemove)
	require.NoError(t, err)

	require.Equal(t, prReassResp.StatusCode(), http.StatusOK)
	require.NotContains(t, prReassResp.JSON200.Pr.AssignedReviewers, userToRemove)

	// list reviews
	revs, err := client.GetUserReviews(ctx, fstAssigned[1])
	require.NoError(t, err)

	require.Equal(t, revs.StatusCode(), http.StatusOK)

	require.Equal(t, revs.JSON200.PullRequests[0].AuthorId, fstIDs[0])
	require.Equal(t, revs.JSON200.PullRequests[0].PullRequestId, prId)
	require.Equal(t, revs.JSON200.PullRequests[0].PullRequestName, prName)
	require.Equal(t, revs.JSON200.PullRequests[0].Status, gen.PullRequestShortStatus("OPEN"))

	// merge pr
	merged, err := client.MergePullRequest(ctx, prId)
	require.NoError(t, err)

	require.Equal(t, merged.StatusCode(), http.StatusOK)
	require.Equal(t, merged.JSON200.Pr.Status, gen.PullRequestStatus("MERGED"))
}
