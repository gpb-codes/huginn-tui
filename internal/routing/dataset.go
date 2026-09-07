// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package routing

import (
	neural "huginn/internal/neural"
)

type Sample struct {
	Features TaskFeatures
	AgentID  string
}

type SyntheticDataset struct {
	AgentIDs []string
	Samples  []Sample
}

func NewSyntheticDataset() *SyntheticDataset {
	return &SyntheticDataset{
		AgentIDs: []string{"opencode", "researcher", "toolmaster"},
		Samples: []Sample{
			{Features: TaskFeatures{0.8, 0.9, 0.7, 0.2, 0.3, 0.5, 0.5}, AgentID: "opencode"},
			{Features: TaskFeatures{0.7, 0.8, 0.6, 0.3, 0.3, 0.6, 0.5}, AgentID: "opencode"},
			{Features: TaskFeatures{0.9, 0.9, 0.8, 0.2, 0.4, 0.7, 0.5}, AgentID: "opencode"},
			{Features: TaskFeatures{0.6, 0.7, 0.5, 0.3, 0.2, 0.4, 0.5}, AgentID: "opencode"},
			{Features: TaskFeatures{0.8, 0.85, 0.7, 0.25, 0.35, 0.55, 0.5}, AgentID: "opencode"},
			{Features: TaskFeatures{0.3, 0.2, 0.1, 0.8, 0.9, 0.6, 0.5}, AgentID: "researcher"},
			{Features: TaskFeatures{0.4, 0.1, 0.1, 0.9, 0.8, 0.7, 0.5}, AgentID: "researcher"},
			{Features: TaskFeatures{0.5, 0.3, 0.2, 0.7, 0.9, 0.5, 0.5}, AgentID: "researcher"},
			{Features: TaskFeatures{0.3, 0.2, 0.1, 0.85, 0.85, 0.8, 0.5}, AgentID: "researcher"},
			{Features: TaskFeatures{0.4, 0.25, 0.15, 0.75, 0.88, 0.65, 0.5}, AgentID: "researcher"},
			{Features: TaskFeatures{0.5, 0.4, 0.9, 0.3, 0.2, 0.4, 0.5}, AgentID: "toolmaster"},
			{Features: TaskFeatures{0.4, 0.3, 0.85, 0.2, 0.3, 0.3, 0.5}, AgentID: "toolmaster"},
			{Features: TaskFeatures{0.6, 0.5, 0.95, 0.2, 0.2, 0.5, 0.5}, AgentID: "toolmaster"},
			{Features: TaskFeatures{0.3, 0.35, 0.8, 0.15, 0.2, 0.35, 0.5}, AgentID: "toolmaster"},
			{Features: TaskFeatures{0.45, 0.42, 0.88, 0.22, 0.25, 0.42, 0.5}, AgentID: "toolmaster"},
		},
	}
}

func (ds *SyntheticDataset) ToTensors() (inputs, targets *neural.Tensor, err error) {
	n := len(ds.Samples)
	inputs, err = neural.New(n, 7)
	if err != nil {
		return nil, nil, err
	}
	targets, err = neural.New(n, len(ds.AgentIDs))
	if err != nil {
		return nil, nil, err
	}
	agentIdx := make(map[string]int)
	for i, id := range ds.AgentIDs {
		agentIdx[id] = i
	}
	for i, s := range ds.Samples {
		f := s.Features
		inputs.Data[i*7+0] = f.Complexity
		inputs.Data[i*7+1] = f.CodeRelated
		inputs.Data[i*7+2] = f.TerminalRequired
		inputs.Data[i*7+3] = f.ResearchRequired
		inputs.Data[i*7+4] = f.ReasoningRequired
		inputs.Data[i*7+5] = f.ContextSize
		inputs.Data[i*7+6] = f.HistoricalSuccess
		if idx, ok := agentIdx[s.AgentID]; ok {
			targets.Data[i*len(ds.AgentIDs)+idx] = 1.0
		}
	}
	return inputs, targets, nil
}

func (ds *SyntheticDataset) AgentCount() int {
	return len(ds.AgentIDs)
}

// NewSyntheticDatasetFor reetiqueta las muestras canónicas a IDs personalizados.
// Asigna código->ids[0], research->ids[1] y terminal->ids[2] para entrenar con roles reales.
func NewSyntheticDatasetFor(agentIDs []string) *SyntheticDataset {
	if len(agentIDs) == 0 {
		return NewSyntheticDataset()
	}
	base := NewSyntheticDataset()
	roleOf := func(canonical string) string {
		switch canonical {
		case "opencode":
			return agentIDs[0]
		case "researcher":
			if len(agentIDs) > 1 {
				return agentIDs[1]
			}
			return agentIDs[0]
		default: // toolmaster and any future canonical label
			if len(agentIDs) > 2 {
				return agentIDs[2]
			}
			return agentIDs[len(agentIDs)-1]
		}
	}
	out := &SyntheticDataset{AgentIDs: agentIDs}
	for _, s := range base.Samples {
		s.AgentID = roleOf(s.AgentID)
		out.Samples = append(out.Samples, s)
	}
	return out
}
