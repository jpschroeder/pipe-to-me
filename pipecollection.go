package main

import (
	"fmt"
	"strings"
)

// PipeCollection is a map of pipes partitioned by a key
type PipeCollection struct {
	// pipe key -> Pipe
	pipes map[string]*Pipe
	stats *PipeStats
}

// WriteCompleted is a called by the individual pipes to collect statistics
func (pc *PipeCollection) WriteCompleted(bytes int) {
	pc.stats.BytesSent += bytes
}

// WriteCompleteHandler is a callback interface used to collect statistics
type WriteCompleteHandler interface {
	WriteCompleted(bytes int)
}

// FindOrCreatePipe finds a pipe or creates one if it doesn't exist
func (pc *PipeCollection) FindOrCreatePipe(key string) *Pipe {
	pipe, exists := pc.pipes[key]
	if !exists {
		pipe = MakePipe(pc)
		pc.pipes[key] = pipe
		pc.stats.PipeCount++
	}
	return pipe
}

// DeletePipeIfEmpty deletes the pipe if it has no attached receivers
func (pc *PipeCollection) DeletePipeIfEmpty(key string, pipe *Pipe) {
	if pipe.ReceiverCount() < 1 && pipe.SenderCount() < 1 {
		delete(pc.pipes, key)
	}
}

// AddReceiver adds a new receiver to a pipe - creates the pipe if it doesn't exist
func (pc *PipeCollection) AddReceiver(key string, receiver RecieveWriter) *Pipe {
	pipe := pc.FindOrCreatePipe(key)
	pipe.AddReceiver(receiver)
	pc.stats.ReceiverCount++
	return pipe
}

// RemoveReceiver removes a receiver from a pipe - removes the pipe if its empty
func (pc *PipeCollection) RemoveReceiver(key string, receiver RecieveWriter) {
	pipe, exists := pc.pipes[key]
	if !exists {
		return
	}
	pipe.RemoveReceiver(receiver)
	pc.DeletePipeIfEmpty(key, pipe)
}

// AddSender adds a new sender to a pipe - creates the pipe if it doesn't exist
func (pc *PipeCollection) AddSender(key string) *Pipe {
	pipe := pc.FindOrCreatePipe(key)
	pipe.AddSender()
	pc.stats.SenderCount++
	return pipe
}

// RemoveSender removes a sender from the pipe - remove the pipe if its empty
func (pc *PipeCollection) RemoveSender(key string, pipe *Pipe) {
	pipe.RemoveSender()
	pc.DeletePipeIfEmpty(key, pipe)
}

// PipeStats holds statistics about a pipe or collection of pipes
type PipeStats struct {
	PipeCount     int
	ReceiverCount int
	SenderCount   int
	BytesSent     int
}

// MegaBytesSent returns the number of megabytes in the statistics
func (ps PipeStats) MegaBytesSent() int {
	return ps.BytesSent / 1000000
}

// ActiveStats returns the statistics for only connected pipes in the collection
func (pc PipeCollection) ActiveStats() PipeStats {
	stats := PipeStats{}
	for _, pipe := range pc.pipes {
		stats.PipeCount++
		stats.ReceiverCount += pipe.ReceiverCount()
		stats.SenderCount += pipe.SenderCount()
		stats.BytesSent += pipe.BytesSent()
	}
	return stats
}

// GlobalStats returns the statistics for all pipes ever to exist in the collection
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

// MakePipeCollection creates an empty collection of pipes
func MakePipeCollection() PipeCollection {
	stats := PipeStats{}
	return PipeCollection{
		pipes: make(map[string]*Pipe),
		stats: &stats,
	}
}
