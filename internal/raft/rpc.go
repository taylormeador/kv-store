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
	Term            int
	LeaderID        int
	PrevLogIndex    int
	PrevLogTerm     int
	Entries         []string
	LeaderCommitIdx int
}

type AppendEntriesResponse struct {
	Term    int
	Success bool
}
