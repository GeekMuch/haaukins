package store_test

import (
	"testing"

	"github.com/aau-network-security/go-ntp/store"
	"github.com/google/uuid"
)

func TestNewTeam(t *testing.T) {
	password := "some_password"
	team := store.NewTeam("some name", "some@email.com", password)

	if team.HashedPassword == password {
		t.Fatalf("expected password to be hashed")
	}
}

func TestTeamSolveTask(t *testing.T) {
	etag, err := store.NewTag("abc")
	if err != nil {
		t.Fatalf("invalid tag: %s", err)
	}
	team := store.NewTeam("some name", "some@email.com", "some_password", []store.Task{
		{FlagTag: store.Tag(etag)},
	}...)

	if err := team.SolveTaskByTag(etag); err != nil {
		t.Fatalf("expected no error when solving task for team: %s", err)
	}

	if team.Tasks[0].CompletedAt == nil {
		t.Fatalf("expected completed at to be non nil when completed")
	}

	if err := team.SolveTaskByTag("unknown-tag"); err == nil {
		t.Fatalf("expected error when solving unknown task")
	}
}

func TestCreateToken(t *testing.T) {
	tt := []struct {
		name  string
		team  *store.Team
		token string
		err   string
	}{
		{name: "Normal", team: &store.Team{Email: "tkp@tkp.dk"}, token: uuid.New().String()},
		{name: "Empty token", team: &store.Team{Email: "tkp@tkp.dk"}, token: "", err: "Token cannot be empty"},
		{name: "Unknown team", token: uuid.New().String(), err: "Unknown team"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := store.NewTeamStore()

			var team store.Team
			if tc.team != nil {
				team = *tc.team
				if err := ts.CreateTeam(team); err != nil {
					t.Fatalf("expected no error when creating team")
				}
			}

			err := ts.CreateTokenForTeam(tc.token, team)
			if err != nil {
				if tc.err != "" {
					if tc.err != err.Error() {
						t.Fatalf("unexpected error (expected: \"%s\") when creating token: %s", tc.err, err)
					}

					return
				}

				t.Fatalf("received error when creating token, but expected none: %s", err)
			}

			if tc.err != "" {
				t.Fatalf("expected error but received none: %s", tc.err)
			}

		})
	}
}

func TestGetTokenForTeam(t *testing.T) {
	ts := store.NewTeamStore()
	team := store.Team{
		Name:  "Test team",
		Email: "tkp@tkp.dk",
	}
	if err := ts.CreateTeam(team); err != nil {
		t.Fatalf("expected no error when creating team")
	}

	token := "token-to-test"
	err := ts.CreateTokenForTeam(token, team)
	if err != nil {
		t.Fatalf("expected no error when creating token")
	}

	tt := []struct {
		name  string
		team  store.Team
		token string
		err   string
	}{
		{name: "Normal", token: token, team: team},
		{name: "No team", token: "invalid-token", err: "Unknown token"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			team, err := ts.GetTeamByToken(tc.token)
			if err != nil {
				if tc.err != "" {
					if tc.err != err.Error() {
						t.Fatalf("unexpected error (expected: \"%s\") when getting team by token: %s", tc.err, err)
					}

					return
				}

				t.Fatalf("received error when getting team by token, but expected none: %s", err)
			}

			if tc.err != "" {
				t.Fatalf("expected error but received none: %s", tc.err)
			}

			if team.Email != tc.team.Email {
				t.Fatalf("received unexpected team: %+v", team)
			}
		})
	}
}

func TestDeleteToken(t *testing.T) {
	tt := []struct {
		name        string
		inputToken  string
		deleteToken string
		err         string
	}{
		{name: "Normal", inputToken: "some_token", deleteToken: "some_token"},
		{name: "Empty token", inputToken: "some_token", deleteToken: ""},
		{name: "Unknown token", inputToken: "some_token", deleteToken: "some_other_token", err: "Unknown token"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := store.NewTeamStore()
			team := store.Team{
				Name:  "Test team",
				Email: "tkp@tkp.dk",
			}
			if err := ts.CreateTeam(team); err != nil {
				t.Fatalf("expected no error when creating team")
			}

			err := ts.CreateTokenForTeam(tc.inputToken, team)
			if err != nil {
				t.Fatalf("expected no error when creating token")
			}

			err = ts.DeleteToken(tc.deleteToken)
			if err != nil {
				if tc.err != "" {
					if tc.err != err.Error() {
						t.Fatalf("unexpected error (expected: \"%s\") when getting team by token: %s", tc.err, err)
					}

					return
				}

				t.Fatalf("received error when getting team by token, but expected none: %s", err)
			}
		})
	}
}