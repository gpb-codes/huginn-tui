// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package neural

import (
	"errors"
	"math"
)

var ErrInvalidProbabilities = errors.New("neural: probabilities must be in (0,1)")

type Loss interface {
	Name() string
	Forward(predicted, target *Tensor) (float64, error)
	Backward(predicted, target *Tensor) (*Tensor, error)
}

type CrossEntropyLoss struct{}

func NewCrossEntropyLoss() *CrossEntropyLoss {
	return &CrossEntropyLoss{}
}

func (ce *CrossEntropyLoss) Name() string { return "cross_entropy" }

func (ce *CrossEntropyLoss) Forward(predicted, target *Tensor) (float64, error) {
	if err := predicted.validate(); err != nil {
		return 0, err
	}
	if err := target.validate(); err != nil {
		return 0, err
	}
	if predicted.Rows != target.Rows || predicted.Cols != target.Cols {
		return 0, ErrDimensionMismatch
	}
	eps := 1e-15
	loss := 0.0
	n := float64(predicted.Rows)
	for i := range predicted.Data {
		p := predicted.Data[i]
		if p < eps {
			p = eps
		} else if p > 1-eps {
			p = 1 - eps
		}
		t := target.Data[i]
		if t > 0 {
			loss -= t * math.Log(p)
		}
	}
	return loss / n, nil
}

func (ce *CrossEntropyLoss) Backward(predicted, target *Tensor) (*Tensor, error) {
	if err := predicted.validate(); err != nil {
		return nil, err
	}
	if err := target.validate(); err != nil {
		return nil, err
	}
	if predicted.Rows != target.Rows || predicted.Cols != target.Cols {
		return nil, ErrDimensionMismatch
	}
	eps := 1e-15
	out := predicted.Clone()
	n := float64(predicted.Rows)
	for i := range out.Data {
		p := out.Data[i]
		if p < eps {
			p = eps
		} else if p > 1-eps {
			p = 1 - eps
		}
		if target.Data[i] > 0 {
			out.Data[i] = -target.Data[i] / p / n
		} else {
			out.Data[i] = 0
		}
	}
	return out, nil
}
