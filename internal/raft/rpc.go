package raft

type RequestVoteRequest struct {
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteResponse struct {
	Term        int
	VoteGranted bool
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
