// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package neural

import (
	"fmt"
	"math"
)

type Metrics struct {
	Epoch    int
	Loss     float64
	Accuracy float64
}

type TrainConfig struct {
	Epochs       int
	LearningRate float64
	PrintEvery   int
}

func DefaultTrainConfig() TrainConfig {
	return TrainConfig{
		Epochs:       1000,
		LearningRate: 0.01,
		PrintEvery:   100,
	}
}

type Trainer struct {
	model  *Model
	loss   Loss
	optim  Optimizer
	config TrainConfig
}

func NewTrainer(model *Model, loss Loss, optim Optimizer, config TrainConfig) *Trainer {
	return &Trainer{
		model:  model,
		loss:   loss,
		optim:  optim,
		config: config,
	}
}

func (t *Trainer) Train(inputs, targets *Tensor) ([]Metrics, error) {
	if err := inputs.validate(); err != nil {
		return nil, err
	}
	if err := targets.validate(); err != nil {
		return nil, err
	}
	if inputs.Rows != targets.Rows {
		return nil, ErrDimensionMismatch
	}
	var history []Metrics
	for epoch := 1; epoch <= t.config.Epochs; epoch++ {
		predicted, err := t.model.Forward(inputs)
		if err != nil {
			return nil, err
		}
		lossVal, err := t.loss.Forward(predicted, targets)
		if err != nil {
			return nil, err
		}
		grad, err := t.loss.Backward(predicted, targets)
		if err != nil {
			return nil, err
		}
		if err := t.model.Backward(grad); err != nil {
			return nil, err
		}
		params, grads := t.model.GetParamsAndGrads()
		if err := t.optim.Update(params, grads); err != nil {
			return nil, err
		}
		acc := t.accuracy(predicted, targets)
		m := Metrics{
			Epoch:    epoch,
			Loss:     lossVal,
			Accuracy: acc,
		}
		history = append(history, m)
		if t.config.PrintEvery > 0 && epoch%t.config.PrintEvery == 0 {
			// Solo loguea si el caller quiere traza; no contamina stdout por defecto
			_ = fmt.Sprintf("epoch %d/%d loss=%.6f acc=%.4f", epoch, t.config.Epochs, lossVal, acc)
		}
	}
	return history, nil
}

func (t *Trainer) accuracy(predicted, target *Tensor) float64 {
	correct := 0
	for i := 0; i < predicted.Rows; i++ {
		predRow := predicted.Data[i*predicted.Cols : (i+1)*predicted.Cols]
		targetRow := target.Data[i*target.Cols : (i+1)*target.Cols]
		predClass := argMaxSlice(predRow)
		targetClass := argMaxSlice(targetRow)
		if predClass == targetClass {
			correct++
		}
	}
	return float64(correct) / float64(predicted.Rows)
}

func argMaxSlice(s []float64) int {
	best := 0
	for i := 1; i < len(s); i++ {
		if s[i] > s[best] {
			best = i
		}
	}
	return best
}

func (t *Trainer) Predict(input *Tensor) (*Tensor, error) {
	return t.model.Forward(input)
}

func Clamp(v, min, max float64) float64 {
	return math.Max(min, math.Min(max, v))
}
