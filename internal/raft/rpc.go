package raft

type RPCType string

const (
	RequestVoteRPC   RPCType = "RequestVote"
	AppendEntriesRPC RPCType = "AppendEntries"
)

type RequestVoteRequest struct {
	Type         RPCType `json:"type"`
	Term         int     `json:"term"`
	CandidateID  int     `json:"candidate_id"`
	LastLogIndex int     `json:"last_log_index"`
	LastLogTerm  int     `json:"last_log_term"`
}

type RequestVoteResponse struct {
	Type        RPCType `json:"type"`
	Term        int     `json:"term"`
	VoteGranted bool    `json:"vote_granted"`
}

type AppendEntriesRequest struct {
	Type              RPCType  `json:"type"`
	Term              int      `json:"term"`
	LeaderID          int      `json:"leader_id"`
	PrevLogIndex      int      `json:"prev_log_index"`
	PrevLogTerm       int      `json:"prev_log_term"`
	Entries           []string `json:"entries"`
	LeaderCommitIndex int      `json:"leader_commit_index"`
}

type AppendEntriesResponse struct {
	Type    RPCType `json:"type"`
	Term    int     `json:"term"`
	Success bool    `json:"success"`
}
