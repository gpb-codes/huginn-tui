// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package neural

import (
	"errors"
	"math"
	"math/rand"
)

var (
	ErrInvalidLearningRate = errors.New("neural: learning rate must be positive")
	ErrInvalidLayerDims    = errors.New("neural: layer dimensions must be positive")
)

type Dense struct {
	Weights     *Tensor
	Bias        *Tensor
	inputCache  *Tensor
	outputCache *Tensor
	gradWeights *Tensor
	gradBias    *Tensor
	inputDim    int
	outputDim   int
}

func NewDense(inputDim, outputDim int, rng *rand.Rand) (*Dense, error) {
	if inputDim <= 0 || outputDim <= 0 {
		return nil, ErrInvalidLayerDims
	}
	scale := math.Sqrt(2.0 / float64(inputDim))
	w, err := Random(outputDim, inputDim, rng)
	if err != nil {
		return nil, err
	}
	for i := range w.Data {
		w.Data[i] *= scale
	}
	b, _ := Zeros(1, outputDim)
	return &Dense{
		Weights:   w,
		Bias:      b,
		inputDim:  inputDim,
		outputDim: outputDim,
	}, nil
}

func (d *Dense) Forward(input *Tensor) (*Tensor, error) {
	if err := input.validate(); err != nil {
		return nil, err
	}
	if input.Cols != d.inputDim {
		return nil, ErrDimensionMismatch
	}
	d.inputCache = input.Clone()
	wt, err := d.Weights.Transpose()
	if err != nil {
		return nil, err
	}
	out, err := input.Mul(wt)
	if err != nil {
		return nil, err
	}
	// Broadcast bias (1, outputDim) across batch
	for i := 0; i < out.Rows; i++ {
		for j := 0; j < out.Cols; j++ {
			out.Data[i*out.Cols+j] += d.Bias.Data[j]
		}
	}
	d.outputCache = out.Clone()
	return out, nil
}

func (d *Dense) Backward(gradOutput *Tensor) (*Tensor, error) {
	if err := gradOutput.validate(); err != nil {
		return nil, err
	}
	if d.inputCache == nil {
		return nil, ErrNoForwardCall
	}
	// gradWeights = gradOutput^T * input
	gradOutputT, err := gradOutput.Transpose()
	if err != nil {
		return nil, err
	}
	d.gradWeights, err = gradOutputT.Mul(d.inputCache)
	if err != nil {
		return nil, err
	}
	// gradBias = sum over batch (rows) per output dim -> (1, outputDim)
	gradBias, _ := New(1, gradOutput.Cols)
	for j := 0; j < gradOutput.Cols; j++ {
		sum := 0.0
		for i := 0; i < gradOutput.Rows; i++ {
			sum += gradOutput.At(i, j)
		}
		gradBias.Data[j] = sum
	}
	d.gradBias = gradBias
	gradInput, err := gradOutput.Mul(d.Weights)
	if err != nil {
		return nil, err
	}
	return gradInput, nil
}

func (d *Dense) GetWeights() *Tensor     { return d.Weights }
func (d *Dense) GetBias() *Tensor        { return d.Bias }
func (d *Dense) GetGradWeights() *Tensor { return d.gradWeights }
func (d *Dense) GetGradBias() *Tensor    { return d.gradBias }
func (d *Dense) InputDim() int           { return d.inputDim }
func (d *Dense) OutputDim() int          { return d.outputDim }
