// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package routing

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	neural "huginn/internal/neural"
)

var (
	ErrModelNotTrained = errors.New("routing: model not trained")
	ErrEmptyTask       = errors.New("routing: empty task")
)

type Task struct {
	ID          string
	Title       string
	Description string
	Type        string
}

type Decision struct {
	AgentID    string
	Confidence float64
	Scores     map[string]float64
}

type MetricsSnapshot struct {
	TrainingLoss      float64
	TrainingAccuracy  float64
	InferenceCount    int64
	AverageConfidence float64
	totalConfidence   float64
}

type NeuralRouter struct {
	featureExtractor *FeatureExtractor
	model            *neural.Model
	trainer          *neural.Trainer
	agentIDs         []string
	rng              *rand.Rand
	mu               sync.RWMutex
	metrics          MetricsSnapshot
	trained          bool
}

func NewNeuralRouter(agentIDs []string, rng *rand.Rand) *NeuralRouter {
	if rng == nil {
		rng = rand.New(rand.NewSource(42))
	}
	return &NeuralRouter{
		featureExtractor: NewFeatureExtractor(),
		agentIDs:         agentIDs,
		rng:              rng,
	}
}

func (nr *NeuralRouter) BuildModel(hiddenDim int) error {
	m := neural.NewModel(7)
	if err := m.AddDense(hiddenDim, nr.rng); err != nil {
		return err
	}
	m.AddActivation(neural.NewReLU())
	if err := m.AddDense(len(nr.agentIDs), nr.rng); err != nil {
		return err
	}
	m.AddActivation(neural.NewSoftmaxActivation())
	nr.model = m
	return nil
}

func (nr *NeuralRouter) Train(ctx context.Context, dataset *SyntheticDataset, config neural.TrainConfig) ([]neural.Metrics, error) {
	nr.mu.Lock()
	defer nr.mu.Unlock()
	if nr.model == nil {
		if err := nr.BuildModel(16); err != nil {
			return nil, err
		}
	}
	loss := neural.NewCrossEntropyLoss()
	optim, err := neural.NewSGD(config.LearningRate)
	if err != nil {
		return nil, err
	}
	nr.trainer = neural.NewTrainer(nr.model, loss, optim, config)
	inputs, targets, err := dataset.ToTensors()
	if err != nil {
		return nil, err
	}
	history, err := nr.trainer.Train(inputs, targets)
	if err != nil {
		return nil, err
	}
	if len(history) > 0 {
		nr.metrics.TrainingLoss = history[len(history)-1].Loss
		nr.metrics.TrainingAccuracy = history[len(history)-1].Accuracy
	}
	nr.trained = true
	return history, nil
}

func (nr *NeuralRouter) Route(ctx context.Context, task Task) (Decision, error) {
	nr.mu.Lock()
	defer nr.mu.Unlock()
	if !nr.trained || nr.model == nil {
		return Decision{}, ErrModelNotTrained
	}
	if task.Title == "" && task.Description == "" {
		return Decision{}, ErrEmptyTask
	}
	features := nr.featureExtractor.Extract(task.Title, task.Description)
	input := features.ToTensor()
	output, err := nr.model.Forward(input)
	if err != nil {
		return Decision{}, err
	}
	scores := make(map[string]float64)
	bestIdx := 0
	bestScore := output.Data[0]
	for i, id := range nr.agentIDs {
		scores[id] = output.Data[i]
		if output.Data[i] > bestScore {
			bestScore = output.Data[i]
			bestIdx = i
		}
	}
	nr.metrics.InferenceCount++
	nr.metrics.totalConfidence += bestScore
	nr.metrics.AverageConfidence = nr.metrics.totalConfidence / float64(nr.metrics.InferenceCount)
	return Decision{
		AgentID:    nr.agentIDs[bestIdx],
		Confidence: bestScore,
		Scores:     scores,
	}, nil
}

func (nr *NeuralRouter) GetMetrics() MetricsSnapshot {
	nr.mu.RLock()
	defer nr.mu.RUnlock()
	return nr.metrics
}

func (nr *NeuralRouter) IsTrained() bool {
	nr.mu.RLock()
	defer nr.mu.RUnlock()
	return nr.trained
}

func (nr *NeuralRouter) SaveWeights(path string) error {
	nr.mu.RLock()
	defer nr.mu.RUnlock()
	if nr.model == nil {
		return ErrModelNotTrained
	}
	return neural.SaveModel(path, nr.model)
}

func (nr *NeuralRouter) LoadWeights(path string) error {
	nr.mu.Lock()
	defer nr.mu.Unlock()
	md, err := neural.LoadModel(path)
	if err != nil {
		return err
	}
	if nr.model == nil {
		if err := nr.BuildModel(16); err != nil {
			return err
		}
	}
	layers := make([]neural.Layer, len(nr.model.Layers()))
	for i, l := range nr.model.Layers() {
		layers[i] = l
	}
	if err := neural.RestoreModel(md, layers); err != nil {
		return err
	}
	nr.trained = true
	return nil
}

func (m *MetricsSnapshot) String() string {
	return fmt.Sprintf("loss=%.6f accuracy=%.4f inferences=%d avg_confidence=%.4f",
		m.TrainingLoss, m.TrainingAccuracy, m.InferenceCount, m.AverageConfidence)
}
