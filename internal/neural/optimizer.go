// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package neural

import "errors"

var ErrNilGradient = errors.New("neural: nil gradient")

type Optimizer interface {
	Name() string
	Update(params, grads []*Tensor) error
	SetLearningRate(lr float64)
	GetLearningRate() float64
}

type SGD struct {
	lr float64
}

func NewSGD(lr float64) (*SGD, error) {
	if lr <= 0 {
		return nil, ErrInvalidLearningRate
	}
	return &SGD{lr: lr}, nil
}

func (s *SGD) Name() string               { return "sgd" }
func (s *SGD) SetLearningRate(lr float64) { s.lr = lr }
func (s *SGD) GetLearningRate() float64   { return s.lr }

func (s *SGD) Update(params, grads []*Tensor) error {
	if len(params) != len(grads) {
		return ErrDimensionMismatch
	}
	for i := range params {
		if params[i] == nil || grads[i] == nil {
			return ErrNilGradient
		}
		if params[i].Rows != grads[i].Rows || params[i].Cols != grads[i].Cols {
			return ErrDimensionMismatch
		}
		for j := range params[i].Data {
			params[i].Data[j] -= s.lr * grads[i].Data[j]
		}
	}
	return nil
}
