// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package neural

import (
	"errors"
	"math"
)

type Activation interface {
	Name() string
	Forward(input *Tensor) (*Tensor, error)
	Backward(gradOutput *Tensor) (*Tensor, error)
}

type ReLU struct {
	input *Tensor
}

func NewReLU() *ReLU {
	return &ReLU{}
}

func (r *ReLU) Name() string { return "relu" }

func (r *ReLU) Forward(input *Tensor) (*Tensor, error) {
	if err := input.validate(); err != nil {
		return nil, err
	}
	r.input = input.Clone()
	return input.Map(func(v float64) float64 {
		if v > 0 {
			return v
		}
		return 0
	}), nil
}

func (r *ReLU) Backward(gradOutput *Tensor) (*Tensor, error) {
	if err := gradOutput.validate(); err != nil {
		return nil, err
	}
	if r.input == nil {
		return nil, ErrNoForwardCall
	}
	return gradOutput.Hadamard(r.input.Map(func(v float64) float64 {
		if v > 0 {
			return 1
		}
		return 0
	}))
}

type Sigmoid struct {
	output *Tensor
}

func NewSigmoid() *Sigmoid {
	return &Sigmoid{}
}

func (s *Sigmoid) Name() string { return "sigmoid" }

func (s *Sigmoid) Forward(input *Tensor) (*Tensor, error) {
	if err := input.validate(); err != nil {
		return nil, err
	}
	out := input.Map(func(v float64) float64 {
		if v >= 0 {
			return 1.0 / (1.0 + math.Exp(-v))
		}
		ev := math.Exp(v)
		return ev / (1.0 + ev)
	})
	s.output = out.Clone()
	return out, nil
}

func (s *Sigmoid) Backward(gradOutput *Tensor) (*Tensor, error) {
	if err := gradOutput.validate(); err != nil {
		return nil, err
	}
	if s.output == nil {
		return nil, ErrNoForwardCall
	}
	grad := s.output.Map(func(v float64) float64 {
		return v * (1 - v)
	})
	return gradOutput.Hadamard(grad)
}

type SoftmaxActivation struct {
	output *Tensor
}

func NewSoftmaxActivation() *SoftmaxActivation {
	return &SoftmaxActivation{}
}

func (s *SoftmaxActivation) Name() string { return "softmax" }

func (s *SoftmaxActivation) Forward(input *Tensor) (*Tensor, error) {
	out, err := Softmax(input)
	if err != nil {
		return nil, err
	}
	s.output = out.Clone()
	return out, nil
}

func (s *SoftmaxActivation) Backward(gradOutput *Tensor) (*Tensor, error) {
	if err := gradOutput.validate(); err != nil {
		return nil, err
	}
	if s.output == nil {
		return nil, ErrNoForwardCall
	}
	out, _ := New(s.output.Rows, s.output.Cols)
	for i := 0; i < s.output.Rows; i++ {
		row := make([]float64, s.output.Cols)
		copy(row, s.output.Data[i*s.output.Cols:(i+1)*s.output.Cols])
		for j := 0; j < s.output.Cols; j++ {
			sum := 0.0
			for k := 0; k < s.output.Cols; k++ {
				if j == k {
					sum += gradOutput.At(i, k) * row[j] * (1 - row[j])
				} else {
					sum += gradOutput.At(i, k) * (-row[j] * row[k])
				}
			}
			out.Data[i*s.output.Cols+j] = sum
		}
	}
	return out, nil
}

var ErrNoForwardCall = errors.New("neural: Forward() must be called before Backward()")
