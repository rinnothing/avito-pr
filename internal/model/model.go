package model

import "time"

type UserId string

type TeamName string

type User struct {
	Id       UserId
	Username string
	TeamName TeamName
	IsActive bool
}

type Team struct {
	Name    TeamName
	Members []UserId
}

type PRStatus int

const (
	PROpen PRStatus = iota + 1
	PRMerged
)

type PRId string

type PullRequest struct {
	Id        PRId
	Name      string
	AuthorId  UserId
	Status    PRStatus
	Reviewers []UserId
	CreatedAt time.Time
	MergedAt  *time.Time
}
