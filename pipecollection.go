package main

import (
	"fmt"
	"strings"
)

// A map of pipes partitioned by a key

type PipeCollection struct {
	// pipe key -> Pipe
	pipes map[string]*Pipe
	stats *PipeStats
}

func (pc *PipeCollection) WriteCompleted(bytes int) {
	pc.stats.BytesSent += bytes
}

type WriteCompleteHandler interface {
	WriteCompleted(bytes int)
}

// find a pipe or create one if it doesn't exist
func (pc *PipeCollection) FindOrCreatePipe(key string) *Pipe {
	pipe, exists := pc.pipes[key]
	if !exists {
		pipe = MakePipe(pc)
		pc.pipes[key] = pipe
		pc.stats.PipeCount++
	}
	return pipe
}

// delete the pipe if it has no attached receivers
func (pc *PipeCollection) DeletePipeIfEmpty(key string, pipe *Pipe) {
	if pipe.ReceiverCount() < 1 && pipe.SenderCount() < 1 {
		delete(pc.pipes, key)
	}
}

// add a new receiver to a pipe - create the pipe if it doesn't exist
func (pc *PipeCollection) AddReceiver(key string, receiver RecieveWriter) *Pipe {
	pipe := pc.FindOrCreatePipe(key)
	pipe.AddReceiver(receiver)
	pc.stats.ReceiverCount++
	return pipe
}

// remove a receiver from a pipe - remove the pipe if its empty
func (pc *PipeCollection) RemoveReceiver(key string, receiver RecieveWriter) {
	pipe, exists := pc.pipes[key]
	if !exists {
		return
	}
	pipe.RemoveReceiver(receiver)
	pc.DeletePipeIfEmpty(key, pipe)
}

// add a new sender to a pipe - create the pipe if it doesn't exist
func (pc *PipeCollection) AddSender(key string) *Pipe {
	pipe := pc.FindOrCreatePipe(key)
	pipe.AddSender()
	pc.stats.SenderCount++
	return pipe
}

// remove a sender from the pipe - remove the pipe if its empty
func (pc *PipeCollection) RemoveSender(key string, pipe *Pipe) {
	pipe.RemoveSender()
	pc.DeletePipeIfEmpty(key, pipe)
}

type PipeStats struct {
	PipeCount     int
	ReceiverCount int
	SenderCount   int
	BytesSent     int
}

func (ps PipeStats) MegaBytesSent() int {
	return ps.BytesSent / 1000000
}

func MakeStats() PipeStats {
	return PipeStats{
		PipeCount:     0,
		ReceiverCount: 0,
		SenderCount:   0,
		BytesSent:     0,
	}
}

func (pc PipeCollection) ActiveStats() PipeStats {
	stats := MakeStats()
	for _, pipe := range pc.pipes {
		stats.PipeCount++
		stats.ReceiverCount += pipe.ReceiverCount()
		stats.SenderCount += pipe.SenderCount()
		stats.BytesSent += pipe.BytesSent()
	}
	return stats
}

func (pc PipeCollection) GlobalStats() PipeStats {
	return *pc.stats
}

func (pc PipeCollection) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d keys\n", len(pc.pipes)))
	for key, pipe := range pc.pipes {
		sb.WriteString(fmt.Sprintf("%s: %s", key, pipe.String()))
	}
	return sb.String()
}

func MakePipeCollection() PipeCollection {
	stats := MakeStats()
	return PipeCollection{
		pipes: make(map[string]*Pipe),
		stats: &stats,
	}
}
